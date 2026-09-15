package catalog

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/richavery/bvr-cli/internal/imagelookup"
	"github.com/stretchr/testify/require"
)

func TestOpen_CreatesSchema(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	repo, err := Open(dbPath)
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()
	p := Product{
		ID:       "test-1",
		Brand:    "TestBrand",
		Category: "handbag",
		Title:    "Test Handbag",
		Aliases:  []string{"Test Bag"},
	}
	require.NoError(t, repo.InsertProduct(ctx, p))
}

func TestSearch_FindsProductsByFTS(t *testing.T) {
	t.Parallel()
	repo := openTestRepo(t)
	defer repo.Close()

	ctx := context.Background()
	require.NoError(t, repo.InsertProduct(ctx, Product{
		ID:           "lv1",
		Brand:        "Louis Vuitton",
		Model:        ns("Neverfull MM"),
		Category:     "handbag",
		Subcategory:  ns("tote"),
		Title:        "Louis Vuitton Neverfull MM Monogram Canvas",
		Material:     ns("coated canvas, vachetta trim"),
		Color:        ns("brown"),
		Pattern:      ns("LV monogram"),
		Hardware:     ns("gold-tone"),
		ReferenceURL: ns("https://example.invalid/lv1"),
		Aliases:      []string{"Neverfull"},
	}))
	require.NoError(t, repo.InsertProduct(ctx, Product{
		ID:           "gucci1",
		Brand:        "Gucci",
		Category:     "handbag",
		Title:        "Gucci GG Canvas Tote",
		Material:     ns("canvas"),
		Pattern:      ns("GG"),
		Hardware:     ns("gold-tone hardware"),
		ReferenceURL: ns("https://example.invalid/gucci1"),
		Aliases:      []string{"GG Tote"},
	}))

	query := "category: handbag; brand: Louis Vuitton; pattern: LV monogram; material: coated canvas"
	candidates, err := repo.Search(ctx, query, 10)
	require.NoError(t, err)
	require.Len(t, candidates, 2)

	// The LV product should rank higher than Gucci.
	require.Equal(t, "lv1", candidates[0].ProductID)
	require.Greater(t, candidates[0].Score, candidates[1].Score)
	require.Contains(t, candidates[0].MatchedSignals, "brand: Louis Vuitton")
	require.NotEmpty(t, candidates[0].ReferenceURL)
}

func TestSearch_EmptyResults(t *testing.T) {
	t.Parallel()
	repo := openTestRepo(t)
	defer repo.Close()

	ctx := context.Background()
	candidates, err := repo.Search(ctx, "category: nonexistent; brand: Nobody", 10)
	require.NoError(t, err)
	require.Empty(t, candidates)
}

func TestSearch_WithEmbeddingBoost(t *testing.T) {
	t.Parallel()
	repo := openTestRepo(t)
	defer repo.Close()

	vec := []float64{0.1, 0.2, 0.3, 0.4, 0.5}
	embedder := func(text string) ([]float64, error) { return vec, nil }
	repo.WithTextEmbedder("test-model", embedder)

	ctx := context.Background()
	require.NoError(t, repo.InsertProduct(ctx, Product{
		ID:       "p1",
		Brand:    "BrandA",
		Category: "sunglasses",
		Title:    "BrandA Wayfarer Sunglasses",
		Material: ns("plastic"),
	}))
	require.NoError(t, repo.SetProductEmbedding(ctx, "p1", "test-model", vec))

	candidates, err := repo.Search(ctx, "category: sunglasses; brand: BrandA", 10)
	require.NoError(t, err)
	require.Len(t, candidates, 1)
	require.Greater(t, candidates[0].Score, 0.5)
}

func TestCache_StoreAndGet(t *testing.T) {
	t.Parallel()
	repo := openTestRepo(t)
	defer repo.Close()

	ctx := context.Background()
	analysis := &imagelookup.ImageAnalysis{
		ImageQuality:  "good",
		ItemCategory:  imagelookup.Observation{Value: "handbag", Confidence: imagelookup.ConfidenceHigh},
		ProbableBrand: imagelookup.Observation{Value: "Louis Vuitton", Confidence: imagelookup.ConfidenceMedium},
	}

	_, err := repo.GetCachedAnalysis(ctx, "hash123", "qwen2.5vl:3b", "v1")
	require.Error(t, err)

	require.NoError(t, repo.StoreCachedAnalysis(ctx, "hash123", "qwen2.5vl:3b", "v1", analysis))

	cached, err := repo.GetCachedAnalysis(ctx, "hash123", "qwen2.5vl:3b", "v1")
	require.NoError(t, err)
	require.Equal(t, "good", cached.ImageQuality)
	require.Equal(t, "handbag", cached.ItemCategory.Value)
}

func TestCache_DifferentPromptsAreSeparate(t *testing.T) {
	t.Parallel()
	repo := openTestRepo(t)
	defer repo.Close()

	ctx := context.Background()
	a1 := &imagelookup.ImageAnalysis{ItemCategory: imagelookup.Observation{Value: "shoes", Confidence: imagelookup.ConfidenceHigh}}

	require.NoError(t, repo.StoreCachedAnalysis(ctx, "hash1", "qwen2.5vl:3b", "v1", a1))
	a2 := &imagelookup.ImageAnalysis{ItemCategory: imagelookup.Observation{Value: "bags", Confidence: imagelookup.ConfidenceHigh}}
	require.NoError(t, repo.StoreCachedAnalysis(ctx, "hash1", "qwen2.5vl:3b", "v2", a2))

	cached1, _ := repo.GetCachedAnalysis(ctx, "hash1", "qwen2.5vl:3b", "v1")
	require.Equal(t, "shoes", cached1.ItemCategory.Value)

	cached2, _ := repo.GetCachedAnalysis(ctx, "hash1", "qwen2.5vl:3b", "v2")
	require.Equal(t, "bags", cached2.ItemCategory.Value)
}

func TestCosineSimilarity_Identical(t *testing.T) {
	sim := CosineSimilarity([]float64{1, 2, 3}, []float64{1, 2, 3})
	require.InDelta(t, 1.0, sim, 0.001)
}

func TestCosineSimilarity_Orthogonal(t *testing.T) {
	sim := CosineSimilarity([]float64{1, 0}, []float64{0, 1})
	require.InDelta(t, 0.0, sim, 0.001)
}

func TestCosineSimilarity_DifferentLengths(t *testing.T) {
	require.Zero(t, CosineSimilarity([]float64{1, 2}, []float64{1, 2, 3}))
}

// --- helpers ---

func openTestRepo(t *testing.T) *Repository {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	repo, err := Open(dbPath)
	require.NoError(t, err)
	return repo
}

func ns(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
