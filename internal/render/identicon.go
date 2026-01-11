// internal/render/identicon.go
package render

import (
	"bytes"
	"crypto/sha256"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"sort"
	"strconv"

	"github.com/zetaoss/zavatar/internal/render/util"
)

func IdenticonPNG(siteSalt string, userID int64, size int) ([]byte, error) {
	if size < 1 {
		size = 1
	}

	userSalt := siteSalt + "|" + strconv.FormatInt(userID, 10)

	bg := color.RGBA{R: 245, G: 245, B: 245, A: 255}
	fg := util.PickColorRGBA(userSalt)

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)

	sum := sha256.Sum256([]byte(userSalt))

	targetPad := util.Max(size/16, 1)
	if size >= 20 && targetPad < 2 {
		targetPad = 2
	}

	inner := util.Max(size-2*targetPad, 5)
	cell := util.Max(inner/5, 1)

	gridSize := 5 * cell
	pad := util.Max((size-gridSize)/2, 0)

	grid := buildGrid(sum)

	u := &image.Uniform{C: fg}
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
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

func buildGrid(seed [32]byte) [5][5]bool {
	const (
		minOn = 10
		maxOn = 17
	)

	desired := minOn + int(seed[0])%(maxOn-minOn+1)

	type slot struct {
		y, x   int
		weight int
		score  uint16
	}

	slots := make([]slot, 0, 15)
	idx := 1

	for y := range 5 {
		for x := range 3 {
			w := 2
			if x == 2 {
				w = 1
			}
			score := uint16(seed[idx])<<8 | uint16(seed[idx+1])
			idx += 2
			slots = append(slots, slot{y: y, x: x, weight: w, score: score})
		}
	}

	sort.Slice(slots, func(i, j int) bool {
		return slots[i].score < slots[j].score
	})

	remTwos, remOnes := 10, 5
	total := 0

	var grid [5][5]bool

	for _, s := range slots {
		if s.weight == 2 {
			remTwos--
		} else {
			remOnes--
		}

		if total+s.weight <= desired && canFill(desired-(total+s.weight), remTwos, remOnes) {
			total += s.weight
			grid[s.y][s.x] = true
			grid[s.y][4-s.x] = true
		}
	}

	return grid
}

func canFill(rem, twos, ones int) bool {
	if rem < 0 {
		return false
	}
	if rem == 0 {
		return true
	}

	max := 2*twos + ones
	if rem > max {
		return false
	}

	lo := rem - 2*twos
	if lo < 0 {
		lo = 0
	}
	hi := rem
	if hi > ones {
		hi = ones
	}
	if lo > hi {
		return false
	}

	return ((rem-lo)%2) == 0 || lo+1 <= hi
}
