// internal/render/letter_svg.go
package render

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

func LetterSVG(name string) []byte {
	ch := pickLetter(name)
	bg := pickColorHex(name)

	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" role="img" aria-label="Avatar">
  <rect width="100" height="100" fill="%s"/>
  <text x="50" y="52"
        text-anchor="middle" dominant-baseline="middle"
        font-family="system-ui, -apple-system, Segoe UI, Roboto, Noto Sans KR, Apple SD Gothic Neo, sans-serif"
        font-weight="700" font-size="56" fill="#fff">%s</text>
</svg>`, bg, xmlEscape(string(ch)))

	return []byte(svg)
}

func pickLetter(name string) rune {
	name = strings.TrimSpace(name)
	if name == "" {
		return '?'
	}
	for _, r := range name {
		if r == ' ' || r == '\t' || r == '\n' {
			continue
		}
		if r >= 0xAC00 && r <= 0xD7A3 {
			return hangulChoseong(r)
		}
		if r >= 'a' && r <= 'z' {
			return r - 32
		}
		if r >= 'A' && r <= 'Z' {
			return r
		}
		if r >= '0' && r <= '9' {
			return r
		}
		return r
	}
	return '?'
}

var choseong = []rune("ㄱㄲㄴㄷㄸㄹㅁㅂㅃㅅㅆㅇㅈㅉㅊㅋㅌㅍㅎ")

func hangulChoseong(r rune) rune {
	sIndex := int(r - 0xAC00)
	ci := sIndex / (21 * 28)
	if ci < 0 || ci >= len(choseong) {
		return 'ㅇ'
	}
	return choseong[ci]
}

func pickColorHex(seed string) string {
	h := sha256.Sum256([]byte(seed))
	r := 40 + (h[0] % 160)
	g := 40 + (h[1] % 160)
	b := 40 + (h[2] % 160)
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
