package model

import "time"

const (
	KindFile = "file"
	KindText = "text"
	KindLink = "link"
)

type Share struct {
	ID           int64
	Kind         string
	Slug         string
	Title        string
	ContentText  string
	StoredPath   string
	OriginalName string
	MIMEType     string
	SizeBytes    int64
	IsPublic     bool
	IsPinned     bool
	ExpiresAt    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
