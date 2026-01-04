// internal/render/util/util_test.go
package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHSVToRGB_GrayWhenSZero(t *testing.T) {
	r, g, b := HSVToRGB(123, 0, 0)
	assert.Equal(t, uint8(0), r)
	assert.Equal(t, uint8(0), g)
	assert.Equal(t, uint8(0), b)

	r, g, b = HSVToRGB(123, 0, 1)
	assert.Equal(t, uint8(255), r)
	assert.Equal(t, uint8(255), g)
	assert.Equal(t, uint8(255), b)

	r, g, b = HSVToRGB(123, 0, 0.5)
	assert.Equal(t, uint8(127), r)
	assert.Equal(t, uint8(127), g)
	assert.Equal(t, uint8(127), b)

	r, g, b = HSVToRGB(123, 0, 0.25)
	assert.Equal(t, uint8(63), r)
	assert.Equal(t, uint8(63), g)
	assert.Equal(t, uint8(63), b)
}

func TestHSVToRGB_PrimariesAndSecondaries(t *testing.T) {
	tests := []struct {
		name    string
		h, s, v float64
		wantR   uint8
		wantG   uint8
		wantB   uint8
	}{
		{"red_0", 0, 1, 1, 255, 0, 0},
		{"yellow_60", 60, 1, 1, 255, 255, 0},
		{"green_120", 120, 1, 1, 0, 255, 0},
		{"cyan_180", 180, 1, 1, 0, 255, 255},
		{"blue_240", 240, 1, 1, 0, 0, 255},
		{"magenta_300", 300, 1, 1, 255, 0, 255},
		{"red_360", 360, 1, 1, 255, 0, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, g, b := HSVToRGB(tc.h, tc.s, tc.v)
			assert.Equal(t, tc.wantR, r)
			assert.Equal(t, tc.wantG, g)
			assert.Equal(t, tc.wantB, b)
		})
	}
}

func TestHSVToRGB_HueWraps(t *testing.T) {
	r1, g1, b1 := HSVToRGB(10, 1, 1)
	r2, g2, b2 := HSVToRGB(370, 1, 1)
	assert.Equal(t, r1, r2)
	assert.Equal(t, g1, g2)
	assert.Equal(t, b1, b2)

	r3, g3, b3 := HSVToRGB(350, 1, 1)
	r4, g4, b4 := HSVToRGB(-10, 1, 1)
	assert.Equal(t, r3, r4)
	assert.Equal(t, g3, g4)
	assert.Equal(t, b3, b4)
}

func TestHSVToRGB_ClampsV(t *testing.T) {
	r, g, b := HSVToRGB(0, 0, 2)
	assert.Equal(t, uint8(255), r)
	assert.Equal(t, uint8(255), g)
	assert.Equal(t, uint8(255), b)

	r, g, b = HSVToRGB(0, 0, -1)
	assert.Equal(t, uint8(0), r)
	assert.Equal(t, uint8(0), g)
	assert.Equal(t, uint8(0), b)
}

func TestHSVToRGB_ClampsRGBWhenSVOutOfRange(t *testing.T) {
	r, g, b := HSVToRGB(0, 2, 1)
	assert.Equal(t, uint8(255), r)
	assert.Equal(t, uint8(0), g)
	assert.Equal(t, uint8(0), b)

	r, g, b = HSVToRGB(120, 2, 1)
	assert.Equal(t, uint8(0), r)
	assert.Equal(t, uint8(255), g)
	assert.Equal(t, uint8(0), b)

	r, g, b = HSVToRGB(240, 2, 1)
	assert.Equal(t, uint8(0), r)
	assert.Equal(t, uint8(0), g)
	assert.Equal(t, uint8(255), b)
}

func TestHSVToRGB_QuarterHuesHaveTwoChannels(t *testing.T) {
	r, g, b := HSVToRGB(30, 1, 1)
	assert.Equal(t, uint8(255), r)
	assert.Equal(t, uint8(127), g)
	assert.Equal(t, uint8(0), b)

	r, g, b = HSVToRGB(90, 1, 1)
	assert.Equal(t, uint8(127), r)
	assert.Equal(t, uint8(255), g)
	assert.Equal(t, uint8(0), b)

	r, g, b = HSVToRGB(150, 1, 1)
	assert.Equal(t, uint8(0), r)
	assert.Equal(t, uint8(255), g)
	assert.Equal(t, uint8(127), b)

	r, g, b = HSVToRGB(210, 1, 1)
	assert.Equal(t, uint8(0), r)
	assert.Equal(t, uint8(127), g)
	assert.Equal(t, uint8(255), b)

	r, g, b = HSVToRGB(270, 1, 1)
	assert.Equal(t, uint8(127), r)
	assert.Equal(t, uint8(0), g)
	assert.Equal(t, uint8(255), b)

	r, g, b = HSVToRGB(330, 1, 1)
	assert.Equal(t, uint8(255), r)
	assert.Equal(t, uint8(0), g)
	assert.Equal(t, uint8(127), b)
}
