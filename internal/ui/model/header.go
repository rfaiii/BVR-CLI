package model

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/richavery/bvr-cli/internal/config"
	"github.com/richavery/bvr-cli/internal/home"
	"github.com/richavery/bvr-cli/internal/session"
	"github.com/richavery/bvr-cli/internal/ui/common"
	"github.com/richavery/bvr-cli/internal/ui/logo"
	"github.com/richavery/bvr-cli/internal/ui/styles"
)

const (
	headerDiag           = "╱"
	minHeaderDiags       = 3
	leftPadding          = 1
	rightPadding         = 1
	diagToDetailsSpacing = 1 // space between diagonal pattern and details section
)

type header struct {
	// cached logo and compact logo
	logo        string
	compactLogo string

	com     *common.Common
	width   int
	compact bool
	frame   int
}

// newHeader creates a new header model.
func newHeader(com *common.Common) *header {
	h := &header{
		com: com,
	}
	h.refresh()
	return h
}

// refresh rebuilds cached logo strings using the current styles. Call
// after the theme changes.
func (h *header) refresh() {
	t := h.com.Styles
	isHyper := h.com.IsHyper()
	_ = isHyper
	h.compactLogo = styles.ApplyBoldForegroundGrad(t.Header.LogoGradCanvas, logo.Wordmark, t.Header.LogoGradFromColor, t.Header.LogoGradToColor) + " "
	// Force drawHeader to re-render the wide logo on the next frame.
	h.width = 0
	h.logo = ""
}

// drawHeader draws the header for the given session.
func (h *header) drawHeader(
	scr uv.Screen,
	area uv.Rectangle,
	session *session.Session,
	compact bool,
	detailsOpen bool,
	width int,
	hyperCredits *int,
	frame int,
	animation string,
	coderMode bool,
	projectSource string,
) {
	t := h.com.Styles
	if width != h.width || compact != h.compact || frame != h.frame {
		h.logo = renderLogo(h.com.Styles, compact, h.com.IsHyper(), width, frame, animation)
	}

	h.width = width
	h.compact = compact
	h.frame = frame

	if !compact || session == nil {
		uv.NewStyledString(h.logo).Draw(scr, area)
		return
	}

	if session.ID == "" {
		return
	}

	var b strings.Builder
	b.WriteString(h.compactLogo)

	availDetailWidth := width - leftPadding - rightPadding - lipgloss.Width(b.String()) - minHeaderDiags - diagToDetailsSpacing
	lspErrorCount := 0
	for _, info := range h.com.Workspace.LSPGetStates() {
		lspErrorCount += info.DiagnosticCount
	}
	details := renderHeaderDetails(
		h.com,
		session,
		lspErrorCount,
		detailsOpen,
		availDetailWidth,
		hyperCredits,
		coderMode,
		projectSource,
	)

	remainingWidth := width -
		lipgloss.Width(b.String()) -
		lipgloss.Width(details) -
		leftPadding -
		rightPadding -
		diagToDetailsSpacing

	if remainingWidth > 0 {
		b.WriteString(t.Header.Diagonals.Render(
			strings.Repeat(headerDiag, max(minHeaderDiags, remainingWidth)),
		))
		b.WriteString(" ")
	}

	b.WriteString(details)

	view := uv.NewStyledString(
		t.Header.Wrapper.Padding(0, rightPadding, 0, leftPadding).Render(b.String()),
	)
	view.Draw(scr, area)
}

// renderHeaderDetails renders the details section of the header.
func renderHeaderDetails(
	com *common.Common,
	session *session.Session,
	lspErrorCount int,
	detailsOpen bool,
	availWidth int,
	hyperCredits *int,
	coderMode bool,
	projectSource string,
) string {
	t := com.Styles

	var parts []string

	if lspErrorCount > 0 {
		parts = append(parts, t.LSP.ErrorDiagnostic.Render(fmt.Sprintf("%s%d", styles.LSPErrorIcon, lspErrorCount)))
	}

	agentCfg := com.Config().Agents[config.AgentCoder]
	model := com.Config().GetModelByType(agentCfg.Model)
	if model != nil && model.ContextWindow > 0 {
		percentage := (float64(session.CompletionTokens+session.PromptTokens) / float64(model.ContextWindow)) * 100
		percentageText := fmt.Sprintf("%d%%", int(percentage))
		if session.EstimatedUsage {
			percentageText = "~" + percentageText
		}
		formattedPercentage := t.Header.Percentage.Render(percentageText)
		parts = append(parts, formattedPercentage)
	}

	if com.IsHyper() && hyperCredits != nil {
		hc := t.Header.HypercreditIcon.Render(styles.HypercreditIcon) + " " + t.Header.Percentage.Render(common.FormatCredits(*hyperCredits))
		parts = append(parts, hc)
	}

	const keystroke = "ctrl+d"
	if detailsOpen {
		parts = append(parts, t.Header.Keystroke.Render(keystroke)+t.Header.KeystrokeTip.Render(" close"))
	} else {
		parts = append(parts, t.Header.Keystroke.Render(keystroke)+t.Header.KeystrokeTip.Render(" open "))
	}

	if coderMode {
		parts = append(parts, t.Header.Percentage.Render("CODER"))
	}

	dot := t.Header.Separator.Render(" • ")
	metadata := strings.Join(parts, dot)
	metadata = dot + metadata

	// Display project source in the header (dynamic, changeable via file browser)
	projectSrc := home.Short(projectSource)
	projectSrcStyled := t.Header.WorkingDir.Render("📁 " + projectSrc + "  ")

	result := projectSrcStyled + metadata
	return ansi.Truncate(result, max(0, availWidth), "…")
}
