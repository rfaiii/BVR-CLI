package dialog

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/richavery/bvr-cli/internal/clipboard"
	"github.com/richavery/bvr-cli/internal/ui/common"
)

// FileBrowserID identifies the project file browser overlay.
const FileBrowserID = "file-browser"

type fileBrowserEntry struct {
	name string
	path string
	dir  bool
}

// FileBrowser is a focused, in-app file finder with preview, metadata, and
// clipboard panes. It stays intentionally compact so it remains useful beside
// the agent conversation.
type FileBrowser struct {
	com        *common.Common
	dir        string
	entries    []fileBrowserEntry
	selected   int
	scroll     int
	showHidden bool
	preview    string
	metadata   string
	clipboard  string
	loading    bool
	loadError  string
	loadSeq    uint64

	up, down, pageUp, pageDown, first, last, toggleHidden, refresh, open, back, copy, attach, external, changeProject, browser, create, close key.Binding
	contentRect, closeRect, panelRect                                                                                                         image.Rectangle
}

var _ Dialog = (*FileBrowser)(nil)

func NewFileBrowser(com *common.Common) (*FileBrowser, tea.Cmd) {
	dir := com.Workspace.WorkingDir()
	if dir == "" {
		dir, _ = os.Getwd()
	}
	f := &FileBrowser{com: com, dir: dir}
	f.up = key.NewBinding(key.WithKeys("up", "k"))
	f.down = key.NewBinding(key.WithKeys("down", "j"))
	f.pageUp = key.NewBinding(key.WithKeys("pgup", "ctrl+u"))
	f.pageDown = key.NewBinding(key.WithKeys("pgdown", "ctrl+d"))
	f.first = key.NewBinding(key.WithKeys("home", "g"))
	f.last = key.NewBinding(key.WithKeys("end", "G"))
	f.toggleHidden = key.NewBinding(key.WithKeys("."), key.WithHelp(".", "show hidden"))
	f.refresh = key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh"))
	f.open = key.NewBinding(key.WithKeys("enter", "right", "l"))
	f.back = key.NewBinding(key.WithKeys("left", "h", "backspace"))
	f.copy = key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "copy path"))
	f.attach = key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "attach"))
	f.external = key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open editor"))
	f.changeProject = key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "switch project"))
	f.browser = key.NewBinding(key.WithKeys("b"), key.WithHelp("b", "web browser"))
	f.create = key.NewBinding(key.WithKeys("ctrl+n"), key.WithHelp("ctrl+n", "create file"))
	f.close = CloseKey
	if b, err := clipboard.Read(clipboard.FormatText); err == nil {
		f.clipboard = strings.TrimSpace(string(b))
	}
	return f, f.requestReload()
}

func (f *FileBrowser) ID() string { return FileBrowserID }

type fileBrowserLoadMsg struct {
	seq      uint64
	dir      string
	entries  []fileBrowserEntry
	preview  string
	metadata string
	err      error
}

type fileBrowserPreviewMsg struct {
	seq      uint64
	path     string
	preview  string
	metadata string
	err      error
}

func (f *FileBrowser) requestReload() tea.Cmd {
	f.loadSeq++
	seq := f.loadSeq
	f.loading = true
	f.loadError = ""
	dir, showHidden, selected := f.dir, f.showHidden, f.selected
	return func() tea.Msg {
		items, err := os.ReadDir(dir)
		if err != nil {
			return fileBrowserLoadMsg{seq: seq, dir: dir, err: err}
		}
		entries := make([]fileBrowserEntry, 0, len(items))
		for _, item := range items {
			if !showHidden && strings.HasPrefix(item.Name(), ".") {
				continue
			}
			entries = append(entries, fileBrowserEntry{name: item.Name(), path: filepath.Join(dir, item.Name()), dir: item.IsDir()})
		}
		selected = min(max(0, selected), max(0, len(entries)-1))
		preview, metadata := "(empty directory)", "Metadata: (none selected)"
		if len(entries) > 0 {
			preview = previewForEntry(entries[selected])
			metadata = metadataForEntry(entries[selected])
		}
		return fileBrowserLoadMsg{seq: seq, dir: dir, entries: entries, preview: preview, metadata: metadata}
	}
}

func (f *FileBrowser) visibleEntryRows() int {
	// One body row is reserved for the current directory path.
	return max(0, f.bodyHeight()-1)
}

func (f *FileBrowser) bodyHeight() int {
	// Draw calculates this from the terminal. Keep the current viewport state
	// valid before the first draw and use a generous fallback for navigation.
	if f.contentRect.Dy() > 0 {
		return f.contentRect.Dy()
	}
	return 20
}

