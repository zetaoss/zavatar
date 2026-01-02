// internal/render/gravatar.go
package render

import (
	"fmt"
	"net/url"
)

func GravatarURL(ghash string, size int) string {
	return fmt.Sprintf("https://www.gravatar.com/avatar/%s?s=%d", url.PathEscape(ghash), size)
}
