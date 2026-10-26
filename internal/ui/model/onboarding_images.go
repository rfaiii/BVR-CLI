package model

import (
	"bytes"
	"embed"
	"image"
	"image/png"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/richavery/bvr-cli/internal/ui/common"
	fimage "github.com/richavery/bvr-cli/internal/ui/image"
	"github.com/richavery/bvr-cli/internal/ui/util"
)

// tourShots carries the first-run screenshots into the binary. They are
// embedded because resources/screenshots/ ships with the repository, not with
// an installed BVR, and the working directory at launch is whatever project the
// user ran the tool from — every "Screenshot: resources/…" line the tour used
// to print pointed at a file that was never there.
//
// Regenerate with scripts/prepare_onboarding_shots.sh.
//
//go:embed shots/*.png
var tourShots embed.FS

// tourShot is one step's picture and the line that labels it.
type tourShot struct {
	id      string // file stem under shots/, and the image cache key
	caption string
}

// shotForStep maps a step to its picture. Notifications has none: no capture in
// resources/screenshots shows the notification area, so the step stays text
// rather than dressing itself in someone else's screenshot.
func shotForStep(step onboardingStep) (tourShot, bool) {
	switch step {
	case onboardingStepHome:
		return tourShot{"home", "Home: project, launch buttons, model, skills, and NODE link."}, true
	case onboardingStepModels:
		return tourShot{"models", "The models catalog, opened with ctrl+m or /models."}, true
	case onboardingStepFileFinder:
		return tourShot{"commands", "File Finder opens from the command menu (ctrl+p, then ctrl+o)."}, true
	case onboardingStepThemes:
		return tourShot{"themes", "The same screen under the pink theme."}, true
	}
	return tourShot{}, false
}

// The picture is sized in cells because that is the grid it is painted onto.
// Margins below are what the step's own text and the dialog padding keep for
// themselves.
const (
	shotMaxCols   = 96
	shotMinCols   = 24
	shotMinRows   = 4
	shotColMargin = 6
	// shotRowBudget is everything the picture does not get: the banner header,
	// the step's heading and lines, the caption, the two prompt lines, and a
	// row of slack. Below what is left the picture is dropped rather than
	// squeezed into a strip that pushes the prompts off screen.
	shotRowBudget = 24
)

// shotLayout is everything a transmit needs, captured on the update goroutine
// so the command never reads shared state while Update owns it.
type shotLayout struct {
	enc    fimage.Encoding
	cell   fimage.CellSize
	cols   int
	rows   int
	isTmux bool
}

// SetImageCapabilities tells the tour how far it can go with pixels. The
// pointer is kept rather than copied so a terminal that answers the graphics
// query after onboarding starts is still honoured on the next step.
func (o *onboardingModel) SetImageCapabilities(caps *common.Capabilities) {
	o.caps = caps
}

// shotLayoutFor picks an encoding and a cell box for the step's picture, or
// false when this terminal has no way to show one.
func (o *onboardingModel) shotLayoutFor(id string) (shotLayout, bool) {
	caps := o.caps
	if caps == nil {
		return shotLayout{}, false
	}
	kitty := caps.SupportsKittyGraphics()
	if !kitty && caps.Profile < colorprofile.ANSI256 {
		// The half-block painter needs enough palette room to approximate the
		// picture; below 256 colors a screenshot is noise and the text wins.
		return shotLayout{}, false
	}
	width, height, err := shotPixelSize(id)
	if err != nil {
		return shotLayout{}, false
	}
	cellW, cellH := caps.CellSize()
	cols, rows := shotCells(caps.Columns, caps.Rows, cellW, cellH, width, height)
	if cols == 0 {
		return shotLayout{}, false
	}
	enc := fimage.EncodingBlocks
	if kitty {
		enc = fimage.EncodingKitty
	}
	_, isTmux := caps.Env.LookupEnv("TMUX")
	return shotLayout{
		enc:    enc,
		cell:   fimage.CellSize{Width: cellW, Height: cellH},
		cols:   cols,
		rows:   rows,
		isTmux: isTmux,
	}, true
}

