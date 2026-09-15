package imagelookup

import "strings"

// SearchText turns the extracted [ImageAnalysis] into a compact,
// label-delimited string suitable for SQLite FTS matching and embedding.
// Only observations with a non-empty value and a confidence other than
// "none" are included.
func SearchText(a ImageAnalysis) string {
	var parts []string

	add := func(label string, values []Observation) {
		for _, v := range values {
			if v.Value != "" && v.Confidence != ConfidenceNone {
				parts = append(parts, label+": "+v.Value)
			}
		}
	}

	if a.ItemCategory.Value != "" {
		parts = append(parts, "category: "+a.ItemCategory.Value)
	}
	if a.ProbableBrand.Value != "" &&
		a.ProbableBrand.Confidence != ConfidenceNone {
		parts = append(parts, "brand: "+a.ProbableBrand.Value)
	}

	add("visible text", a.VisibleText)
	add("logo", a.LogosAndMarks)
	add("silhouette", a.Silhouette)
	add("material", a.Materials)
	add("color", a.Colors)
	add("pattern", a.Patterns)
	add("hardware", a.Hardware)
	add("construction", a.Construction)
	add("cue", a.DistinguishingCues)

	return strings.Join(parts, "; ")
}
