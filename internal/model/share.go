package model

import "time"

const (
	KindFile = "file"
	KindText = "text"
	KindLink = "link"
)

type ShareFile struct {
	ID           int64
	ShareID      int64
	StoredPath   string
	OriginalName string
	MIMEType     string
	SizeBytes    int64
	CreatedAt    time.Time
}

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
	Files        []ShareFile
}
