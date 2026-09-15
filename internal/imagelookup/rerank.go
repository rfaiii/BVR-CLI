package imagelookup

import "strings"

// SignalWeights defines the relative importance of each signal in the
// deterministic scoring formula. The default weights sum to 1.0:
//
//	score = 0.30·Brand + 0.20·Category + 0.15·Pattern + 0.10·Material
//	       + 0.10·Hardware + 0.10·Silhouette + 0.05·Dimensions
var DefaultSignalWeights = SignalWeights{
	Brand:      0.30,
	Category:   0.20,
	Pattern:    0.15,
	Material:   0.10,
	Hardware:   0.10,
	Silhouette: 0.10,
	Dimensions: 0.05,
}

// SignalWeights holds the per-signal multipliers used by [ComputeScore].
type SignalWeights struct {
	Brand      float64
	Category   float64
	Pattern    float64
	Material   float64
	Hardware   float64
	Silhouette float64
	Dimensions float64
}

// SignalMatches holds per-signal match scores (0.0–1.0) for a single candidate
// product against the extracted analysis, along with human-readable lists of
// which signals matched and which conflicted.
type SignalMatches struct {
	Brand      float64
	Category   float64
	Pattern    float64
	Material   float64
	Hardware   float64
	Silhouette float64
	Dimensions float64

	MatchedSignals []string
	Conflicts      []string
}

// ComputeScore applies the weighted scoring formula to a set of [SignalMatches]
// using the given [SignalWeights]. The result is in the range 0–1.
func ComputeScore(m SignalMatches, w SignalWeights) float64 {
	return w.Brand*m.Brand +
		w.Category*m.Category +
		w.Pattern*m.Pattern +
		w.Material*m.Material +
		w.Hardware*m.Hardware +
		w.Silhouette*m.Silhouette +
		w.Dimensions*m.Dimensions
}

// ConfidenceToScore converts a [Confidence] level to a 0–1 match score used by
// the deterministic scorer. High confidence maps to 1.0; medium to 0.7; low to
// 0.4; none to 0.0.
func ConfidenceToScore(c Confidence) float64 {
	switch c {
	case ConfidenceHigh:
		return 1.0
	case ConfidenceMedium:
		return 0.7
	case ConfidenceLow:
		return 0.4
	default:
		return 0.0
	}
}

// ParsedSignals extracts structured signals from the SearchText query string.
// The query has the form "category: handbag; brand: Louis Vuitton; logo: LV; ..."
// so we split on ";" and parse "label: value" pairs.
type ParsedSignals struct {
	Category    string
	Brand       string
	Logos       []string
	Silhouettes []string
	Materials   []string
	Colors      []string
	Patterns    []string
	Hardware    []string
	Cues        []string
	CategoryConf   Confidence
	BrandConf   Confidence
}

// ParseSignals parses a SearchText query string into [ParsedSignals]. If a
// signal cannot be parsed from the query, the corresponding field is left
// empty.
func ParseSignals(query string) ParsedSignals {
	s := ParsedSignals{
		CategoryConf: ConfidenceNone,
		BrandConf:    ConfidenceNone,
	}
	parts := strings.Split(query, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		colon := strings.Index(part, ":")
		if colon < 0 {
			continue
		}
		label := strings.TrimSpace(part[:colon])
		value := strings.TrimSpace(part[colon+1:])
		switch label {
		case "category":
			s.Category = value
			s.CategoryConf = ConfidenceMedium
		case "brand":
			s.Brand = value
			s.BrandConf = ConfidenceMedium
		case "logo":
			s.Logos = append(s.Logos, value)
		case "silhouette":
			s.Silhouettes = append(s.Silhouettes, value)
		case "material":
			s.Materials = append(s.Materials, value)
		case "color":
			s.Colors = append(s.Colors, value)
		case "pattern":
			s.Patterns = append(s.Patterns, value)
		case "hardware":
			s.Hardware = append(s.Hardware, value)
		case "cue":
			s.Cues = append(s.Cues, value)
		}
	}
	return s
}

// MatchString checks whether a candidate value matches a target string in a
// case-insensitive, substring-aware way. Both strings are normalised to lower
// case and trimmed of surrounding whitespace.
func MatchString(candidate, target string) bool {
	c := strings.TrimSpace(strings.ToLower(candidate))
	t := strings.TrimSpace(strings.ToLower(target))
	if c == "" || t == "" {
		return false
	}
	return strings.Contains(c, t) || strings.Contains(t, c)
}
