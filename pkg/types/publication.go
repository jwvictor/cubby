package types

import "unicode"

type VisibilityType string

const (
	Public     VisibilityType = "public"
	SingleUser VisibilityType = "user"
)

type Post struct {
	Id            string              `json:"post_id"`
	OwnerId       string              `json:"owner_id"`
	Visibility    []VisibilitySetting `json:"visibility"`
	BlobId        string              `json:"blob_id"`
	PublicationId string              `json:"publication_id"`
}

type VisibilitySetting struct {
	Type     VisibilityType `json:"type"`
	Audience string         `json:"audience"`
}

func ValidateCustomPostId(postId string) bool {
	if len(postId) > 64 {
		return false
	}
	return isAllowedCustomId(postId)
}

func SanitizePostId(title string) string {
	var postId string
	for _, r := range title {
		if r == ' ' {
			postId += "-"
		} else if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' || r == '_' {
			postId += string(r)
		}
	}
	if len(postId) > 64 {
		postId = postId[:64]
	}
	return postId
}

func isAllowedCustomId(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && r != '-' && r != '_' && !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}

func (p *Post) HasPermission(userId string) bool {
	for _, x := range p.Visibility {
		switch x.Type {
		case Public:
			return true
		case SingleUser:
			if userId == x.Audience {
				return true
			}
		}
	}
	return false
}
