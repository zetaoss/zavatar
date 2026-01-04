// internal/domain/key.go
package domain

import "fmt"

func KeyIdenticonPNG(userID int64, size int) string {
	return fmt.Sprintf("%d-s%d.png", userID, size)
}
func KeyLetterSVG(userID int64) string {
	return fmt.Sprintf("%d-letter.svg", userID)
}
