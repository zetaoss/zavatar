// internal/domain/key.go
package domain

import "fmt"

func KeyIdenticonPNG(siteSaltHash string, userID int64, size int) string {
	return fmt.Sprintf("%s/%d-s%d.png", siteSaltHash, userID, size)
}
func KeyLetterSVG(siteSaltHash string, userID int64) string {
	return fmt.Sprintf("%s/%d-letter.svg", siteSaltHash, userID)
}
