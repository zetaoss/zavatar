// internal/render/util/util.go
package util

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
