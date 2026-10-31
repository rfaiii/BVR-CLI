package dialog

import (
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/richavery/bvr-cli/internal/bots"
	"github.com/richavery/bvr-cli/internal/config"
	"github.com/richavery/bvr-cli/internal/ui/common"
	"github.com/richavery/bvr-cli/internal/ui/styles"
	"github.com/richavery/bvr-cli/internal/workspace"
	"github.com/stretchr/testify/require"
)

func newBots(t *testing.T, width int, volume string) *Bots {
	t.Helper()
	theme := styles.DarkBvrTheme()
	d := NewBots(&common.Common{Styles: &theme}, colorprofile.Ascii, volume, bots.NewFake().Instant())
	d.SetWidth(width)
	return d
}

func TestBotsNavigationKeys(t *testing.T) {
	d := newBots(t, 96, "25")
	require.Equal(t, "dingus", d.SelectedID())

	d.HandleMsg(tea.KeyPressMsg{Code: tea.KeyRight})
	require.Equal(t, "brobro", d.SelectedID())
	d.HandleMsg(keyRune('l'))
	require.Equal(t, "pillo", d.SelectedID())
	d.HandleMsg(keyRune('h'))
	require.Equal(t, "brobro", d.SelectedID())

	d.SetWidth(40)
	before := d.SelectedID()
	d.HandleMsg(keyRune('j'))
	require.NotEqual(t, before, d.SelectedID())
	d.HandleMsg(tea.KeyPressMsg{Code: tea.KeyUp})
	require.Equal(t, before, d.SelectedID())

	d.SetWidth(10)
	d.index = 0
	d.HandleMsg(keyRune('j'))
	require.Equal(t, "brobro", d.SelectedID())

	require.IsType(t, ActionClose{}, d.HandleMsg(tea.KeyPressMsg{Code: tea.KeyEscape}))
}

func TestBotsEnterOpensChatAndEscReturns(t *testing.T) {
	d := newBots(t, 96, "25")
	action := d.HandleMsg(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.True(t, d.InChat())
	cue, ok := action.(ActionPlayCue)
	require.True(t, ok)
	require.Contains(t, cue.Types, "quick-notify-01")
	require.NotEmpty(t, d.lines)
	require.Contains(t, d.lines[0].full, "Dingus")

	require.Nil(t, d.HandleMsg(tea.KeyPressMsg{Code: tea.KeyEscape}))
	require.False(t, d.InChat())
}

func TestGnarfBoopCueIsOnUntilMuted(t *testing.T) {
	d := newBots(t, 96, "25")
	for d.SelectedID() != "gnarf" {
		d.HandleMsg(keyRune('l'))
	}
	action := d.HandleMsg(tea.KeyPressMsg{Code: tea.KeyEnter})
	cue := action.(ActionPlayCue)
	require.Contains(t, cue.Types, "quick-notify-03")

	d.HandleMsg(tea.KeyPressMsg{Code: tea.KeyEscape})
	d.HandleMsg(keyRune('m'))
	require.True(t, d.Muted())
	action = d.HandleMsg(tea.KeyPressMsg{Code: tea.KeyEnter})
	_, isCue := action.(ActionPlayCue)
	require.False(t, isCue)
}

func TestSilentVolumeAndNoSoundDropCues(t *testing.T) {
	d := newBots(t, 96, "silent")
	action := d.HandleMsg(tea.KeyPressMsg{Code: tea.KeyEnter})
	_, isCue := action.(ActionPlayCue)
	require.False(t, isCue)

	t.Setenv("NO_SOUND", "1")
	loud := newBots(t, 96, "25")
	action = loud.HandleMsg(tea.KeyPressMsg{Code: tea.KeyEnter})
	_, isCue = action.(ActionPlayCue)
	require.False(t, isCue)
}

func TestBotsPaletteEntry(t *testing.T) {
	theme := styles.DarkBvrTheme()
	com := &common.Common{
		Styles:    &theme,
		Workspace: configWorkspace{cfg: &config.Config{}},
	}
	commands, err := NewCommands(com, "", false, false, false, nil, nil)
	require.NoError(t, err)
	commands.list.SetFilter("/bots")
	var found *CommandItem
	for _, item := range commands.list.FilteredItems() {
		ci, ok := item.(*CommandItem)
		if ok && ci.id == "bots" {
			found = ci
		}
	}
	require.NotNil(t, found)
	require.Equal(t, "Bots", found.title)
	require.Contains(t, found.aliases, "/bots")
	require.Equal(t, ActionOpenDialog{DialogID: BotsID}, found.Action())
}

func TestPickerASCIICapture(t *testing.T) {
	theme := styles.DarkBvrTheme()
	d := NewBots(&common.Common{Styles: &theme}, colorprofile.Ascii, "25", bots.NewFake())
	d.SetWidth(96)
	screen := uv.NewScreenBuffer(100, 40)
	d.Draw(screen, image.Rect(0, 0, 100, 40))
	plain := ansi.Strip(screen.Render())
	for _, name := range []string{"ＤＩＮＧＵＳ", "ＢＲＯＢＲＯ", "ＰＩＬＬＯ", "ＬＵＶＡＨ", "ＧＮＡＲＦ", "DEMO", "BOTS"} {
		require.Contains(t, plain, name)
	}
	require.Contains(t, plain, "█")
	path := filepath.Join("..", "..", "..", "docs", "media", "bots-picker.txt")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		require.NoError(t, os.WriteFile(path, []byte(plain), 0o644))
	}
	want, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, string(want), plain)
}

type configWorkspace struct {
	workspace.Workspace
	cfg *config.Config
}

func (w configWorkspace) Config() *config.Config { return w.cfg }

func keyRune(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

func TestTypewriterUsesBotPace(t *testing.T) {
	d := newBots(t, 96, "silent")
	d.HandleMsg(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.Equal(t, 0, d.lines[0].shown)
	d.HandleMsg(writerTickMsg{})
	require.Equal(t, 1, d.lines[0].shown)

	fast := newBots(t, 96, "silent")
	fast.HandleMsg(keyRune('l'))
	fast.HandleMsg(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.Equal(t, "brobro", fast.SelectedID())
	fast.HandleMsg(writerTickMsg{})
	require.GreaterOrEqual(t, fast.lines[0].shown, 3)
}

func TestSubmitShowsEmote(t *testing.T) {
	d := newBots(t, 96, "silent")
	d.HandleMsg(tea.KeyPressMsg{Code: tea.KeyEnter})
	d.input.SetValue("hello")
	action := d.HandleMsg(tea.KeyPressMsg{Code: tea.KeyEnter})
	cmd := action.(ActionCmd).Cmd
	require.NotNil(t, cmd)
	msg := cmd().(streamOpenMsg)
	require.NoError(t, msg.err)
	d.HandleMsg(msg)
	var sawEmote bool
	for {
		chunk := waitChunk(msg.gen, d.stream)().(streamChunkMsg)
		d.HandleMsg(chunk)
		if chunk.done {
			break
		}
		if chunk.emote != "" {
			sawEmote = true
		}
	}
	var botText string
	for _, ln := range d.lines {
		if ln.fromBot && strings.Contains(ln.full, "hello") {
			botText = ln.full
			require.NotEmpty(t, ln.emote)
		}
	}
	require.NotEmpty(t, botText)
	require.True(t, sawEmote)
}
