package catalog

import (
	"context"
	_ "embed"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/richavery/bvr-cli/internal/imagelookup"
)

//go:embed schema.sql
var schemaSQL string

// Product is a single catalog record. It mirrors the products table. Columns
// that may be NULL are represented with sql.Null* types.
type Product struct {
	ID               string
	Brand            string
	Line             sql.NullString
	Model            sql.NullString
	StyleCode        sql.NullString
	Category         string
	Subcategory      sql.NullString
	Title            string
	Description      sql.NullString
	Material         sql.NullString
	Color            sql.NullString
	Pattern          sql.NullString
	Hardware         sql.NullString
	DimensionsText   sql.NullString
	WidthCm          sql.NullFloat64
	HeightCm         sql.NullFloat64
	DepthCm          sql.NullFloat64
	HandleDropCm     sql.NullFloat64
	ClosureType      sql.NullString
	InteriorFeatures sql.NullString
	ExteriorFeatures sql.NullString
	Era              sql.NullString
	Gender           sql.NullString
	ReferenceURL     sql.NullString
	SourceName       sql.NullString
	SourceLicense    sql.NullString
	Notes            sql.NullString
	Aliases          []string
	ImagePaths       []string
}

// Repository wraps a SQLite catalog database and implements
// [imagelookup.CatalogSearcher] and [imagelookup.CacheStore].
type Repository struct {
	db           *sql.DB
	textEmbedder TextEmbedder
	embedModel   string
}

// Open opens (creating if necessary) the catalog database at dbPath, applies
// the schema, and returns a ready-to-use [Repository].
func Open(dbPath string) (*Repository, error) {
	if dbPath == "" {
		return nil, errors.New("catalog database path is required")
	}

	dsn := dbPath + "?_txlock=immediate&_pragma=busy_timeout(30000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open catalog database: %w", err)
	}
	db.SetMaxOpenConns(1)

	r := &Repository{db: db}
	if err := r.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply catalog schema: %w", err)
	}
	return r, nil
}

// WithTextEmbedder sets the function used to generate query embeddings. When
// unset, Search skips embedding-similarity boosting.
func (r *Repository) WithTextEmbedder(model string, embed TextEmbedder) *Repository {
	r.embedModel = model
	r.textEmbedder = embed
	return r
}

// Close releases the underlying database connection.
func (r *Repository) Close() error {
	if r.db == nil {
		return nil
	}
	return r.db.Close()
}

// migrate runs the schema SQL against the database.
func (r *Repository) migrate() error {
	_, err := r.db.ExecContext(context.Background(), schemaSQL)
	return err
}

