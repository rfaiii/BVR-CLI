package bots

import (
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"os"
)

// WritePilloGIF writes one loop of idle, skeptical, talk, and select.
func WritePilloGIF(path string) error {
	pal := pilloPalette()
	var frames []*image.Paletted
	var delays []int
	for _, name := range []string{"idle", "skeptical", "talk", "select"} {
		seq := pilloClip.Seqs[name]
		for _, step := range seq.Steps {
			frames = append(frames, paintPillo(pilloClip.Frames[step.Frame], pal, 8))
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

// WritePilloStrip writes the six frames in a row.
func WritePilloStrip(path string) error {
	const scale = 10
	const gap = 8
	w := pilloW*scale*6 + gap*7
	h := pilloH*scale + 8
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bg := color.RGBA{8, 8, 10, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, bg)
		}
	}
	for i, rows := range pilloClip.Frames {
		ox := i * (pilloW*scale + gap)
		blit(img, rows, ox, 4, scale)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func pilloPalette() color.Palette {
	return color.Palette{
		color.RGBA{8, 8, 10, 255},
		PilloBody,
		PilloEye,
	}
}

func paintPillo(rows []string, pal color.Palette, scale int) *image.Paletted {
	img := image.NewPaletted(image.Rect(0, 0, pilloW*scale, pilloH*scale), pal)
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			img.SetColorIndex(x, y, 0)
		}
	}
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

func blit(dst *image.RGBA, rows []string, ox, oy, scale int) {
	for y, row := range rows {
		for x, ch := range row {
			var c color.RGBA
			switch ch {
			case 'B':
				c = PilloBody
			case 'E':
				c = PilloEye
			default:
				continue
			}
			for dy := 0; dy < scale; dy++ {
				for dx := 0; dx < scale; dx++ {
					dst.SetRGBA(ox+x*scale+dx, oy+y*scale+dy, c)
				}
			}
		}
	}
}
