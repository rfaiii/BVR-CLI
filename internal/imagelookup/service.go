package imagelookup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"time"
)

// CatalogSearcher provides ranked catalog candidates for a search query.
type CatalogSearcher interface {
	Search(ctx context.Context, query string, limit int) ([]Candidate, error)
}

// CacheStore persists vision-model analyses so that re-running the feature on
// the same image (with the same model and prompt version) is instant. It is
// optional—pass nil to the service to disable caching.
type CacheStore interface {
	GetCachedAnalysis(ctx context.Context, imageSHA256, visionModel, promptVersion string) (*ImageAnalysis, error)
	StoreCachedAnalysis(ctx context.Context, imageSHA256, visionModel, promptVersion string, analysis *ImageAnalysis) error
}

// Service orchestrates the image-lookup workflow: validate → normalize →
// vision analysis → catalog search → rank → render.
type Service struct {
	Config  Config
	Vision  *OllamaClient
	Catalog CatalogSearcher
	Cache   CacheStore
}

// NewService creates a [Service] wired with an [OllamaClient] from the config.
func NewService(cfg Config) *Service {
	return &Service{
		Config:  cfg,
		Vision:  NewOllamaClient(cfg.OllamaBaseURL, cfg.VisionModel, cfg.EmbedModel),
	}
}

// Lookup runs the full image-lookup pipeline on imagePath.
func (s *Service) Lookup(ctx context.Context, imagePath string) (LookupResult, error) {
	maxBytes := s.Config.MaxImageMB * 1024 * 1024
	if err := ValidateImage(imagePath, maxBytes); err != nil {
		return LookupResult{}, err
	}

	// Check the cache before running inference.
	imageHash, err := hashFile(imagePath)
	if err != nil {
		return LookupResult{}, fmt.Errorf("hash image: %w", err)
	}
	if s.Cache != nil {
		if cached, cacheErr := s.Cache.GetCachedAnalysis(ctx, imageHash, s.Config.VisionModel, PromptVersion()); cacheErr == nil && cached != nil {
			analysis := *cached
			return s.buildResult(analysis)
		}
	}

	// Normalize the image for inference (keep original for inspection).
	normalizedPath, cleanup, err := NormalizeImage(imagePath, s.Config.MaxDimension)
	if err != nil {
		return LookupResult{}, fmt.Errorf("normalize image: %w", err)
	}
	defer cleanup()

	analysis, err := s.Vision.AnalyzeImage(ctx, normalizedPath)
	if err != nil {
		return LookupResult{}, fmt.Errorf("vision analysis: %w", err)
	}

	// Cache the successful analysis.
	if s.Cache != nil {
		a := analysis
		_ = s.Cache.StoreCachedAnalysis(ctx, imageHash, s.Config.VisionModel, PromptVersion(), &a)
	}

	return s.buildResult(analysis)
}

// buildResult runs the catalog search against the analysis and assembles the
// final [LookupResult].
func (s *Service) buildResult(analysis ImageAnalysis) (LookupResult, error) {
	query := SearchText(analysis)

	result := LookupResult{
		Analysis:        analysis,
		NeedsMorePhotos: len(analysis.NextPhotosNeeded) > 0,
	}

	if s.Catalog != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		candidates, err := s.Catalog.Search(ctx, query, s.Config.MaxResults)
		if err != nil {
			return LookupResult{}, fmt.Errorf("search local catalog: %w", err)
		}
		result.Candidates = candidates
	}

	switch {
	case len(result.Candidates) == 0:
		result.Recommendation =
			"No close local-catalog match was found. Add clear photos " +
			"of logos, interior labels, hardware, and measurements."
	case result.Candidates[0].Score >= 0.82:
		result.Recommendation =
			"Strong candidate match. Compare the listed distinguishing details before labeling the item."
	case result.Candidates[0].Score >= 0.60:
		result.Recommendation =
			"Possible match. Collect the requested additional photos and compare the top candidates."
	default:
		result.Recommendation =
			"Low-confidence match. Use the extracted attributes to expand the local catalog or research manually."
	}

	return result, nil
}

// hashFile computes the SHA-256 hash of a file, used as a cache key.
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var h hash.Hash = sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
