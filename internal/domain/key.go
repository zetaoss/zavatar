// internal/domain/key.go
package domain

import (
	"fmt"
	"path"
)

func KeyAvatar(prefix string, t AvatarTypeCode, userID int64, size int) string {
	name := fmt.Sprintf("%d-t%d-s%d.png", userID, int(t), size)
	if prefix == "" {
		return name
	}
	return path.Join(prefix, name)
}
