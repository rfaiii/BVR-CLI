package bots

import (
	"image"
	"image/color"
	"image/gif"
	"os"
)

// RenderCompactPreview draws the five small mascots with their names.
func RenderCompactPreview() *image.RGBA {
	ids := []string{"dingus", "brobro", "pillo", "luvah", "gnarf"}
	grids := make([]MascotGrid, 0, len(ids))
	for _, id := range ids {
		rows, body, eye, ok := Compact(id)
		if !ok {
			continue
		}
		grids = append(grids, MascotGrid{Name: id, Body: body, Eye: eye, Rows: rows})
	}
	return RenderMascotPreview(grids)
}

// RenderChatPreview is a still of the NES chat: small Dingus, his name,
// the DEMO badge, and one short exchange.
func RenderChatPreview() *image.RGBA {
	const scale = 8
	rows, body, eye, _ := Compact("dingus")
	sw, sh := len(rows[0])*scale, len(rows)*scale
	width, height := 520, 280
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	bg := color.RGBA{8, 8, 10, 255}
	fill(img, bg)
	ox, oy := 28, 36
	for y, row := range rows {
		for x, ch := range row {
			var c color.RGBA
			switch ch {
			case 'B':
				c = body
			case 'E':
				c = eye
			default:
				continue
			}
			rect(img, ox+x*scale, oy+y*scale, scale, scale, c)
		}
	}
	drawLabel(img, ox+sw+24, oy+8, "DINGUS")
	drawLabel(img, width-28-labelWidth("DEMO"), 16, "DEMO")
	frame := color.RGBA{R: 0xFE, G: 0xFD, B: 0xFE, A: 0xFF}
	boxX, boxY, boxW, boxH := 24, oy+sh+20, width-48, 96
	rect(img, boxX, boxY, boxW, 4, frame)
	rect(img, boxX, boxY+boxH-4, boxW, 4, frame)
	rect(img, boxX, boxY, 4, boxH, frame)
	rect(img, boxX+boxW-4, boxY, 4, boxH, frame)
	drawLabel(img, boxX+16, boxY+16, "YOU  HEY")
	drawLabel(img, boxX+16, boxY+48, "DINGUS  OK")
	return img
}

// WriteCompactPilloGIF writes the small pill's idle, skeptical, talk,
// and select sequences.
func WriteCompactPilloGIF(path string) error {
	pal := color.Palette{
		color.RGBA{8, 8, 10, 255},
		PilloBody,
		PilloEye,
	}
	clip := PilloClipSmall()
	var frames []*image.Paletted
	var delays []int
	for _, name := range []string{"idle", "skeptical", "talk", "select"} {
		seq := clip.Seqs[name]
		for _, step := range seq.Steps {
			frames = append(frames, paintAny(clip.Frames[step.Frame], pal, 10))
			delays = append(delays, step.Ticks*10)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gif.EncodeAll(f, &gif.GIF{
		Image: frames,
		Delay: delays,
		Config: image.Config{
			ColorModel: pal,
			Width:      frames[0].Bounds().Dx(),
			Height:     frames[0].Bounds().Dy(),
		},
	})
}

func paintAny(rows []string, pal color.Palette, scale int) *image.Paletted {
	w, h := len(rows[0])*scale, len(rows)*scale
	img := image.NewPaletted(image.Rect(0, 0, w, h), pal)
	for y, row := range rows {
		for x, ch := range row {
			idx := uint8(0)
			switch ch {
			case 'B':
				idx = 1
			case 'E':
				idx = 2
			default:
				continue
			}
			for dy := 0; dy < scale; dy++ {
				for dx := 0; dx < scale; dx++ {
					img.SetColorIndex(x*scale+dx, y*scale+dy, idx)
				}
			}
		}
	}
	return img
}

func fill(img *image.RGBA, c color.RGBA) {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			img.SetRGBA(x, y, c)
		}
	}
}

func rect(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			img.SetRGBA(xx, yy, c)
		}
	}
}
