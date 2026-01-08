// internal/render/util/util_test.go
package util

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHSVToRGB_GrayWhenSZero(t *testing.T) {
	r, g, b := HSVToRGB(123, 0, 0)
	require.Equal(t, uint8(0), r)
	require.Equal(t, uint8(0), g)
	require.Equal(t, uint8(0), b)

	r, g, b = HSVToRGB(123, 0, 1)
	require.Equal(t, uint8(255), r)
	require.Equal(t, uint8(255), g)
	require.Equal(t, uint8(255), b)

	r, g, b = HSVToRGB(123, 0, 0.5)
	require.Equal(t, uint8(127), r)
	require.Equal(t, uint8(127), g)
	require.Equal(t, uint8(127), b)

	r, g, b = HSVToRGB(123, 0, 0.25)
	require.Equal(t, uint8(63), r)
	require.Equal(t, uint8(63), g)
	require.Equal(t, uint8(63), b)
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
			require.Equal(t, tc.wantR, r)
			require.Equal(t, tc.wantG, g)
			require.Equal(t, tc.wantB, b)
		})
	}
}

func TestHSVToRGB_HueWraps(t *testing.T) {
	r1, g1, b1 := HSVToRGB(10, 1, 1)
	r2, g2, b2 := HSVToRGB(370, 1, 1)
	require.Equal(t, r1, r2)
	require.Equal(t, g1, g2)
	require.Equal(t, b1, b2)

	r3, g3, b3 := HSVToRGB(350, 1, 1)
	r4, g4, b4 := HSVToRGB(-10, 1, 1)
	require.Equal(t, r3, r4)
	require.Equal(t, g3, g4)
	require.Equal(t, b3, b4)
}

func TestHSVToRGB_ClampsV(t *testing.T) {
	r, g, b := HSVToRGB(0, 0, 2)
	require.Equal(t, uint8(255), r)
	require.Equal(t, uint8(255), g)
	require.Equal(t, uint8(255), b)

	r, g, b = HSVToRGB(0, 0, -1)
	require.Equal(t, uint8(0), r)
	require.Equal(t, uint8(0), g)
	require.Equal(t, uint8(0), b)
}

func TestHSVToRGB_ClampsRGBWhenSVOutOfRange(t *testing.T) {
	r, g, b := HSVToRGB(0, 2, 1)
	require.Equal(t, uint8(255), r)
	require.Equal(t, uint8(0), g)
	require.Equal(t, uint8(0), b)

	r, g, b = HSVToRGB(120, 2, 1)
	require.Equal(t, uint8(0), r)
	require.Equal(t, uint8(255), g)
	require.Equal(t, uint8(0), b)

	r, g, b = HSVToRGB(240, 2, 1)
	require.Equal(t, uint8(0), r)
	require.Equal(t, uint8(0), g)
	require.Equal(t, uint8(255), b)
}

func TestHSVToRGB_QuarterHuesHaveTwoChannels(t *testing.T) {
	r, g, b := HSVToRGB(30, 1, 1)
	require.Equal(t, uint8(255), r)
	require.Equal(t, uint8(127), g)
	require.Equal(t, uint8(0), b)

	r, g, b = HSVToRGB(90, 1, 1)
	require.Equal(t, uint8(127), r)
	require.Equal(t, uint8(255), g)
	require.Equal(t, uint8(0), b)

	r, g, b = HSVToRGB(150, 1, 1)
	require.Equal(t, uint8(0), r)
	require.Equal(t, uint8(255), g)
	require.Equal(t, uint8(127), b)

	r, g, b = HSVToRGB(210, 1, 1)
	require.Equal(t, uint8(0), r)
	require.Equal(t, uint8(127), g)
	require.Equal(t, uint8(255), b)

	r, g, b = HSVToRGB(270, 1, 1)
	require.Equal(t, uint8(127), r)
	require.Equal(t, uint8(0), g)
	require.Equal(t, uint8(255), b)

	r, g, b = HSVToRGB(330, 1, 1)
	require.Equal(t, uint8(255), r)
	require.Equal(t, uint8(0), g)
	require.Equal(t, uint8(127), b)
}

func TestPickColorRGBA_Deterministic(t *testing.T) {
	t.Parallel()

	c1 := PickColorRGBA("user|alice")
	c2 := PickColorRGBA("user|alice")

	require.Equal(t, c1, c2, "same input must produce same color")
}

func TestPickColorRGBA_DifferentInputs(t *testing.T) {
	t.Parallel()

	c1 := PickColorRGBA("user|alice")
	c2 := PickColorRGBA("user|bob")

	require.NotEqual(t, c1, c2, "different inputs should produce different colors")
}

func TestPickColorRGBA_ValidRange(t *testing.T) {
	t.Parallel()

	tests := []string{
		"user|alice",
		"user|bob",
		"user|홍길동",
		"user|123",
		"",
	}

	for _, seed := range tests {
		seed := seed
		t.Run(seed, func(t *testing.T) {
			t.Parallel()

			c := PickColorRGBA(seed)

			requireColorRange(t, c)
		})
	}
}

func TestPickColorRGBA_LumaClampDoesNotOverflow(t *testing.T) {
	t.Parallel()

	for i := 0; i < 256; i++ {
		c := PickColorRGBA(string(rune(i)))

		requireColorRange(t, c)
	}
}

// ---------- helpers ----------

func requireColorRange(t *testing.T, c color.RGBA) {
	t.Helper()

	require.GreaterOrEqual(t, c.R, uint8(0))
	require.GreaterOrEqual(t, c.G, uint8(0))
	require.GreaterOrEqual(t, c.B, uint8(0))
	require.Equal(t, uint8(255), c.A)

	require.LessOrEqual(t, c.R, uint8(255))
	require.LessOrEqual(t, c.G, uint8(255))
	require.LessOrEqual(t, c.B, uint8(255))
}
