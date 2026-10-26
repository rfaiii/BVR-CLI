package styles

import (
	"fmt"
	"image/color"
	"strings"
	"sync"

	"charm.land/lipgloss/v2"
	"github.com/rivo/uniseg"
)

// gradCache memoizes finished gradient renders. The same wordmark gradient is
// rebuilt on every frame by the logo/header code; keys include the base style
// so distinct style instances never share a result incorrectly. Entries stay
// small: one per (input, colors, base) combination actually used per theme.
var gradCache sync.Map

func gradKey(input string, bold bool, base lipgloss.Style, color1, color2 color.Color) string {
	return fmt.Sprintf("%t|%v|%v|%v|%s", bold, color1, color2, base, input)
}

// ForegroundGrad returns a slice of strings representing the input string
// rendered with a horizontal gradient foreground from color1 to color2. Each
// string in the returned slice corresponds to a grapheme cluster in the input
// string. If bold is true, the rendered strings will be bolded.
func ForegroundGrad(base lipgloss.Style, input string, bold bool, color1, color2 color.Color) []string {
	if input == "" {
		return []string{""}
	}
	if len(input) == 1 {
		style := base.Foreground(color1)
		if bold {
			style.Bold(true)
		}
		return []string{style.Render(input)}
	}
	var clusters []string
	gr := uniseg.NewGraphemes(input)
	for gr.Next() {
		clusters = append(clusters, string(gr.Runes()))
	}

	ramp := lipgloss.Blend1D(len(clusters), color1, color2)
	for i, c := range ramp {
		style := base.Foreground(c)
		if bold {
			style.Bold(true)
		}
		clusters[i] = style.Render(clusters[i])
	}
	return clusters
}

// ApplyForegroundGrad renders a given string with a horizontal gradient
// foreground.
func ApplyForegroundGrad(base lipgloss.Style, input string, color1, color2 color.Color) string {
	if input == "" {
		return ""
	}
	key := gradKey(input, false, base, color1, color2)
	if cached, ok := gradCache.Load(key); ok {
		return cached.(string)
	}
	var o strings.Builder
	clusters := ForegroundGrad(base, input, false, color1, color2)
	for _, c := range clusters {
		fmt.Fprint(&o, c)
	}
	out := o.String()
	gradCache.Store(key, out)
	return out
}

// ApplyBoldForegroundGrad renders a given string with a horizontal gradient
// foreground.
func ApplyBoldForegroundGrad(base lipgloss.Style, input string, color1, color2 color.Color) string {
	if input == "" {
		return ""
	}
	key := gradKey(input, true, base, color1, color2)
	if cached, ok := gradCache.Load(key); ok {
		return cached.(string)
	}
	var o strings.Builder
	clusters := ForegroundGrad(base, input, true, color1, color2)
	for _, c := range clusters {
		fmt.Fprint(&o, c)
	}
	out := o.String()
	gradCache.Store(key, out)
	return out
}
