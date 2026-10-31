package bots

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/colorprofile"
	"github.com/stretchr/testify/require"
)

func TestSpritesAreSixteenSquareWithEyesOnTheBody(t *testing.T) {
	t.Parallel()
	for _, bot := range Catalog() {
		t.Run(bot.ID, func(t *testing.T) {
			t.Parallel()
			require.Len(t, bot.Sprite.Open, SpriteH)
			require.Len(t, bot.Sprite.Blink, SpriteH)
			openEyes := 0
			blinkEyes := 0
			for y := range SpriteH {
				require.Len(t, bot.Sprite.Open[y], SpriteW)
				require.Len(t, bot.Sprite.Blink[y], SpriteW)
				for x := range SpriteW {
					if bot.Sprite.Open[y][x] == PixEye {
						openEyes++
						require.True(t, isBody(bot.Sprite.Open[y][x]) || surroundingBody(bot, y, x), "eye off body")
					}
					if bot.Sprite.Blink[y][x] == PixEye {
						blinkEyes++
						require.Equal(t, PixEye, bot.Sprite.Open[y][x], "blink eye left the open oval")
					}
				}
			}
			require.GreaterOrEqual(t, openEyes, 12)
			require.GreaterOrEqual(t, blinkEyes, 8)
			require.Less(t, blinkEyes, openEyes)
		})
	}
}

func surroundingBody(bot Bot, y, x int) bool {
	// The eye pixel itself replaces the body, so the open frame no longer
	// holds the body color there. A neighbor in the blink-or-open grid
	// that is still body, or a neighbor that is also an eye, is enough
	// to show the oval sits inside the blob.
	for _, d := range [][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}} {
		yy, xx := y+d[0], x+d[1]
		if yy < 0 || xx < 0 || yy >= SpriteH || xx >= SpriteW {
			continue
		}
		p := bot.Sprite.Open[yy][xx]
		if isBody(p) || p == PixEye || p == PixOrange || p == PixStar {
			return true
		}
	}
	return false
}

func TestEyeContrastOnEveryBody(t *testing.T) {
	t.Parallel()
	eye, _ := PixEye.rgba()
	for _, bot := range Catalog() {
		body, ok := bot.Color.rgba()
		require.True(t, ok, bot.ID)
		// Eyes stay near-black so they read on white Dingus too.
		require.Less(t, int(eye.R)+int(eye.G)+int(eye.B), 80, bot.ID)
		require.Greater(t, dist2(eye, body), 10000, bot.ID)
	}
}

func TestRenderHalfBlocksGolden(t *testing.T) {
	t.Parallel()
	for _, bot := range Catalog() {
		t.Run(bot.ID, func(t *testing.T) {
			t.Parallel()
			got := Render(bot.Sprite, false, ColorNone, 0, false)
			lines := strings.Split(got, "\n")
			require.Len(t, lines, SpriteH/2)
			for _, ln := range lines {
				require.Equal(t, SpriteW, len([]rune(ln)))
			}
			require.Contains(t, got, "█")
			golden(t, bot.ID+"-open", got)

			blink := Render(bot.Sprite, true, ColorNone, 0, false)
			require.NotEqual(t, got, blink)
			golden(t, bot.ID+"-blink", blink)
		})
	}
}

func TestRenderTruecolorAnd256AndNoColor(t *testing.T) {
	t.Parallel()
	bot, ok := ByID("dingus")
	require.True(t, ok)
	trueColor := Render(bot.Sprite, false, ColorTrue, 0, true)
	require.Contains(t, trueColor, "\x1b[38;2;20;20;28m")
	require.Contains(t, trueColor, "\x1b[38;2;252;252;252m")
	require.Contains(t, trueColor, "▀")
	require.NotContains(t, trueColor, "\x1b[38;5;")

	indexed := Render(bot.Sprite, false, ColorANSI256, 0, false)
	require.Contains(t, indexed, "\x1b[38;5;")
	require.NotContains(t, indexed, "\x1b[38;2;")

	plain := Render(bot.Sprite, false, ColorNone, 0, false)
	require.NotContains(t, plain, "\x1b")
}

func TestSelectedTileAndBob(t *testing.T) {
	t.Parallel()
	bot, ok := ByID("brobro")
	require.True(t, ok)
	plain := Render(bot.Sprite, false, ColorTrue, 0, false)
	tiled := Render(bot.Sprite, false, ColorTrue, 0, true)
	require.NotEqual(t, plain, tiled)
	require.Contains(t, tiled, "\x1b[38;2;58;58;70m")

	bobbed := Render(bot.Sprite, false, ColorNone, 1, false)
	require.NotEqual(t, Render(bot.Sprite, false, ColorNone, 0, false), bobbed)
	require.True(t, strings.HasPrefix(bobbed, strings.Repeat(" ", SpriteW)+"\n") || strings.HasPrefix(strings.Split(bobbed, "\n")[0], " "))
}

func TestModeFromProfileRespectsNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	require.Equal(t, ColorNone, ModeFromProfile(colorprofile.TrueColor))
}

func TestAudibleHonorsSilentAndNoSound(t *testing.T) {
	require.True(t, Audible("25"))
	require.True(t, Audible(""))
	require.False(t, Audible("silent"))
	require.False(t, Audible("MUTE"))
	t.Setenv("NO_SOUND", "1")
	require.False(t, Audible("25"))
}

func TestChunkyNameFitsCard(t *testing.T) {
	t.Parallel()
	for _, bot := range Catalog() {
		name := ChunkyName(bot.Name)
		require.LessOrEqual(t, len([]rune(name))*2, SpriteW)
		r := []rune(name)[0]
		require.GreaterOrEqual(t, r, 'Ａ')
		require.LessOrEqual(t, r, 'Ｚ')
	}
}

func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		require.NoError(t, os.MkdirAll("testdata", 0o755))
		require.NoError(t, os.WriteFile(path, []byte(got), 0o644))
	}
	want, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, string(want), got)
}
