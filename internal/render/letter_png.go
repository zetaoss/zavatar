// internal/render/letter_png.go
package render

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strings"
	"unicode"

	"github.com/zetaoss/zavatar/internal/render/util"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

func LetterPNG(siteSalt, name string, size int) ([]byte, error) {
	if size < 1 {
		size = 1
	}

	label := pickLetters(name)
	if label == "" {
		label = "?"
	}

	bg := util.PickColorRGBA(siteSalt + "|" + name)
	fg := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)

	face, err := makeLetterFace(size, runeCount(label))
	if err != nil {
		return nil, err
	}
	defer func() { _ = face.Close() }()

	drawCenteredLabel(img, face, fg, label)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func makeLetterFace(imgSize int, n int) (font.Face, error) {
	ratio := 0.66
	if n > 1 {
		ratio = 0.54
	}
	fontPx := float64(imgSize) * ratio
	if fontPx < 8 {
		fontPx = 8
	}

	ft, err := opentype.Parse(gobold.TTF)
	if err != nil {
		return nil, fmt.Errorf("parse font: %w", err)
	}

	face, err := opentype.NewFace(ft, &opentype.FaceOptions{
		Size:    fontPx,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("new face: %w", err)
	}
	return face, nil
}

func drawCenteredLabel(dst *image.RGBA, face font.Face, fg color.Color, label string) {
	n := runeCount(label)

	track := 0
	if n > 1 {
		track = -max(dst.Bounds().Dx()/50, 1)
	}

	m := face.Metrics()
	ascent := m.Ascent.Ceil()
	descent := m.Descent.Ceil()

	height := ascent + descent
	baselineY := dst.Bounds().Dy()/2 + ascent - height/2

	rs := []rune(label)

	total := 0
	widths := make([]int, 0, len(rs))
	for i, r := range rs {
		w := measureRune(face, r)
		widths = append(widths, w)
		total += w
		if i != len(rs)-1 {
			total += track
		}
	}

	startX := (dst.Bounds().Dx() - total) / 2

	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(fg),
		Face: face,
	}

	x := startX
	for i, r := range rs {
		d.Dot = fixed.P(x, baselineY)
		d.DrawString(string(r))
		x += widths[i]
		if i != len(rs)-1 {
			x += track
		}
	}
}

func measureRune(face font.Face, r rune) int {
	d := &font.Drawer{Face: face}
	return d.MeasureString(string(r)).Ceil()
}

func runeCount(s string) int { return len([]rune(s)) }

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func pickLetters(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "?"
	}

	parts := strings.Fields(name)
	out := make([]rune, 0, 2)

	appendFirst := func(s string) {
		for _, r := range s {
			if !isGoodLabelRune(r) {
				continue
			}
			out = append(out, unicode.ToUpper(r))
			return
		}
	}

	for _, p := range parts {
		appendFirst(p)
		if len(out) >= 2 {
			break
		}
	}

	if len(out) == 0 {
		appendFirst(name)
	}

	if len(out) == 0 {
		return "?"
	}

	if len(out) > 2 {
		out = out[:2]
	}
	return string(out)
}

func isGoodLabelRune(r rune) bool {
	if r <= 0x1F || (r >= 0x7F && r <= 0x9F) {
		return false
	}
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}