// sendShot queues the current step's picture for the terminal. Decoding and
// encoding run on the command goroutine: a 960x600 capture is ~2.3 MB of
// pixels once expanded, which has no business happening between frames.
func (o *onboardingModel) sendShot() tea.Cmd {
	shot, _ := shotForStep(o.step)
	o.shot = shot
	o.shotCols, o.shotRows = 0, 0
	if shot.id == "" {
		return nil
	}
	layout, ok := o.shotLayoutFor(shot.id)
	if !ok {
		return nil
	}
	o.shotEnc, o.shotCols, o.shotRows = layout.enc, layout.cols, layout.rows
	if fimage.HasTransmitted(shot.id, layout.cols, layout.rows) {
		// Already on the terminal at this size: no decode, no re-transmit.
		return nil
	}
	return func() tea.Msg {
		img, err := decodeShot(shot.id)
		if err != nil {
			return util.InfoMsg{Type: util.InfoTypeError, Msg: "tour screenshot unavailable"}
		}
		cmd := layout.enc.Transmit(shot.id, img, layout.cell, layout.cols, layout.rows, layout.isTmux)
		if cmd == nil {
			// Already on the terminal at this size.
			return nil
		}
		return cmd()
	}
}

// renderShot returns the step's picture with its caption, or fallback when the
// terminal cannot show pixels or the bytes have not landed yet. With no fallback
// given, the caption alone keeps the step from looking broken.
func (o *onboardingModel) renderShot(fallback string) string {
	shot, ok := shotForStep(o.step)
	if !ok {
		return fallback
	}
	caption := o.com.Styles.Initialize.Content.Render(shot.caption)
	if fallback == "" {
		fallback = caption
	}
	if o.shot.id != shot.id || o.shotCols == 0 {
		return fallback
	}
	body := o.shotEnc.Render(shot.id, o.shotCols, o.shotRows)
	if body == "" {
		return fallback
	}
	return body + "\n" + caption
}

// shotPixelSize reads only the PNG header. The cell box has to be known before
// the picture is decoded, and this is the one part that is cheap enough to do
// on the update goroutine.
func shotPixelSize(id string) (width, height int, err error) {
	cfg, _, err := decodeShotConfig(id)
	if err != nil {
		return 0, 0, err
	}
	return cfg.Width, cfg.Height, nil
}

func decodeShotConfig(id string) (image.Config, string, error) {
	data, err := tourShots.ReadFile("shots/" + id + ".png")
	if err != nil {
		return image.Config{}, "", err
	}
	return image.DecodeConfig(bytes.NewReader(data))
}

func decodeShot(id string) (image.Image, error) {
	data, err := tourShots.ReadFile("shots/" + id + ".png")
	if err != nil {
		return nil, err
	}
	return png.Decode(bytes.NewReader(data))
}

// shotCells fits an imgW x imgH picture into the terminal as a cell box,
// keeping its aspect ratio. It returns 0,0 when there is no room worth trying.
func shotCells(termCols, termRows, cellW, cellH, imgW, imgH int) (cols, rows int) {
	if termCols <= 0 || termRows <= 0 || imgW <= 0 || imgH <= 0 {
		return 0, 0
	}
	cols = min(termCols-shotColMargin, shotMaxCols)
	if cols < shotMinCols {
		return 0, 0
	}
	rows = shotRowsFor(cols, cellW, cellH, imgW, imgH)
	budget := termRows - shotRowBudget
	if rows > budget {
		rows = budget
		if rows < shotMinRows {
			return 0, 0
		}
		cols = shotColsFor(rows, cellW, cellH, imgW, imgH)
		if cols < shotMinCols {
			return 0, 0
		}
	}
	return cols, rows
}

// A cell is cellW x cellH pixels and the picture keeps its aspect ratio, so
// rows = cols * cellW * imgH / (cellH * imgW). Terminals that never answered
// the cell-size query get the 2:1 cell every modern one actually has.
func shotRowsFor(cols, cellW, cellH, imgW, imgH int) int {
	cellW, cellH = safeCell(cellW, cellH)
	return (cols*cellW*imgH + cellH*imgW - 1) / (cellH * imgW)
}

func shotColsFor(rows, cellW, cellH, imgW, imgH int) int {
	cellW, cellH = safeCell(cellW, cellH)
	return rows * cellH * imgW / (cellW * imgH)
}

func safeCell(cellW, cellH int) (int, int) {
	if cellW <= 0 || cellH <= 0 {
		return 1, 2
	}
	return cellW, cellH
}
