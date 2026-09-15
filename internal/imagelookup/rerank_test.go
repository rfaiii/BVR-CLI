package imagelookup

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComputeScore_WeightedSum(t *testing.T) {
	t.Parallel()
	m := SignalMatches{
		Brand:      1.0,
		Category:   1.0,
		Pattern:    0.0,
		Material:   1.0,
		Hardware:   0.0,
		Silhouette: 0.3,
		Dimensions: 0.0,
	}
	// 0.30*1 + 0.20*1 + 0.15*0 + 0.10*1 + 0.10*0 + 0.10*0.3 + 0.05*0
	// = 0.30 + 0.20 + 0 + 0.10 + 0 + 0.03 + 0 = 0.63
	score := ComputeScore(m, DefaultSignalWeights)
	require.InDelta(t, 0.63, score, 0.001)
}

func TestComputeScore_PerfectMatch(t *testing.T) {
	t.Parallel()
	m := SignalMatches{
		Brand:      1.0,
		Category:   1.0,
		Pattern:    1.0,
		Material:   1.0,
		Hardware:   1.0,
		Silhouette: 1.0,
		Dimensions: 1.0,
	}
	score := ComputeScore(m, DefaultSignalWeights)
	require.InDelta(t, 1.0, score, 0.001)
}

func TestComputeScore_NoMatch(t *testing.T) {
	t.Parallel()
	m := SignalMatches{}
	score := ComputeScore(m, DefaultSignalWeights)
	require.Zero(t, score)
}

func TestConfidenceToScore(t *testing.T) {
	t.Parallel()
	tests := []struct {
		c    Confidence
		want float64
	}{
		{ConfidenceHigh, 1.0},
		{ConfidenceMedium, 0.7},
		{ConfidenceLow, 0.4},
		{ConfidenceNone, 0.0},
		{Confidence("unknown"), 0.0},
	}
	for _, tt := range tests {
		t.Run(string(tt.c), func(t *testing.T) {
			require.Equal(t, tt.want, ConfidenceToScore(tt.c))
		})
	}
}

func TestParseSignals_ExtractsLabelsAndValues(t *testing.T) {
	t.Parallel()
	query := "category: handbag; brand: Louis Vuitton; logo: LV; material: coated canvas"
	s := ParseSignals(query)
	require.Equal(t, "handbag", s.Category)
	require.Equal(t, "Louis Vuitton", s.Brand)
	require.Contains(t, s.Logos, "LV")
	require.Contains(t, s.Materials, "coated canvas")
}

func TestParseSignals_EmptyQuery(t *testing.T) {
	t.Parallel()
	s := ParseSignals("")
	require.Empty(t, s.Category)
	require.Empty(t, s.Brand)
	require.Empty(t, s.Logos)
}

func TestMatchString(t *testing.T) {
	t.Parallel()
	require.True(t, MatchString("Louis Vuitton", "louis vuitton"))
	require.True(t, MatchString("brown monogram canvas", "monogram"))
	require.False(t, MatchString("", "test"))
	require.False(t, MatchString("test", ""))
}
