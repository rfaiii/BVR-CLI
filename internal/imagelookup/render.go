package imagelookup

import (
	"fmt"
	"strings"
)

// confidenceLabel returns a human-readable label for a confidence score.
func confidenceLabel(c Confidence) string {
	return string(c)
}

// RenderResult formats a [LookupResult] as human-readable CLI text, matching
// the documented command output format.
func RenderResult(result LookupResult) string {
	var b strings.Builder

	b.WriteString("LOOKUP IMAGE\n\n")

	if result.Analysis.ItemCategory.Value != "" {
		fmt.Fprintf(&b, "Category: %s (%s)\n",
			result.Analysis.ItemCategory.Value,
			confidenceLabel(result.Analysis.ItemCategory.Confidence))
	}
	if result.Analysis.ProbableBrand.Value != "" {
		fmt.Fprintf(&b, "Probable brand: %s (%s)\n",
			result.Analysis.ProbableBrand.Value,
			confidenceLabel(result.Analysis.ProbableBrand.Confidence))
	}

	if len(result.Analysis.DistinguishingCues) > 0 {
		b.WriteString("\nVisible cues:\n")
		for _, cue := range result.Analysis.DistinguishingCues {
			b.WriteString("- " + cue.Value)
			if cue.Confidence != ConfidenceNone {
				b.WriteString(fmt.Sprintf(" (%s)", confidenceLabel(cue.Confidence)))
			}
			if cue.Evidence != "" {
				b.WriteString(" — " + cue.Evidence)
			}
			b.WriteString("\n")
		}
	}

	if len(result.Analysis.VisibleText) > 0 {
		b.WriteString("\nVisible text:\n")
		for _, t := range result.Analysis.VisibleText {
			b.WriteString("- " + t.Value + "\n")
		}
	}

	if len(result.Analysis.LogosAndMarks) > 0 {
		b.WriteString("\nLogos and marks:\n")
		for _, lm := range result.Analysis.LogosAndMarks {
			b.WriteString("- " + lm.Value + "\n")
		}
	}

	if len(result.Candidates) > 0 {
		b.WriteString("\nTop catalog candidates:\n")
		for i, c := range result.Candidates {
			fmt.Fprintf(&b, "%d. %s — %.2f\n", i+1, c.Title, c.Score)
			if len(c.MatchedSignals) > 0 {
				b.WriteString("   Matches: " + strings.Join(c.MatchedSignals, ", ") + "\n")
			}
			if len(c.Conflicts) > 0 {
				b.WriteString("   Conflicts: " + strings.Join(c.Conflicts, ", ") + "\n")
			}
		}
	} else {
		b.WriteString("\nNo close local-catalog match was found.\n")
	}

	if len(result.Analysis.NextPhotosNeeded) > 0 {
		b.WriteString("\nRecommended next photos:\n")
		for _, p := range result.Analysis.NextPhotosNeeded {
			b.WriteString("- " + p + "\n")
		}
	}

	if result.Recommendation != "" {
		b.WriteString("\n" + result.Recommendation + "\n")
	}

	b.WriteString("\nNote: Image-based identification is not authentication.\n")
	return b.String()
}
