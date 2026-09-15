package model

import (
	"image"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/ultraviolet/layout"
	"github.com/richavery/bvr-cli/internal/ggwave"
	"github.com/richavery/bvr-cli/internal/home"
	"github.com/richavery/bvr-cli/internal/ui/anim"
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
	// and Image Recognition, each prefixed with a nerd-font glyph from
	// the superfile icon set so they read as actionable buttons.
	// Arranged in a single row with even spacing.
	iconColor := t.Header.LogoGradToColor
	terminalIcon := lipgloss.NewStyle().Foreground(iconColor).Render("\ue795") // superfile icon.Terminal
	folderIcon := lipgloss.NewStyle().Foreground(iconColor).Render("\uf07b") // superfile icon.Directory
	paperIcon := lipgloss.NewStyle().Foreground(iconColor).Render("\uf15b")  // superfile icon.File
	cameraIcon := lipgloss.NewStyle().Foreground(iconColor).Render("\uf030") // superfile icon.Camera

	buttonStyle := lipgloss.NewStyle().
		Foreground(accent).
		Bold(true).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Padding(0, 1)

	commandButton := buttonStyle.Render(terminalIcon + " " + "OPEN COMMANDS — ctrl+p")
	folderButton := buttonStyle.Render(folderIcon + " " + "OPEN FILE FINDER — ctrl+o")
	createButton := buttonStyle.Render(paperIcon + " " + "CREATE FILE — ctrl+n")
	imageBtn := buttonStyle.Render(cameraIcon + " " + "IMAGE RECOGNITION — ctrl+b")

	buttons := commandButton + "  " + folderButton + "  " + createButton + "  " + imageBtn

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
	mascotState := m.beaverGaze
	if time.Now().Before(m.beaverBoopUntil) {
		mascotState = anim.StateClickBoop
	}
	hero := anim.LargeMascotFrame(mascotState, m.bannerFrame, m.beaverErrored)
	heroTop := m.layout.main.Min.Y + 1 + lipgloss.Height(cwdStyled) + 1
	m.beaverRect = image.Rect(
		m.layout.main.Min.X,
		heroTop,
		m.layout.main.Min.X+lipgloss.Width(hero),
		heroTop+lipgloss.Height(hero),
	)

	// Click rectangles for the single-row home button grid.
	// All 4 buttons are on the same row with even spacing.
	btnTop := landingButtonTop(m.layout.main.Min.Y, lipgloss.Height(cwdStyled), lipgloss.Height(hero))
	cmdH := lipgloss.Height(commandButton)
	folderH := lipgloss.Height(folderButton)
	createH := lipgloss.Height(createButton)
	imageH := lipgloss.Height(imageBtn)
	btnGap := 2 // gap between each button
	col1Width := lipgloss.Width(commandButton)
	col2Width := lipgloss.Width(folderButton)
	col3Width := lipgloss.Width(createButton)
	col4Width := lipgloss.Width(imageBtn)
	rowHeight := max(cmdH, max(folderH, max(createH, imageH)))
	// Single row: buttons placed side by side with even gaps
	x0 := m.layout.main.Min.X
	m.commandButtonRect = image.Rect(x0, btnTop, x0+col1Width, btnTop+rowHeight)
	x1 := x0 + col1Width + btnGap
	m.finderButtonRect = image.Rect(x1, btnTop, x1+col2Width, btnTop+rowHeight)
	x2 := x1 + col2Width + btnGap
	m.createFileButtonRect = image.Rect(x2, btnTop, x2+col3Width, btnTop+rowHeight)
	x3 := x2 + col3Width + btnGap
	m.imageButtonRect = image.Rect(x3, btnTop, x3+col4Width, btnTop+rowHeight)

	waveWidth := min(56, max(18, width-4))
	waveLabel := lipgloss.NewStyle().Foreground(alt).Render("GGWAVE NODE LINK  ")
	wave := lipgloss.NewStyle().Foreground(accent).Render(ggwave.Waveform(waveWidth, m.ggwaveFrame, m.ggwaveMode, m.ggwaveSpeed))
	parts := []string{cwdStyled, "", hero, "", buttons, "", modelLine, "", waveLabel + wave}
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

func landingButtonTop(mainY, cwdHeight, heroHeight int) int {
	return mainY + cwdHeight + 1 + heroHeight + 1
}
