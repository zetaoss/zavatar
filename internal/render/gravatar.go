// internal/render/gravatar.go
package render

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var reMD5 = regexp.MustCompile(`^[a-f0-9]{32}$`)

func ValidGHash(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return reMD5.MatchString(s)
}

func GravatarURL(ghash string, size int) string {
	ghash = strings.TrimSpace(strings.ToLower(ghash))
	u := url.URL{
		Scheme: "https",
		Host:   "www.gravatar.com",
		Path:   "/avatar/" + ghash,
	}
	q := u.Query()
	q.Set("s", strconv.Itoa(size))
	q.Set("d", "404")
	u.RawQuery = q.Encode()
	return u.String()
}
