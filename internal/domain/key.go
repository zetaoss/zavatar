// internal/domain/key.go
package domain

import "fmt"

func KeyAvatar(userID int64, size int) string {
	return fmt.Sprintf("%d-s%d.png", userID, size)
}
