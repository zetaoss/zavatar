// internal/domain/size.go
package domain

import "strconv"

const DefaultSize = 320

var PersistentSizes = []int{16, 32, 64, 128, 320}

func NormalizeSizeInt(req int) int {
	if req <= 0 {
		return DefaultSize
	}

	for _, s := range PersistentSizes {
		if req <= s {
			return s
		}
	}
	return DefaultSize
}

func NormalizeSizeQuery(v string) int {
	if v == "" {
		return DefaultSize
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return DefaultSize
	}
	return NormalizeSizeInt(n)
}