func (f *FileBrowser) ensureSelectedVisible(viewport int) {
	if viewport <= 0 {
		f.scroll = 0
		return
	}
	if f.selected < f.scroll {
		f.scroll = f.selected
	}
	if f.selected >= f.scroll+viewport {
		f.scroll = f.selected - viewport + 1
	}
	f.scroll = min(max(0, f.scroll), max(0, len(f.entries)-viewport))
}

func previewForEntry(e fileBrowserEntry) string {
	if e.dir {
		return "directory\n\nPress enter to open"
	}
	b, err := os.ReadFile(e.path)
	if err != nil {
		return err.Error()
	}
	if len(b) > 12000 {
		b = b[:12000]
	}
	if strings.IndexByte(string(b), 0) >= 0 {
		return "(binary file)"
	}
	return string(b)
}

func metadataForEntry(e fileBrowserEntry) string {
	info, err := os.Stat(e.path)
	if err != nil {
		return "Metadata: " + err.Error()
	}
	return fmt.Sprintf("Metadata: %s  •  %d bytes  •  %s", e.name, info.Size(), info.ModTime().Format("2006-01-02 15:04"))
}

func (f *FileBrowser) requestPreview() tea.Cmd {
	if len(f.entries) == 0 || f.selected < 0 || f.selected >= len(f.entries) {
		f.preview = "(empty directory)"
		return nil
	}
	f.loadSeq++
	seq, entry := f.loadSeq, f.entries[f.selected]
	f.loading = true
	f.loadError = ""
	return func() tea.Msg {
		return fileBrowserPreviewMsg{seq: seq, path: entry.path, preview: previewForEntry(entry), metadata: metadataForEntry(entry)}
	}
}

func fileBrowserAction(cmd tea.Cmd) Action {
	if cmd == nil {
		return nil
	}
	return ActionCmd{Cmd: cmd}
}

