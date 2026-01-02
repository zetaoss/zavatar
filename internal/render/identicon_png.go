package render

import (
	"bytes"
	"crypto/sha256"
	"image"
	"image/color"
	"image/draw"
	"image/png"
)

func IdenticonPNG(seed string, size int) ([]byte, error) {
	bg := color.RGBA{R: 245, G: 245, B: 245, A: 255}
	fg := pickColorRGBA(seed)

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)

	h := sha256.Sum256([]byte(seed))

	grid := [5][5]bool{}
	bi := 0
	for y := 0; y < 5; y++ {
		for x := 0; x < 3; x++ {
			b := h[bi%len(h)]
			bi++
			on := (b & 1) == 1
			grid[y][x] = on
			grid[y][4-x] = on
		}
	}

	pad := size / 10
	if pad < 2 {
		pad = 2
	}
	cell := (size - 2*pad) / 5
	if cell < 1 {
		cell = 1
	}

	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			if !grid[y][x] {
				continue
			}
			x0 := pad + x*cell
			y0 := pad + y*cell
			r := image.Rect(x0, y0, x0+cell, y0+cell)
			draw.Draw(img, r, &image.Uniform{C: fg}, image.Point{}, draw.Src)
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func pickColorRGBA(seed string) color.RGBA {
	h := sha256.Sum256([]byte(seed))
	r := 60 + (h[0] % 180)
	g := 60 + (h[1] % 180)
	b := 60 + (h[2] % 180)
	return color.RGBA{R: r, G: g, B: b, A: 255}
}
