// internal/render/util/util.go
package util

import (
	"crypto/sha256"
	"image/color"
)

func HSVToRGB(h, s, v float64) (uint8, uint8, uint8) {
	if s <= 0 {
		x := uint8(clamp01(v) * 255)
		return x, x, x
	}

	h = h - float64(int(h/360.0))*360.0
	if h < 0 {
		h += 360
	}

	c := v * s
	x := c * (1 - abs(mod(h/60.0, 2)-1))
	m := v - c

	var rp, gp, bp float64
	switch {
	case h < 60:
		rp, gp, bp = c, x, 0
	case h < 120:
		rp, gp, bp = x, c, 0
	case h < 180:
		rp, gp, bp = 0, c, x
	case h < 240:
		rp, gp, bp = 0, x, c
	case h < 300:
		rp, gp, bp = x, 0, c
	default:
		rp, gp, bp = c, 0, x
	}

	r := uint8(clamp01(rp+m) * 255)
	g := uint8(clamp01(gp+m) * 255)
	b := uint8(clamp01(bp+m) * 255)
	return r, g, b
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func mod(a, b float64) float64 {
	return a - float64(int(a/b))*b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func PickColorRGBA(userSalt string) color.RGBA {
	sum := sha256.Sum256([]byte(userSalt))

	h := int(sum[0]) * 360 / 256
	s := 0.62 + (float64(sum[1])/255.0)*0.24
	v := 0.52 + (float64(sum[2])/255.0)*0.20

	r, g, b := HSVToRGB(float64(h), s, v)

	luma := 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
	if luma > 210 {
		r = uint8(float64(r) * 0.85)
		g = uint8(float64(g) * 0.85)
		b = uint8(float64(b) * 0.85)
	}

	return color.RGBA{R: r, G: g, B: b, A: 255}
}
