// internal/domain/size.go
package domain

import "strconv"

const (
	Size16  = 16
	Size64  = 64
	Size256 = 256

	DefaultSize = Size64
)

var SupportedSizes = []int{Size16, Size64, Size256}

func NormalizeSize(q string) int {
	if q == "" {
		return DefaultSize
	}
	n, err := strconv.Atoi(q)
	if err != nil || n <= 0 {
		return DefaultSize
	}
	for _, v := range SupportedSizes {
		if n <= v {
			return v
		}
	}
	return Size256
}
