package imagelookup

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSearchText_BasicFields(t *testing.T) {
	a := ImageAnalysis{
		ItemCategory: Observation{Value: "handbag", Confidence: ConfidenceHigh, Evidence: "tote silhouette"},
		ProbableBrand: Observation{
			Value:      "Louis Vuitton",
			Confidence: ConfidenceMedium,
			Evidence:   "LV monogram visible",
		},
		VisibleText: []Observation{
			{Value: "Louis Vuitton", Confidence: ConfidenceHigh, Evidence: "on stamp"},
		},
		Materials: []Observation{
			{Value: "coated canvas", Confidence: ConfidenceHigh, Evidence: "canvas texture visible"},
		},
		Colors: []Observation{
			{Value: "brown", Confidence: ConfidenceHigh, Evidence: "monogram canvas color"},
		},
		Patterns: []Observation{
			{Value: "LV monogram", Confidence: ConfidenceMedium, Evidence: "repeating pattern"},
		},
	}

	text := SearchText(a)
	require.Contains(t, text, "category: handbag")
	require.Contains(t, text, "brand: Louis Vuitton")
	require.Contains(t, text, "visible text: Louis Vuitton")
	require.Contains(t, text, "material: coated canvas")
	require.Contains(t, text, "color: brown")
	require.Contains(t, text, "pattern: LV monogram")
}

func TestSearchText_NoneConfidenceExcluded(t *testing.T) {
	a := ImageAnalysis{
		ProbableBrand: Observation{
			Value:      "Some Brand",
			Confidence: ConfidenceNone,
			Evidence:   "",
		},
		VisibleText: []Observation{
			{Value: "readable text", Confidence: ConfidenceHigh, Evidence: "stamp"},
			{Value: "unreadable", Confidence: ConfidenceNone, Evidence: ""},
		},
	}

	text := SearchText(a)
	require.NotContains(t, text, "brand: Some Brand")
	require.Contains(t, text, "visible text: readable text")
	require.NotContains(t, text, "unreadable")
}

func TestSearchText_EmptyAnalysis(t *testing.T) {
	text := SearchText(ImageAnalysis{})
	require.Empty(t, text)
}
