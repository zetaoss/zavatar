// internal/render/gravatar.go
package render

import (
	"fmt"
)

func GravatarURL(ghash string, size int) string {
	if size < 1 {
		size = 1
	}
	return fmt.Sprintf("https://www.gravatar.com/avatar/%s?s=%d", ghash, size)
}
