// internal/render/letter_svg.go
package render

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"unicode"

	"github.com/zetaoss/zavatar/internal/render/util"
)

func LetterSVG(siteSalt, name string) []byte {
	label := pickLetters(name)
	if label == "" {
		label = "?"
	}
	bg := pickColorHex(siteSalt + "|" + name)

	fontSize := 66
	letterSpacing := "0"
	y := 53

	n := len([]rune(label))
	if n > 1 {
		fontSize = 54
		letterSpacing = "-2"
		y = 52
	}

	svg := fmt.Sprintf(
		`<?xml version="1.0" encoding="UTF-8"?><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" role="img" aria-label="Avatar"><rect width="100" height="100" fill="%s"/><text x="50" y="%d" text-anchor="middle" dominant-baseline="middle" font-family="system-ui,-apple-system,Segoe UI,Roboto,Noto Sans KR,Apple SD Gothic Neo,sans-serif" font-weight="700" font-size="%d" letter-spacing="%s" fill="#fff">%s</text></svg>`,
		bg,
		y,
		fontSize,
		letterSpacing,
		xmlEscape(label),
	)

	return []byte(svg)
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

func pickColorHex(seed string) string {
	sum := sha256.Sum256([]byte(seed))

	h := float64(int(sum[0]) * 360 / 256)
	s := 0.62 + (float64(sum[1])/255.0)*0.24
	v := 0.40 + (float64(sum[2])/255.0)*0.25

	r, g, b := util.HSVToRGB(h, s, v)

	luma := 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
	if luma > 170 {
		r = uint8(float64(r) * 0.78)
		g = uint8(float64(g) * 0.78)
		b = uint8(float64(b) * 0.78)
	}

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func xmlEscape(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 8)
	for _, r := range s {
		if r <= 0x1F || (r >= 0x7F && r <= 0x9F) {
			continue
		}
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&apos;")
		default:
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" {
		return "?"
	}
	return out
}
