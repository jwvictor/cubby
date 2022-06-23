package types

import (
	"net/http"
	"strings"
	"time"
)

type AttachmentType string

const (
	Attachment          = "attachment"
	EncryptedAttachment = "encrypted_attachment"
	EncryptedBody       = "encrypted_body"
)

type BlobHistoryItem struct {
	AsOf      time.Time `json:"history_time"`
	BlobValue *Blob     `json:"blob_value"`
}

type BlobBinaryAttachment struct {
	Description string         `json:"description"`
	Data        []byte         `json:"data"`
	Type        AttachmentType `json:"type"`
}

type Blob struct {
	Id             string                 `json:"id"`
	Title          string                 `json:"title"`
	Type           string                 `json:"type"`
	Data           string                 `json:"data"`
	RawData        []BlobBinaryAttachment `json:"raw_data"`
	Importance     int                    `json:"importance"`
	Tags           []string               `json:"tags"`
	Deleted        bool                   `json:"deleted"`
	ChildIds       []string               `json:"child_ids,omitempty"`
	Children       []*Blob                `json:"children,omitempty"`
	OwnerId        string                 `json:"owner_id,omitempty"`
	ParentId       string                 `json:"parent_id,omitempty"`
	ExpireTime     *time.Time             `json:"expire_time,omitempty"`
	VersionHistory []*BlobHistoryItem     `json:"version_history,omitempty"`
}

func (blob *Blob) IsEncryptedAndEmpty() bool {
	if blob.Data == "" {
		for _, x := range blob.RawData {
			if x.Type == EncryptedBody {
				return true
			}
		}
	}
	return false
}

func (blob *Blob) EncryptedBody() *BlobBinaryAttachment {
	for _, x := range blob.RawData {
		if x.Type == EncryptedBody {
			return &x
		}
	}
	return nil
}

type BlobSkeleton struct {
	Id         string          `json:"id"`
	Title      string          `json:"title"`
	Type       string          `json:"type"`
	Importance int             `json:"importance"`
	Tags       []string        `json:"tags"`
	Deleted    bool            `json:"deleted"`
	Children   []*BlobSkeleton `json:"children,omitempty"`
	OwnerId    string          `json:"owner_id,omitempty"`
	ParentId   string          `json:"parent_id,omitempty"`
}

type PostResponse struct {
	Posts         []*Post `json:"posts"`
	Blobs         []*Blob `json:"blobs,omitempty"`
	Body          string  `json:"body,omitempty"`
	EncryptedBody []byte  `json:"encrypted_body,omitempty"`
}

func (b *PostResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type BasicAuthCredentials struct {
	UserEmail    string `json:"email"`
	UserPassword string `json:"password"`
	DisplayName  string `json:"display_name,omitempty"`
}

type AuthTokens struct {
	AccessToken  string `json:"access"`
	RefreshToken string `json:"refresh"`
}

func cloneAttachments(attachments []BlobBinaryAttachment) []BlobBinaryAttachment {
	var cloned []BlobBinaryAttachment
	for _, x := range attachments {
		cloned = append(cloned, x)
	}
	return cloned
}

func (b *Blob) Clone() *Blob {
	return &Blob{
		Id:             b.Id,
		Title:          b.Title,
		Type:           b.Type,
		Data:           b.Data,
		RawData:        cloneAttachments(b.RawData),
		Importance:     b.Importance,
		Tags:           b.Tags,
		Deleted:        b.Deleted,
		ChildIds:       b.ChildIds,
		Children:       b.Children,
		OwnerId:        b.OwnerId,
		ParentId:       b.ParentId,
		VersionHistory: b.VersionHistory,
		ExpireTime:     b.ExpireTime,
	}
}

func _tagsMatch(tags []string, searchString string) bool {
	for _, t := range tags {
		if t == searchString {
			return true
		}
	}
	return false
}

func (b *Blob) Matches(searchString string) bool {
	return strings.Contains(b.Title, searchString) || strings.Contains(b.Data, searchString) || b.Type == searchString || _tagsMatch(b.Tags, searchString)
}

type BlobResponse struct {
	Blobs []*Blob `json:"blobs"`
}

type BlobList struct {
	RootBlobs []*BlobSkeleton `json:"root_blobs"`
}

type UserResponse struct {
	Id          string `json:"user_id"`
	Email       string `json:"user_email"`
	DisplayName string `json:"display_name"`
}