// Search finds catalog products matching the given query (typically the
// [imagelookup.SearchText] output). It combines FTS5 keyword matching,
// deterministic signal scoring, and optional embedding similarity.
func (r *Repository) Search(ctx context.Context, query string, limit int) ([]imagelookup.Candidate, error) {
	if limit <= 0 {
		limit = 10
	}

	signals := imagelookup.ParseSignals(query)
	ftsQuery := ftsQuery(query)

	// Fetch candidate products via FTS. Pull more than needed so scoring
	// and re-ranking can narrow to the final limit.
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			p.id, p.brand, p.title, p.category, p.pattern, p.material,
			p.hardware, p.dimensions_text, p.reference_url
		FROM products_fts
		JOIN products p ON p.id = products_fts.product_id
		WHERE products_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`, ftsQuery, limit*5)
	if err != nil {
		return nil, fmt.Errorf("FTS search: %w", err)
	}
	defer rows.Close()

	type scoredProduct struct {
		product   Product
		score     float64
		matched   []string
		conflicts []string
	}

	var results []scoredProduct
	for rows.Next() {
		var p Product
		var pattern, material, hardware, dims, refURL sql.NullString

		if err := rows.Scan(
			&p.ID, &p.Brand, &p.Title, &p.Category,
			&pattern, &material, &hardware, &dims, &refURL,
		); err != nil {
			return nil, fmt.Errorf("scan product row: %w", err)
		}
		p.Pattern = pattern
		p.Material = material
		p.Hardware = hardware
		p.DimensionsText = dims
		p.ReferenceURL = refURL

		m := imagelookup.SignalMatches{}
		m.Brand = matchSignal(&m, signals.Brand, p.Brand, "brand")
		m.Category = matchSignal(&m, signals.Category, p.Category, "category")
		m.Pattern = matchAnySignal(&m, signals.Patterns, patStr(pattern), "pattern")
		m.Material = matchAnySignal(&m, signals.Materials, patStr(material), "material")
		m.Hardware = matchAnySignal(&m, signals.Hardware, patStr(hardware), "hardware")
		m.Silhouette = 0.3
		m.Dimensions = 0.1

		score := imagelookup.ComputeScore(m, imagelookup.DefaultSignalWeights)
		results = append(results, scoredProduct{
			product:   p,
			score:     score,
			matched:   m.MatchedSignals,
			conflicts: m.Conflicts,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Embedding-similarity boost (optional).
	if r.textEmbedder != nil && r.embedModel != "" {
		queryVec, err := r.textEmbedder(query)
		if err == nil && len(queryVec) > 0 {
			for i := range results {
				embStr, embErr := r.getEmbedding(ctx, results[i].product.ID)
				if embErr != nil || embStr == "" {
					continue
				}
				prodVec, decErr := DecodeEmbedding(embStr)
				if decErr != nil {
					continue
				}
				sim := CosineSimilarity(queryVec, prodVec)
				results[i].score = results[i].score*0.5 + sim*0.5
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	out := make([]imagelookup.Candidate, 0, min(limit, len(results)))
	for i := range results {
		if i >= limit {
			break
		}
		out = append(out, imagelookup.Candidate{
			ProductID:      results[i].product.ID,
			Brand:          results[i].product.Brand,
			Title:          results[i].product.Title,
			Category:       results[i].product.Category,
			ReferenceURL:   patStr(results[i].product.ReferenceURL),
			Score:          roundScore(results[i].score),
			MatchedSignals: results[i].matched,
			Conflicts:      results[i].conflicts,
		})
	}
	return out, nil
}

// InsertProduct adds a product to the catalog, including its aliases and FTS
// searchable text. If the product already exists (by ID) it is replaced.
func (r *Repository) InsertProduct(ctx context.Context, p Product) error {
	searchableText := buildSearchableText(p)

	_, err := r.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO products (
			id, brand, line, model, style_code, category, subcategory,
			title, description, material, color, pattern, hardware,
			dimensions_text, width_cm, height_cm, depth_cm, handle_drop_cm,
			closure_type, interior_features, exterior_features, era, gender,
			reference_url, source_name, source_license, notes, updated_at
			) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP)
	`, p.ID, p.Brand, p.Line, p.Model, p.StyleCode, p.Category, p.Subcategory,
		p.Title, p.Description, p.Material, p.Color, p.Pattern, p.Hardware,
		p.DimensionsText, p.WidthCm, p.HeightCm, p.DepthCm, p.HandleDropCm,
		p.ClosureType, p.InteriorFeatures, p.ExteriorFeatures, p.Era, p.Gender,
		p.ReferenceURL, p.SourceName, p.SourceLicense, p.Notes)
	if err != nil {
		return fmt.Errorf("insert product: %w", err)
	}

	// Upsert FTS row.
	_, _ = r.db.ExecContext(ctx, `
		INSERT INTO products_fts (
			product_id, brand, line, model, style_code, category,
			material, color, pattern, searchable_text
		) VALUES (?,?,?,?,?,?,?,?,?)
		ON CONFLICT(product_id) DO UPDATE SET
			brand = excluded.brand,
			line = excluded.line,
			model = excluded.model,
			style_code = excluded.style_code,
			category = excluded.category,
			material = excluded.material,
			color = excluded.color,
			pattern = excluded.pattern,
			searchable_text = excluded.searchable_text
	`, p.ID, p.Brand, p.Line, p.Model, p.StyleCode, p.Category,
		p.Material, p.Color, p.Pattern, searchableText)

	for _, alias := range p.Aliases {
		_, _ = r.db.ExecContext(ctx,
			`INSERT OR IGNORE INTO product_aliases (product_id, alias) VALUES (?,?)`,
			p.ID, alias)
	}
	return nil
}

