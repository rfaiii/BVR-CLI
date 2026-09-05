package model

import (
	"image"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/ultraviolet/layout"
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

	// Home buttons: Command palette launcher (top) and File Finder (bottom),
	// each prefixed with a nerd-font glyph from the superfile icon set so
	// they read as actionable buttons. They are stacked vertically with a
	// blank line between them for cushioning, left-aligned one above the
	// other.
	iconColor := t.Header.LogoGradToColor
	terminalIcon := lipgloss.NewStyle().Foreground(iconColor).Render("\ue795")    // superfile icon.Terminal
	folderIcon := lipgloss.NewStyle().Foreground(iconColor).Render("\uf07b")      // superfile icon.Directory
	paperIcon := lipgloss.NewStyle().Foreground(iconColor).Render("\uf15b")       // superfile icon.File
	browserIcon := lipgloss.NewStyle().Foreground(iconColor).Render("\U000f0208") // superfile icon.Browser

	buttonStyle := lipgloss.NewStyle().
		Foreground(accent).
		Bold(true).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Padding(0, 1)

	commandButton := buttonStyle.Render(terminalIcon + " " + "OPEN COMMANDS — ctrl+p")
	folderButton := buttonStyle.Render(folderIcon + " " + "OPEN FILE FINDER — ctrl+shift+f")
	createButton := buttonStyle.Render(paperIcon + " " + "CREATE FILE — ctrl+n")
	browserBtn := buttonStyle.Render(browserIcon + " " + "WEB BROWSER — ctrl+b")
	buttons := commandButton + "\n\n" + folderButton + "\n\n" + createButton + "\n\n" + browserBtn

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

	// Click rectangles for the stacked home buttons (Command on top,
	// File Finder below, CREATE FILE at the bottom). All start at the
	// left edge; each subsequent rectangle is pushed down past the
	// previous button plus the gap line.
	btnTop := m.layout.main.Min.Y + 3
	cmdH := lipgloss.Height(commandButton)
	folderH := lipgloss.Height(folderButton)
	createH := lipgloss.Height(createButton)
	browserH := lipgloss.Height(browserBtn)
	m.commandButtonRect = image.Rect(
		m.layout.main.Min.X, btnTop,
		m.layout.main.Min.X+lipgloss.Width(commandButton), btnTop+cmdH,
	)
	m.finderButtonRect = image.Rect(
		m.layout.main.Min.X, btnTop+cmdH+1,
		m.layout.main.Min.X+lipgloss.Width(folderButton), btnTop+cmdH+1+folderH,
	)
	m.createFileButtonRect = image.Rect(
		m.layout.main.Min.X, btnTop+cmdH+1+folderH+1,
		m.layout.main.Min.X+lipgloss.Width(createButton), btnTop+cmdH+1+folderH+1+createH,
	)
	m.browserButtonRect = image.Rect(
		m.layout.main.Min.X, btnTop+cmdH+1+folderH+1+createH+1,
		m.layout.main.Min.X+lipgloss.Width(browserBtn), btnTop+cmdH+1+folderH+1+createH+1+browserH,
	)
	parts := []string{cwdStyled, "", buttons, "", modelLine}

	// Idle beaver mascot beneath the status monitors. It shows the normal
	// dense Alpha variant, or the x-ray Beta variant (X_X eyes) when the agent
	// errors. The mascot tracks the cursor/prompt: it faces left when the
	// cursor sits on the left half of the terminal and right when it sits on
	// the right half, and idles in a slow "rest" pose between direction
	// changes instead of cycling on the banner ticker.
	parts = append(parts, "", anim.BeaverFrame(m.beaverFacing, m.beaverErrored, m.beaverResting))
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
		Height(m.layout.main.Dy() - 1).
		PaddingTop(1).
		Render(
			lipgloss.JoinVertical(lipgloss.Left, infoSection, "", content),
		)
}
