package dialog

import (
	"fmt"
	"image"
	"net/http"
	"net/url"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	codebergreadability "codeberg.org/readeck/go-readability/v2"
	"github.com/JohannesKaufmann/html-to-markdown/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/richavery/bvr-cli/internal/ui/common"
)

// BrowserID identifies the web browser dialog.
const BrowserID = "browser"

// Browser is a terminal-based article reader that fetches a URL,
// extracts the main content, converts it to Markdown, and renders
// it as styled ANSI text in a scrollable viewport.
type Browser struct {
	com        *common.Common
	url        string
	title      string
	content    string
	loading    bool
	err        error
	help       help.Model
	input      textinput.Model
	spinner    spinner.Model
	viewport   viewport.Model
	keyMap     struct{ Close, Enter, Back, Forward, Home key.Binding }
	closeRect  image.Rectangle
	panelRect  image.Rectangle
	history    []string
	historyIdx int
}

var _ Dialog = (*Browser)(nil)

// NewBrowser creates a new browser dialog pre-filled with the given URL.
func NewBrowser(com *common.Common, url string) (*Browser, tea.Cmd) {
	b := &Browser{com: com, url: url}
	b.help = help.New()
	b.help.Styles = com.Styles.DialogHelpStyles()
	b.input = textinput.New()
	b.input.SetVirtualCursor(false)
	b.input.SetStyles(com.Styles.TextInput)
	b.input.Placeholder = "Enter URL..."
	b.input.SetValue(url)
	b.input.Focus()
	b.spinner = spinner.New()
	b.spinner.Spinner = spinner.Dot
	b.spinner.Style = com.Styles.Dialog.Spinner
	b.keyMap.Close = CloseKey
	b.keyMap.Close.SetHelp("esc", "close")
	b.keyMap.Enter = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "fetch"))
	b.keyMap.Back = key.NewBinding(key.WithKeys("alt+left", "ctrl+b"), key.WithHelp("alt+←/ctrl+b", "back"))
	b.keyMap.Forward = key.NewBinding(key.WithKeys("alt+right", "ctrl+f"), key.WithHelp("alt+→/ctrl+f", "forward"))
	b.keyMap.Home = key.NewBinding(key.WithKeys("ctrl+h"), key.WithHelp("ctrl+h", "home"))
	b.history = []string{}
	b.historyIdx = -1
	return b, nil
}

// ID implements Dialog.
func (b *Browser) ID() string { return BrowserID }

// HandleMsg implements Dialog.
func (b *Browser) HandleMsg(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, b.keyMap.Close) {
			return ActionClose{}
		}
		if b.loading {
			return nil
		}
		if key.Matches(msg, b.keyMap.Home) {
			b.url = ""
			b.input.SetValue("")
			b.content = ""
			b.title = ""
			b.err = nil
			b.history = []string{}
			b.historyIdx = -1
			b.input.Focus()
			return nil
		}
		if key.Matches(msg, b.keyMap.Back) {
			if b.historyIdx > 0 {
				b.historyIdx--
				b.url = b.history[b.historyIdx]
				b.input.SetValue(b.url)
				b.loading = true
				b.input.Blur()
				b.content = ""
				b.err = nil
				b.title = ""
				return ActionCmd{Cmd: b.fetchArticle(b.url)}
			}
			return nil
		}
		if key.Matches(msg, b.keyMap.Forward) {
			if b.historyIdx < len(b.history)-1 {
				b.historyIdx++
				b.url = b.history[b.historyIdx]
				b.input.SetValue(b.url)
				b.loading = true
				b.input.Blur()
				b.content = ""
				b.err = nil
				b.title = ""
				return ActionCmd{Cmd: b.fetchArticle(b.url)}
			}
			return nil
		}

		var cmd tea.Cmd
		// Only update input if it is focused. If blurred, allow viewport to scroll.
		if b.input.Focused() {
			b.input, cmd = b.input.Update(msg)
		}
		if b.content != "" && !b.input.Focused() {
			b.viewport, cmd = b.viewport.Update(msg)
			// allow re-focusing input if needed (e.g. they hit tab)
			if msg.String() == "tab" {
				b.input.Focus()
				return nil
			}
		}

		if msg.String() == "enter" && b.input.Focused() {
			newUrl := b.input.Value()
			if newUrl == "" {
				return nil
			}
			b.url = newUrl
			if b.historyIdx >= 0 {
				b.history = b.history[:b.historyIdx+1]
			}
			b.history = append(b.history, b.url)
			b.historyIdx = len(b.history) - 1

			b.loading = true
			b.input.Blur()
			b.content = ""
			b.err = nil
			b.title = ""
			return ActionCmd{Cmd: b.fetchArticle(b.url)}
		}
		if cmd != nil {
			return ActionCmd{Cmd: cmd}
		}
	case spinner.TickMsg:
		if b.loading {
			var cmd tea.Cmd
			b.spinner, cmd = b.spinner.Update(msg)
			if cmd != nil {
				return ActionCmd{Cmd: cmd}
			}
		}
	case articleRenderedMsg:
		b.loading = false
		b.input.Focus()
		if msg.err != nil {
			b.err = msg.err
			b.content = ""
		} else {
			b.content = msg.content
		}
		return nil
	case tea.MouseClickMsg:
		if msg.Button == uv.MouseLeft && image.Pt(msg.X, msg.Y).In(b.closeRect) {
			return ActionClose{}
		}
		// Click-away to dismiss: a left click outside the panel closes the browser.
		if msg.Button == uv.MouseLeft && !b.panelRect.Empty() && !image.Pt(msg.X, msg.Y).In(b.panelRect) {
			return ActionClose{}
		}
	}
	return nil
}