func (f *FileBrowser) HandleMsg(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case fileBrowserLoadMsg:
		if msg.seq != f.loadSeq || msg.dir != f.dir {
			return nil
		}
		f.loading = false
		f.entries, f.preview, f.metadata, f.loadError = msg.entries, msg.preview, msg.metadata, ""
		if msg.err != nil {
			f.loadError, f.preview, f.metadata = msg.err.Error(), msg.err.Error(), "Metadata: unavailable"
		}
		f.selected = min(max(0, f.selected), max(0, len(f.entries)-1))
		f.ensureSelectedVisible(f.visibleEntryRows())
		return nil
	case fileBrowserPreviewMsg:
		if msg.seq != f.loadSeq || len(f.entries) == 0 || f.entries[f.selected].path != msg.path {
			return nil
		}
		f.loading = false
		f.preview, f.metadata, f.loadError = msg.preview, msg.metadata, ""
		return nil
	}
	if mouse, ok := msg.(tea.MouseClickMsg); ok {
		if mouse.Button == uv.MouseLeft && image.Pt(mouse.X, mouse.Y).In(f.closeRect) {
			return ActionClose{}
		}
		// Click-away to dismiss: a left click outside the panel closes the finder.
		if mouse.Button == uv.MouseLeft && !f.panelRect.Empty() && !image.Pt(mouse.X, mouse.Y).In(f.panelRect) {
			return ActionClose{}
		}
		if !image.Pt(mouse.X, mouse.Y).In(f.contentRect) || len(f.entries) == 0 {
			return nil
		}
		row := mouse.Y - f.contentRect.Min.Y
		// The first body row is the current directory; entries begin below it.
		visible := f.visibleEntryRows()
		if row <= 0 || row > visible {
			return nil
		}
		index := f.scroll + row - 1
		if index < 0 || index >= len(f.entries) {
			return nil
		}
		f.selected = index
		cmd := f.requestPreview()
		if f.entries[f.selected].dir {
			f.dir = f.entries[f.selected].path
			f.selected = 0
			cmd = f.requestReload()
		}
		return fileBrowserAction(cmd)
	}
	if wheel, ok := msg.(tea.MouseWheelMsg); ok {
		if !image.Pt(wheel.X, wheel.Y).In(f.contentRect) || len(f.entries) == 0 {
			return nil
		}
		switch wheel.Button {
		case uv.MouseWheelUp:
			f.selected = max(0, f.selected-1)
		case uv.MouseWheelDown:
			f.selected = min(len(f.entries)-1, f.selected+1)
		default:
			return nil
		}
		f.ensureSelectedVisible(f.visibleEntryRows())
		return fileBrowserAction(f.requestPreview())
	}
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(k, f.close):
			return ActionClose{}
		case key.Matches(k, f.up):
			if f.selected > 0 {
				f.selected--
				f.ensureSelectedVisible(f.visibleEntryRows())
				if action := fileBrowserAction(f.requestPreview()); action != nil {
					return action
				}
			}
		case key.Matches(k, f.down):
			if f.selected+1 < len(f.entries) {
				f.selected++
				f.ensureSelectedVisible(f.visibleEntryRows())
				if action := fileBrowserAction(f.requestPreview()); action != nil {
					return action
				}
			}
		case key.Matches(k, f.pageUp):
			f.selected = max(0, f.selected-f.visibleEntryRows())
			f.ensureSelectedVisible(f.visibleEntryRows())
			if action := fileBrowserAction(f.requestPreview()); action != nil {
				return action
			}
		case key.Matches(k, f.pageDown):
			f.selected = min(max(0, len(f.entries)-1), f.selected+f.visibleEntryRows())
			f.ensureSelectedVisible(f.visibleEntryRows())
			if action := fileBrowserAction(f.requestPreview()); action != nil {
				return action
			}
		case key.Matches(k, f.first):
			f.selected, f.scroll = 0, 0
			if action := fileBrowserAction(f.requestPreview()); action != nil {
				return action
			}
		case key.Matches(k, f.last):
			f.selected = max(0, len(f.entries)-1)
			f.ensureSelectedVisible(f.visibleEntryRows())
			if action := fileBrowserAction(f.requestPreview()); action != nil {
				return action
			}
		case key.Matches(k, f.toggleHidden):
			f.showHidden = !f.showHidden
			f.selected, f.scroll = 0, 0
			return fileBrowserAction(f.requestReload())
		case key.Matches(k, f.refresh):
			return fileBrowserAction(f.requestReload())
		case key.Matches(k, f.copy):
			if len(f.entries) > 0 {
				clipboard.WriteText(f.entries[f.selected].path)
				f.clipboard = f.entries[f.selected].path
			}
		case key.Matches(k, f.attach):
			if len(f.entries) > 0 && !f.entries[f.selected].dir {
				return ActionFileBrowserSelected{Path: f.entries[f.selected].path}
			}
		case key.Matches(k, f.external):
			if len(f.entries) > 0 && !f.entries[f.selected].dir {
				return ActionFileBrowserOpenExternal{Path: f.entries[f.selected].path}
			}
		case key.Matches(k, f.changeProject):
			if len(f.entries) > 0 && f.entries[f.selected].dir {
				return ActionChangeProject{Path: f.entries[f.selected].path}
			}
		case key.Matches(k, f.browser):
			return ActionOpenDialog{DialogID: BrowserID}
		case key.Matches(k, f.create):
			return ActionOpenDialog{DialogID: CreateFileID}
		case key.Matches(k, f.back):
			parent := filepath.Dir(f.dir)
			if parent != f.dir {
				f.dir = parent
				f.selected = 0
				f.scroll = 0
				return fileBrowserAction(f.requestReload())
			}
		case key.Matches(k, f.open):
			if len(f.entries) > 0 && f.entries[f.selected].dir {
				f.dir = f.entries[f.selected].path
				f.selected = 0
				f.scroll = 0
				return fileBrowserAction(f.requestReload())
			}
		}
	}
	return nil
}

