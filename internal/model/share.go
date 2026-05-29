package model

import "time"

const (
	KindFile = "file"
	KindText = "text"
	KindLink = "link"
)

type ShareFile struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"`
	ShareID      int64     `gorm:"index;not null"`
	StoredPath   string    `gorm:"not null"`
	OriginalName string    `gorm:"not null"`
	MIMEType     string    `gorm:"column:mime_type"`
	SizeBytes    int64     `gorm:"column:size_bytes"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (ShareFile) TableName() string {
	return "share_files"
}

type Share struct {
	ID           int64      `gorm:"primaryKey;autoIncrement"`
	Kind         string     `gorm:"not null"`
	Slug         string     `gorm:"uniqueIndex;not null"`
	Title        string     `gorm:"not null"`
	ContentText  string     `gorm:"column:content_text"`
	StoredPath   string     `gorm:"column:stored_path"`
	OriginalName string     `gorm:"column:original_name"`
	MIMEType     string     `gorm:"column:mime_type"`
	SizeBytes    int64      `gorm:"column:size_bytes"`
	IsPublic     int        `gorm:"column:is_public;type:integer"`
	IsPinned     int        `gorm:"column:is_pinned;type:integer"`
	ExpiresAt    *time.Time `gorm:"column:expires_at"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
	Files        []ShareFile `gorm:"foreignKey:ShareID;constraint:OnDelete:CASCADE"`
}

func (Share) TableName() string {
	return "shares"
}
