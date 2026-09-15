package imagelookup

import (
	"os"
	"strconv"
)

// Config holds configuration for the image-lookup feature. All fields can be
// populated from environment variables via [FromEnv].
type Config struct {
	OllamaBaseURL string
	VisionModel   string
	EmbedModel    string
	CatalogDB     string
	MaxImageMB    int64
	MaxDimension  int
	MaxResults    int
}

// DefaultConfig returns sensible defaults for local, Ollama-based image
// identification. The defaults point at a catalog database relative to the
// current working directory so the feature works out of the box.
func DefaultConfig() Config {
	return Config{
		OllamaBaseURL: "http://localhost:11434",
		VisionModel:   "qwen2.5vl:3b",
		EmbedModel:    "nomic-embed-text",
		CatalogDB:     "./data/catalog.db",
		MaxImageMB:    12,
		MaxDimension:  1600,
		MaxResults:    10,
	}
}

// FromEnv builds a [Config] from environment variables, falling back to
// [DefaultConfig] for any variable that is unset or invalid.
func FromEnv() Config {
	cfg := DefaultConfig()
	if v := os.Getenv("OLLAMA_BASE_URL"); v != "" {
		cfg.OllamaBaseURL = v
	}
	if v := os.Getenv("BVR_VISION_MODEL"); v != "" {
		cfg.VisionModel = v
	}
	if v := os.Getenv("BVR_EMBED_MODEL"); v != "" {
		cfg.EmbedModel = v
	}
	if v := os.Getenv("BVR_CATALOG_DB"); v != "" {
		cfg.CatalogDB = v
	}
	if v := os.Getenv("BVR_LOOKUP_MAX_IMAGE_MB"); v != "" {
		if mb, err := parseInt64(v); err == nil && mb > 0 {
			cfg.MaxImageMB = mb
		}
	}
	if v := os.Getenv("BVR_LOOKUP_MAX_DIMENSION"); v != "" {
		if dim, err := parseInt(v); err == nil && dim > 0 {
			cfg.MaxDimension = dim
		}
	}
	return cfg
}

// PromptVersion returns a string that identifies the current vision-prompt
// and schema version. It is used as part of the image-analysis cache key so
// that changing the prompt invalidates stale cached results.
func PromptVersion() string {
	return "v1"
}

// parseInt is a small helper that parses an integer from a string.
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// parseInt64 is a small helper that parses a 64-bit integer from a string.
func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// Confidence represents how certain the vision model is that an observation
// is correct.
type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
	ConfidenceNone   Confidence = "none"
)

// Observation is a single visible attribute extracted from the image, along
// with the model's self-assessed confidence and supporting evidence.
type Observation struct {
	Value      string     `json:"value"`
	Confidence Confidence `json:"confidence"`
	Evidence   string     `json:"evidence"`
}

// ImageAnalysis is the structured output expected from the vision model. Every
// field is an *observation*, not a conclusion. The model must never determine
// authenticity.
type ImageAnalysis struct {
	ImageQuality       string        `json:"image_quality"`
	ItemCategory       Observation   `json:"item_category"`
	ProbableBrand      Observation   `json:"probable_brand"`
	VisibleText        []Observation `json:"visible_text"`
	LogosAndMarks      []Observation `json:"logos_and_marks"`
	Silhouette         []Observation `json:"silhouette"`
	Materials          []Observation `json:"materials"`
	Colors             []Observation `json:"colors"`
	Patterns           []Observation `json:"patterns"`
	Hardware           []Observation `json:"hardware"`
	Construction       []Observation `json:"construction"`
	DimensionsEstimate []Observation `json:"dimensions_estimate"`
	DistinguishingCues []Observation `json:"distinguishing_cues"`
	DamageOrWear       []Observation `json:"damage_or_wear"`
	Unknowns           []string      `json:"unknowns"`
	NextPhotosNeeded   []string      `json:"next_photos_needed"`
	AuthenticityNote   string        `json:"authenticity_note"`
}

// Candidate is a catalog product returned as a possible match for the
// analyzed image.
type Candidate struct {
	ProductID      string   `json:"product_id"`
	Brand          string   `json:"brand"`
	Title          string   `json:"title"`
	Category       string   `json:"category"`
	ReferenceURL   string   `json:"reference_url"`
	ImagePath      string   `json:"image_path,omitempty"`
	Score          float64  `json:"score"`
	MatchedSignals []string `json:"matched_signals"`
	Conflicts      []string `json:"conflicts"`
}

// LookupResult is the complete output of an image lookup: the vision-model
// analysis, ranked catalog candidates, and a human-readable recommendation.
type LookupResult struct {
	Analysis        ImageAnalysis `json:"analysis"`
	Candidates      []Candidate   `json:"candidates"`
	Recommendation  string        `json:"recommendation"`
	NeedsMorePhotos bool          `json:"needs_more_photos"`
}
