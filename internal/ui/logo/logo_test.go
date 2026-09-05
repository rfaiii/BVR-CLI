package logo

import (
	"image/color"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestRenderTextWordmark(t *testing.T) {
	got := Render(lipgloss.NewStyle(), "", true, Opts{
		TitleColorA: color.White,
		TitleColorB: color.White,
		Text:        true,
	})

	if !strings.Contains(ansi.Strip(got), Wordmark) {
		t.Fatalf("Render() = %q, want it to contain %q", got, Wordmark)
	}
}

func TestRenderTextWordmarkIsStableForHyper(t *testing.T) {
	got := Render(lipgloss.NewStyle(), "", true, Opts{
		TitleColorA: color.White,
		TitleColorB: color.White,
		Hyper:       true,
		Text:        true,
	})
	plain := ansi.Strip(got)

	if !strings.Contains(plain, Wordmark) || strings.Contains(plain, "HYPER") {
		t.Fatalf("Render() = %q, want only the %q wordmark", got, Wordmark)
	}
}

func TestRenderWideWordmarkUsesTextAndGlitchFramesChange(t *testing.T) {
	base := lipgloss.NewStyle()
	first := Render(base, "", false, Opts{
		TitleColorA: color.White,
		TitleColorB: color.White,
		Width:       100,
		Text:        true,
		Animated:    true,
		Frame:       0,
	})
	second := Render(base, "", false, Opts{
		TitleColorA: color.White,
		TitleColorB: color.White,
		Width:       100,
		Text:        true,
		Animated:    true,
		Frame:       1,
	})

	plainFirst := ansi.Strip(first)
	if strings.Count(plainFirst, "█") < 20 {
		t.Fatalf("wide Render() = %q, want the custom ASCII banner", first)
	}
	if strings.Contains(plainFirst, Wordmark) {
		t.Fatalf("wide Render() unexpectedly used compact text wordmark: %q", first)
	}
	if first == second {
		t.Fatal("glitch pattern did not change between frames")
	}
}

func TestRenderWideWordmarkIncludesAnimationLine(t *testing.T) {
	got := Render(lipgloss.NewStyle(), "", false, Opts{
		TitleColorA: color.White,
		TitleColorB: color.White,
		Width:       100,
		Text:        true,
		Animated:    true,
		Animation:   "BANNER-ANIMATION",
	})

	if !strings.Contains(ansi.Strip(got), "BANNER-ANIMATION") {
		t.Fatalf("Render() = %q, want the animation line", got)
	}
}

func TestRenderBlockWordmark(t *testing.T) {
	got := Render(lipgloss.NewStyle(), "", false, Opts{
		TitleColorA: color.White,
		TitleColorB: color.White,
		Width:       120,
	})

	plain := ansi.Strip(got)
	if !strings.Contains(plain, "-█▀▀▀") {
		t.Fatalf("Render() = %q, want the block CLI suffix", got)
	}
}
