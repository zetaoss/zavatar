package domain

type UserProfile struct {
	Name  string
	Type  string // "letter" | "identicon" | "gravatar"
	GHash string // gravatar hash (md5 hex)
}
