package domain

import "fmt"

func KeyPNG(v int, userID int64, size int) string {
	return fmt.Sprintf("v%d/%d/s%d.png", v, userID, size)
}
func KeyLetterSVG(v int, userID int64) string {
	return fmt.Sprintf("v%d/%d/letter.svg", v, userID)
}
