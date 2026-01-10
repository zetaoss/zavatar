// internal/domain/profile.go
package domain

type AvatarType string

const (
	AvatarTypeIdenticon AvatarType = "identicon"
	AvatarTypeLetter    AvatarType = "letter"
	AvatarTypeGravatar  AvatarType = "gravatar"
)

type AvatarTypeCode int

const (
	AvatarTypeCodeIdenticon AvatarTypeCode = 1
	AvatarTypeCodeLetter    AvatarTypeCode = 2
	AvatarTypeCodeGravatar  AvatarTypeCode = 3
)

type UserProfile struct {
	Name  string
	Type  AvatarType
	GHash string
}

func NormalizeAvatarType(t AvatarType) AvatarType {
	switch t {
	case AvatarTypeIdenticon, AvatarTypeLetter, AvatarTypeGravatar:
		return t
	default:
		return AvatarTypeIdenticon
	}
}
func AvatarTypeFromCode(code AvatarTypeCode) AvatarType {
	switch code {
	case AvatarTypeCodeLetter:
		return AvatarTypeLetter
	case AvatarTypeCodeGravatar:
		return AvatarTypeGravatar
	case AvatarTypeCodeIdenticon:
		fallthrough
	default:
		return AvatarTypeIdenticon
	}
}