func (f *FileBrowser) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	// Calculate the outer rectangle first. Lip Gloss Width/Height apply to the
	// content box, so subtract the border and padding before rendering. This
	// keeps the final panel inside the terminal regardless of file contents.
	panelTheme := f.com.Styles.Dialog.FileBrowser
	panelStyle := panelTheme.Panel.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(panelTheme.Border.GetForeground()).
		Padding(1)
	// Use the available vertical space while retaining a small terminal inset.
	// This intentionally covers the large application banner while the Finder is
	// open, giving both panes more room without changing the global header.
	width := min(max(1, area.Dx()-2), 120)
	height := max(1, area.Dy()-2)
	contentW := max(1, width-panelStyle.GetHorizontalFrameSize())
	contentH := max(1, height-panelStyle.GetVerticalFrameSize())
	center := common.CenterRect(area, width, height)
	f.panelRect = center
	contentX := center.Min.X + 1 + 1
	contentY := center.Min.Y + 1 + 1

	// Reserve fixed rows for the title, pane header rule, divider, and footer. The
	// remaining body is a fixed-height, two-pane viewport; long paths and previews
	// are clipped.
	titleH := min(1, contentH)
	paneHeaderH := min(1, max(0, contentH-titleH))
	dividerH := min(1, max(0, contentH-titleH-paneHeaderH))
	footerH := min(3, max(0, contentH-titleH-paneHeaderH-dividerH))
	bodyH := max(0, contentH-titleH-paneHeaderH-dividerH-footerH)
	leftW, rightW, listW := finderPaneWidths(contentW)
	f.contentRect = image.Rect(contentX, contentY+titleH+paneHeaderH, contentX+leftW, contentY+titleH+paneHeaderH+bodyH)
	f.ensureSelectedVisible(max(0, bodyH-1))
	// Include the title row and a few cells around the label so the control is
	// easy to hit with a mouse even when the terminal font is narrow.
	closeText := closeLabel(contentW)
	closeW := ansi.StringWidth(closeText)
	// Keep the control one cell inside the content edge so it reads as part of
	// the panel rather than an overlay hanging off its border.
	closeX := contentX + contentW - closeW - 1
	f.closeRect = image.Rect(closeX, contentY, closeX+closeW, contentY+titleH)

	leftLines := []string{"📁 " + f.dir}
	start := min(max(0, f.scroll), len(f.entries))
	end := min(len(f.entries), start+max(0, bodyH-1))
	for _, e := range f.entries[start:end] {
		icon := "▸"
		if !e.dir {
			icon = "·"
		}
		leftLines = append(leftLines, fmt.Sprintf("%s %s", icon, e.name))
	}
	visibleEntries := max(0, bodyH-1)
	// Always reserve one track column, even when the list fits. This keeps the
	// pane geometry invariant as directories change size.
	leftLines = fixedLines(leftLines, listW, bodyH)
	leftRendered := make([]string, len(leftLines))
	for i, line := range leftLines {
		// fixedLines has already made this row exactly listW cells wide. Do not
		// apply a second Lip Gloss width here: it can wrap long styled rows and
		// break the scrollbar into disconnected pieces.
		style := panelTheme.Entry
		if i > 0 && start+i-1 == f.selected {
			style = panelTheme.Selected
		} else if i == 0 {
			style = panelTheme.Directory
		}
		leftRendered[i] = style.Render(line)
	}

	preview := f.preview
	if preview == "" {
		preview = "(select a file to preview)"
	}
	rightLines := fixedLines(strings.Split(preview, "\n"), rightW, bodyH)
	rightRendered := make([]string, len(rightLines))
	for i, line := range rightLines {
		rightRendered[i] = panelTheme.Preview.Render(line)
	}
	bodyLines := make([]string, bodyH)
	for i := range bodyLines {
		// The panes are drawn into their own bounded rectangles below. Keeping
		// the panel's body canvas blank prevents Lip Gloss from reflowing a
		// composed ANSI row (notably RTF/control-heavy text) across the divider.
		bodyLines[i] = strings.Repeat(" ", contentW)
	}

	meta := f.metadata
	if meta == "" {
		meta = "Metadata: (loading)"
	}
	if f.loading {
		meta += "  •  loading…"
	}
	clip := "Clipboard: (empty)"
	if f.clipboard != "" {
		clip = "Clipboard: " + f.clipboard
	}
	footerLines := fixedLines([]string{meta, clip, "↑↓ navigate  a attach  o editor  s switch project  r refresh  enter open  y copy  esc close  b web browser  ctrl+n create file"}, contentW, footerH)
	titleW := max(1, contentW-closeW-1)
	titleText := padRight(ansi.Truncate("BVR FILE FINDER", titleW, "…"), titleW) + strings.Repeat(" ", contentW-titleW)
	titleCell := panelTheme.Title.Render(titleText)
	contentLines := make([]string, 0, contentH)
	contentLines = append(contentLines, titleCell)
	if paneHeaderH > 0 {
		paneW := leftW + 1 + rightW
		contentLines = append(contentLines, panelTheme.Rule.Render(strings.Repeat("─", paneW)+strings.Repeat(" ", max(0, contentW-paneW))))
	}
	contentLines = append(contentLines, bodyLines...)
	if dividerH > 0 {
		contentLines = append(contentLines, panelTheme.Rule.Render(strings.Repeat("─", contentW)))
	}
	for i, line := range footerLines {
		// The footer's first lines (metadata, clipboard) use the theme ALT
		// COLOR; the trailing help line keeps the muted footer style.
		if i < 2 {
			contentLines = append(contentLines, panelTheme.Metadata.Render(line))
		} else {
			contentLines = append(contentLines, panelTheme.Footer.Render(line))
		}
	}
	// Every row is fixed before styling; apply the known content dimensions once
	// so the outer panel cannot grow when preview text changes.
	view := panelStyle.Width(contentW).Height(contentH).Render(strings.Join(contentLines, "\n"))
	panel := common.CenterRect(area, width, height)
	actualContentX := panel.Min.X + panelStyle.GetHorizontalFrameSize()/2
	actualContentY := panel.Min.Y + panelStyle.GetVerticalFrameSize()/2
	f.contentRect = image.Rect(actualContentX, actualContentY+titleH+paneHeaderH, actualContentX+leftW, actualContentY+titleH+paneHeaderH+bodyH)
	closeX = actualContentX + contentW - closeW - 1
	f.closeRect = image.Rect(closeX, actualContentY, closeX+closeW, actualContentY+titleH)
	uv.NewStyledString(view).Draw(scr, panel)

	// Draw each pane independently with wrapping disabled and a hard rectangle
	// boundary. A long or control-heavy preview can therefore only be clipped
	// inside the preview pane; it cannot spill into the file list.
	for i := range leftRendered {
		y := actualContentY + titleH + paneHeaderH + i
		uv.NewStyledString(leftRendered[i]).Draw(scr, image.Rect(actualContentX, y, actualContentX+listW, y+1))
		uv.NewStyledString(panelTheme.Rule.Render("│")).Draw(scr, image.Rect(actualContentX+leftW, y, actualContentX+leftW+1, y+1))
		uv.NewStyledString(rightRendered[i]).Draw(scr, image.Rect(actualContentX+leftW+1, y, actualContentX+leftW+1+rightW, y+1))
	}

	// Draw controls after the fixed panel so text layout cannot wrap or move
	// them. These rectangles are the same ones used by HandleMsg.
	lipglossClose := panelTheme.Close.Render(closeText)
	uv.NewStyledString(lipglossClose).Draw(scr, f.closeRect)
	if visibleEntries > 0 {
		bar := scrollbarLines(visibleEntries, len(f.entries), f.scroll)
		barX := actualContentX + listW
		for i, glyph := range bar {
			uv.NewStyledString(panelTheme.Close.Render(glyph)).Draw(scr, image.Rect(barX, actualContentY+titleH+paneHeaderH+i, barX+1, actualContentY+titleH+paneHeaderH+i+1))
		}
	}
	return nil
}

