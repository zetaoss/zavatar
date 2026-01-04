// internal/render/identicon_png.go
package render

import (
	"bytes"
	"crypto/sha256"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	"github.com/zetaoss/zavatar/internal/render/util"
)

func IdenticonPNG(userSalt string, size int) ([]byte, error) {
	if size < 1 {
		size = 1
	}

	bg := color.RGBA{R: 245, G: 245, B: 245, A: 255}
	fg := pickColorRGBA(userSalt)

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)

	sum := sha256.Sum256([]byte(userSalt))

	targetPad := max(size/10, 1)
	if size >= 20 && targetPad < 2 {
		targetPad = 2
	}

	inner := max(size-2*targetPad, 5)
	cell := max(inner/5, 1)

	gridSize := 5 * cell
	pad := max((size-gridSize)/2, 0)

	const threshold = 85

	grid := [5][5]bool{}
	stream := newBitStream(sum[:])

	for y := range 5 {
		for x := range 3 {
			v := stream.nextByte()
			on := v < threshold
			grid[y][x] = on
			grid[y][4-x] = on
		}
	}

	u := &image.Uniform{C: fg}
	for y := range 5 {
		for x := range 5 {
			if !grid[y][x] {
				continue
			}
			x0 := pad + x*cell
			y0 := pad + y*cell
			r := image.Rect(x0, y0, x0+cell, y0+cell)
			draw.Draw(img, r, u, image.Point{}, draw.Src)
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func pickColorRGBA(userSalt string) color.RGBA {
	sum := sha256.Sum256([]byte(userSalt))

	h := int(sum[0]) * 360 / 256
	s := 0.62 + (float64(sum[1])/255.0)*0.24
	v := 0.52 + (float64(sum[2])/255.0)*0.20

	r, g, b := util.HSVToRGB(float64(h), s, v)

	luma := 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
	if luma > 210 {
		r = uint8(float64(r) * 0.85)
		g = uint8(float64(g) * 0.85)
		b = uint8(float64(b) * 0.85)
	}

	return color.RGBA{R: r, G: g, B: b, A: 255}
}

type bitStream struct {
	block  [32]byte
	i      int
	bitPos uint8
}

func newBitStream(bytes []byte) *bitStream {
	var bs bitStream
	copy(bs.block[:], bytes)
	return &bs
}

func (s *bitStream) refill() {
	s.block = sha256.Sum256(s.block[:])
	s.i = 0
	s.bitPos = 0
}

func (s *bitStream) nextByte() byte {
	if s.i >= len(s.block) {
		s.refill()
	}
	b := s.block[s.i]
	s.i++
	return b
}