// SetProductEmbedding stores or replaces the text embedding for a product.
func (r *Repository) SetProductEmbedding(ctx context.Context, productID, model string, vec []float64) error {
	embJSON, err := EncodeEmbedding(vec)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO product_embeddings
			(product_id, model, embedding_json, searchable_text, updated_at)
		VALUES (?,?,?,?,CURRENT_TIMESTAMP)
	`, productID, model, embJSON, "")
	return err
}

func (r *Repository) getEmbedding(ctx context.Context, productID string) (string, error) {
	var emb string
	err := r.db.QueryRowContext(ctx,
		`SELECT embedding_json FROM product_embeddings WHERE product_id = ?`,
		productID).Scan(&emb)
	return emb, err
}

// GetCachedAnalysis returns a cached ImageAnalysis if one exists for the
// given image hash, vision model, and prompt version.
func (r *Repository) GetCachedAnalysis(ctx context.Context, imageSHA256, visionModel, promptVersion string) (*imagelookup.ImageAnalysis, error) {
	var jsonStr string
	err := r.db.QueryRowContext(ctx,
		`SELECT analysis_json FROM image_lookup_cache
		 WHERE image_sha256 = ? AND vision_model = ? AND prompt_version = ?`,
		imageSHA256, visionModel, promptVersion).Scan(&jsonStr)
	if err != nil {
		return nil, err
	}
	var a imagelookup.ImageAnalysis
	if err := json.Unmarshal([]byte(jsonStr), &a); err != nil {
		return nil, fmt.Errorf("unmarshal cached analysis: %w", err)
	}
	return &a, nil
}

// StoreCachedAnalysis persists an analysis for future use.
func (r *Repository) StoreCachedAnalysis(ctx context.Context, imageSHA256, visionModel, promptVersion string, analysis *imagelookup.ImageAnalysis) error {
	jsonStr, err := json.Marshal(analysis)
	if err != nil {
		return fmt.Errorf("marshal analysis: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO image_lookup_cache
			(image_sha256, vision_model, prompt_version, analysis_json, created_at)
		VALUES (?,?,?,?,CURRENT_TIMESTAMP)
	`, imageSHA256, visionModel, promptVersion, string(jsonStr))
	return err
}

// --- internal helpers ---

// ftsQuery converts the SearchText label:value format into a space-separated
// bag of terms for an FTS5 MATCH expression.
func ftsQuery(query string) string {
	parts := strings.Split(query, ";")
	var terms []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		idx := strings.Index(part, ":")
		if idx < 0 {
			continue
		}
		value := strings.TrimSpace(part[idx+1:])
		if value != "" {
			terms = append(terms, value)
		}
	}
	if len(terms) == 0 {
		return query
	}
	return strings.Join(terms, " ")
}

// buildSearchableText concatenates all searchable text fields for a product.
func buildSearchableText(p Product) string {
	var sb strings.Builder
	write := func(s string) {
		if s != "" {
			sb.WriteString(" ")
			sb.WriteString(s)
		}
	}
	write(p.Brand)
	write(p.Title)
	write(patStr(p.Line))
	write(patStr(p.Model))
	write(p.Category)
	write(patStr(p.Subcategory))
	write(patStr(p.Material))
	write(patStr(p.Color))
	write(patStr(p.Pattern))
	write(patStr(p.Hardware))
	for _, alias := range p.Aliases {
		write(alias)
	}
	return strings.TrimSpace(sb.String())
}

// matchSignal returns 1.0 if the product field matches the extracted signal,
// recording the match (or conflict) on the SignalMatches struct.
func matchSignal(m *imagelookup.SignalMatches, target, fieldValue, signalName string) float64 {
	if target == "" {
		return 0
	}
	if fieldValue == "" {
		return 0
	}
	if imagelookup.MatchString(fieldValue, target) {
		m.MatchedSignals = append(m.MatchedSignals, signalName+": "+target)
		return 1.0
	}
	m.Conflicts = append(m.Conflicts, signalName+" mismatch: expected "+target+", found "+fieldValue)
	return 0
}

// matchAnySignal checks a product field against multiple extracted signals.
func matchAnySignal(m *imagelookup.SignalMatches, targets []string, fieldValue, signalName string) float64 {
	if fieldValue == "" || len(targets) == 0 {
		return 0
	}
	matched := false
	for _, target := range targets {
		if imagelookup.MatchString(fieldValue, target) {
			m.MatchedSignals = append(m.MatchedSignals, signalName+": "+target)
			matched = true
			break
		}
	}
	if matched {
		return 0.7
	}
	return 0
}

func patStr(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func roundScore(s float64) float64 {
	return math.Round(s*100) / 100
}