// finderPaneWidths leaves a deliberate three-cell gutter before the panel's
// right frame. The extra space is important for terminals whose styled-string
// renderer handles control-heavy input differently at a hard rectangle edge.
func finderPaneWidths(contentW int) (leftW, rightW, listW int) {
	const separatorW = 1
	rightMargin := min(3, max(0, contentW-3))
	innerW := max(1, contentW-rightMargin)
	leftW = max(1, (innerW-separatorW)*2/5)
	rightW = max(1, innerW-leftW-separatorW)
	listW = max(1, leftW-1)
	return leftW, rightW, listW
}

func closeLabel(contentW int) string {
	if contentW < 10 {
		return "X"
	}
	return "[X]"
}

func scrollbarLines(viewport, total, offset int) []string {
	if viewport <= 0 {
		return nil
	}
	if total <= viewport {
		return slices.Repeat([]string{"█"}, viewport)
	}
	thumb := max(1, viewport*viewport/total)
	start := (viewport - thumb) * offset / max(1, total-viewport)
	lines := make([]string, viewport)
	for i := range lines {
		lines[i] = "░"
		if i >= start && i < start+thumb {
			lines[i] = "█"
		}
	}
	return lines
}

// fixedLines turns arbitrary text into an exact width/height viewport. It
// deliberately never wraps: wrapping is what previously let long paths,
// metadata, and file previews change the dialog's final size.
func fixedLines(lines []string, width, height int) []string {
	if width <= 0 || height <= 0 {
		return nil
	}
	width, height = max(1, width), max(1, height)
	result := make([]string, 0, height)
	for _, line := range lines {
		if len(result) == height {
			break
		}
		// Normalize control characters before clipping. In particular, a carriage
		// return in a preview must not move the terminal cursor back over the
		// beginning of a row and make an apparently clipped line overflow.
		line = strings.NewReplacer("\t", " ", "\r", "").Replace(line)
		line = ansi.Truncate(line, width, "…")
		result = append(result, padRight(line, width))
	}
	for len(result) < height {
		result = append(result, strings.Repeat(" ", width))
	}
	return result
}

func padRight(value string, width int) string {
	return value + strings.Repeat(" ", max(0, width-ansi.StringWidth(value)))
}