// Draw implements Dialog.
func (b *Browser) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	t := b.com.Styles
	dialogStyle := t.Dialog.View
	dw := min(area.Dx(), 120)
	dh := max(area.Dy()-2, 10)
	titleText := "Web Browser"
	if b.title != "" {
		titleText = b.title
	}
	title := common.DialogTitle(t, titleText, dw, t.Dialog.TitleGradFromColor, t.Dialog.TitleGradToColor)
	inputArea := lipgloss.JoinHorizontal(lipgloss.Top, t.Dialog.InputPrompt.Render("URL "), b.input.View())
	var contentStr string
	if b.loading {
		contentStr = b.spinner.View() + " Loading article..."
	} else if b.err != nil {
		contentStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5caa")).Render("Error: " + b.err.Error())
	} else if b.content != "" {
		contentArea := dw - 2
		b.viewport.SetWidth(contentArea)
		b.viewport.SetHeight(max(1, dh-6))
		b.viewport.SetContent(b.content)
		contentStr = b.viewport.View()
	} else {
		contentStr = "Enter a URL and press Enter to load an article."
	}
	helpView := renderDialogHelp(t, &b.help, b, dw)
	view := lipgloss.JoinVertical(lipgloss.Left, t.Dialog.Title.Render(title), dialogStyle.Render(inputArea), contentStr, helpView)
	dialog := dialogStyle.Width(dw).Render(view)
	center := common.CenterRect(area, lipgloss.Width(dialog), lipgloss.Height(dialog))
	uv.NewStyledString(dialog).Draw(scr, center)

	// Close button [X] at top-right of the dialog title area.
	closeText := closeLabel(dw)
	closeW := ansi.StringWidth(closeText)
	closeX := center.Min.X + dw - closeW - 1
	b.closeRect = image.Rect(closeX, center.Min.Y, closeX+closeW, center.Min.Y+1)
	b.panelRect = center
	closeStyle := t.Dialog.FileBrowser.Close
	uv.NewStyledString(closeStyle.Render(closeText)).Draw(scr, b.closeRect)
	return &tea.Cursor{Position: tea.Position{X: center.Min.X + 1, Y: center.Min.Y}}
}

// StartLoading implements LoadingDialog.
func (b *Browser) StartLoading() tea.Cmd {
	if b.loading {
		return nil
	}
	b.loading = true
	return b.spinner.Tick
}

// StopLoading implements LoadingDialog.
func (b *Browser) StopLoading() { b.loading = false }

// ShortHelp implements help.KeyMap.
func (b *Browser) ShortHelp() []key.Binding {
	return []key.Binding{b.keyMap.Enter, b.keyMap.Close}
}

// FullHelp implements help.KeyMap.
func (b *Browser) FullHelp() [][]key.Binding {
	return [][]key.Binding{{b.keyMap.Enter}, {b.keyMap.Close}}
}

// closeLabel returns the close button text.
func (b *Browser) fetchArticle(pageURL string) tea.Cmd {
	return func() tea.Msg {
		req, err := http.NewRequest("GET", pageURL, nil)
		if err != nil {
			return articleRenderedMsg{err: err}
		}
		req.Header.Set("User-Agent", "BVR-CLI/1.0")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return articleRenderedMsg{err: err}
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return articleRenderedMsg{err: fmt.Errorf("HTTP %d", resp.StatusCode)}
		}
		parsedURL, err := url.Parse(pageURL)
		if err != nil {
			return articleRenderedMsg{err: err}
		}
		article, err := codebergreadability.FromReader(resp.Body, parsedURL)
		if err != nil {
			return articleRenderedMsg{err: err}
		}
		b.title = article.Title()
		var htmlBuilder strings.Builder
		if err := article.RenderHTML(&htmlBuilder); err != nil {
			return articleRenderedMsg{err: err}
		}
		htmlContent := htmlBuilder.String()
		markdown, err := htmltomarkdown.ConvertString(htmlContent)
		if err != nil {
			return articleRenderedMsg{err: err}
		}
		ansi, err := renderMarkdownANSI(markdown)
		if err != nil {
			return articleRenderedMsg{err: err}
		}
		return articleRenderedMsg{content: ansi}
	}
}

// renderMarkdownANSI renders Markdown to ANSI-styled terminal text.
func renderMarkdownANSI(markdown string) (string, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(80),
		glamour.WithStylesFromJSONBytes([]byte(`{"document":{"margin":1},"heading":{"color":"#b972ff","bold":true},"link":{"color":"#39f66b","underline":true}}`)),
	)
	if err != nil {
		return "", err
	}
	return renderer.Render(markdown)
}

// articleRenderedMsg is a message sent when the article has been fetched and rendered.
type articleRenderedMsg struct {
	content string
	err     error
}

// SetURL sets the URL for the browser and updates the input field.
func (b *Browser) SetURL(url string) {
	b.url = url
	b.input.SetValue(url)
}
