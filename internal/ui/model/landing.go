package model

import (
	"image"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/ultraviolet/layout"
	"github.com/richavery/bvr-cli/internal/ggwave"
	"github.com/richavery/bvr-cli/internal/home"
	"github.com/richavery/bvr-cli/internal/workspace"
)

// selectedLargeModel returns the currently selected large language model from
// the agent coordinator, if one exists.
func (m *UI) selectedLargeModel() *workspace.AgentModel {
	if m.com.Workspace.AgentIsReady() {
		model := m.com.Workspace.AgentModel()
		return &model
	}
	return nil
}

// landingView renders the landing page view showing the current working
// directory, model information, and LSP/MCP status in a two-column layout.
func (m *UI) landingView() string {
	t := m.com.Styles
	width := m.layout.main.Dx()
	cwd := home.Short(m.com.Workspace.WorkingDir())

	accent := t.ThemeColor.Accent
	alt := t.ThemeColor.Alt
	if accent == nil {
		accent = lipgloss.Color("#FF4FA3")
	}
	if alt == nil {
		alt = lipgloss.Color("#B56CFF")
	}

	// Bold, underlined project location rendered in the accent color so it
	// reads clearly against the dark background.
	cwdStyled := lipgloss.NewStyle().
		Foreground(accent).
		Bold(true).
		Underline(true).
		Render(cwd)

	// Home buttons: Command palette launcher, File Finder, Create File,
	// and Chat with Grok Bot, each prefixed with a nerd-font glyph from
	// the superfile icon set so they read as actionable buttons.
	// Arranged vertically with a blank line between each for even spacing.
	iconColor := t.Header.LogoGradToColor
	terminalIcon := lipgloss.NewStyle().Foreground(iconColor).Render("\ue795") // superfile icon.Terminal
	folderIcon := lipgloss.NewStyle().Foreground(iconColor).Render("\uf07b")   // superfile icon.Directory
	paperIcon := lipgloss.NewStyle().Foreground(iconColor).Render("\uf15b")    // superfile icon.File
	// nf-fa-comment. superfile's icon.go has no Chat constant; this is the
	// comment glyph from the same Font Awesome set as icon.Directory and
	// icon.File, so it reads as a conversation next to those buttons.
	chatIcon := lipgloss.NewStyle().Foreground(iconColor).Render("\uf075")

	buttonStyle := lipgloss.NewStyle().
		Foreground(accent).
		Bold(true).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Padding(0, 1)

	commandButton := buttonStyle.Render(terminalIcon + " " + "OPEN COMMANDS — ctrl+p")
	folderButton := buttonStyle.Render(folderIcon + " " + "OPEN FILE FINDER — ctrl+o")
	createButton := buttonStyle.Render(paperIcon + " " + "CREATE FILE — ctrl+n")
	botsBtn := buttonStyle.Render(chatIcon + " " + landingBotsButton)

	// Stack buttons vertically (one above the other) with a blank line
	// between each for even, balanced spacing.
	buttons := lipgloss.JoinVertical(lipgloss.Left, commandButton, "", folderButton, "", createButton, "", botsBtn)

	// Prominent MODEL / PROVIDER line on the homescreen so it's immediately
	// visible. Uses ACCENT for the values and ALT for the labels.
	modelName := ""
	providerName := ""
	if lm := m.selectedLargeModel(); lm != nil {
		modelName = lm.CatwalkCfg.Name
		if pcfg, ok := m.com.Config().Providers.Get(lm.ModelCfg.Provider); ok {
			providerName = pcfg.Name
		}
	}
	modelLine := lipgloss.NewStyle().Foreground(alt).Render("MODEL  ") +
		lipgloss.NewStyle().Foreground(accent).Bold(true).Render(modelName) +
		lipgloss.NewStyle().Foreground(alt).Render("   PROVIDER  ") +
		lipgloss.NewStyle().Foreground(accent).Bold(true).Render(providerName)

	// Dense ASCII BVR mascot from the boot splash (boot.BeaverFramesDenseAlpha).
	// The pose is resolved by beaverFrameKey so the homescreen and the
	// working-screen sidebar animate identically. The mascot is never wider
	// than 13 cells, so the layout and click hitbox stay stable.
	hero := m.beaverFrame()

	// Buttons now appear before the hero (near the header). The hero is
	// positioned after the CWD line, blank, and the vertical button stack.
	buttonsHeight := lipgloss.Height(buttons)
	heroTop := m.layout.main.Min.Y + 1 + lipgloss.Height(cwdStyled) + 1 + buttonsHeight + 1
	m.beaverRect = image.Rect(
		m.layout.main.Min.X,
		heroTop,
		m.layout.main.Min.X+lipgloss.Width(hero),
		heroTop+lipgloss.Height(hero),
	)

	// Click rectangles for the vertical home button stack.
	// Buttons are stacked top-to-bottom with a 1-row gap between each.
	// Use each button's actual rendered height for accurate hit detection.
	btnTop := landingButtonTop(m.layout.main.Min.Y, lipgloss.Height(cwdStyled))
	cmdH := lipgloss.Height(commandButton)
	folderH := lipgloss.Height(folderButton)
	createH := lipgloss.Height(createButton)
	imageH := lipgloss.Height(botsBtn)
	btnGap := 1 // blank line between buttons (from the "" separators)
	col1Width := lipgloss.Width(commandButton)
	col2Width := lipgloss.Width(folderButton)
	col3Width := lipgloss.Width(createButton)
	col4Width := lipgloss.Width(botsBtn)
	x0 := m.layout.main.Min.X

	y := btnTop
	m.commandButtonRect = image.Rect(x0, y, x0+col1Width, y+cmdH)
	y += cmdH + btnGap
	m.finderButtonRect = image.Rect(x0, y, x0+col2Width, y+folderH)
	y += folderH + btnGap
	m.createFileButtonRect = image.Rect(x0, y, x0+col3Width, y+createH)
	y += createH + btnGap
	m.botsButtonRect = image.Rect(x0, y, x0+col4Width, y+imageH)

	waveWidth := min(56, max(18, width-4))
	waveLabel := lipgloss.NewStyle().Foreground(alt).Render("GGWAVE NODE LINK  ")
	wave := lipgloss.NewStyle().Foreground(accent).Render(ggwave.Waveform(waveWidth, m.ggwaveFrame, m.ggwaveMode, m.ggwaveSpeed))

	// Layout order (top to bottom):
	//   CWD line → blank → buttons (vertical stack) → blank → hero → blank →
	//   model line → blank → ggwave waveform
	parts := []string{cwdStyled, "", buttons, "", hero, "", modelLine, "", waveLabel + wave}
	infoSection := lipgloss.JoinVertical(lipgloss.Left, parts...)

	var remainingHeightArea image.Rectangle
	layout.Vertical(
		layout.Len(lipgloss.Height(infoSection)+1),
		layout.Fill(1),
	).Split(m.layout.main).Assign(new(image.Rectangle), &remainingHeightArea)

	// Left column: NODE, MCP, LSP stacked vertically (top to bottom). The right
	// side holds SKILLS. These are short status monitors, so they share a narrow
	// left column, leaving the bulk of the width for the (often long) skill list.
	leftW := min(30, (width-2)/3)
	rightW := max(1, width-leftW-1)
	sectionH := max(1, remainingHeightArea.Dy())

	nodeSection := m.nodeInfo(leftW, sectionH)
	mcpSection := m.mcpInfo(leftW, sectionH, false)
	lspSection := m.lspInfo(leftW, sectionH, false)
	skillsSection := m.skillsInfo(rightW, sectionH, false)

	leftColumn := lipgloss.JoinVertical(lipgloss.Left, nodeSection, " ", mcpSection, " ", lspSection)
	content := lipgloss.JoinHorizontal(lipgloss.Left, leftColumn, " ", skillsSection)

	return lipgloss.NewStyle().
		Width(width).
		Height(m.layout.main.Dy()).
		Render(
			lipgloss.JoinVertical(lipgloss.Left, infoSection, "", content),
		)
}

// landingBotsButton is the homescreen control that opens the bot picker.
const landingBotsButton = "CHAT W/ GROK BOT — ctrl+b"

// landingButtonTop returns the Y coordinate for the top of the vertical
// button stack on the landing page. Buttons appear after the CWD line and
// a single blank separator line.
func landingButtonTop(mainY, cwdHeight int) int {
	return mainY + 1 + cwdHeight
}
