package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-bin/internal/config"
	"go-bin/internal/model"

	"gorm.io/gorm"
)

type CreateParams struct {
	Kind   string
	Title  string
	Text   string
	Link   string
	Public bool
	Pin    bool
	Expire string
	Files  []*multipart.FileHeader
}

type Service struct {
	db  *gorm.DB
	cfg config.Config
}

func New(db *gorm.DB, cfg config.Config) *Service {
	return &Service{db: db, cfg: cfg}
}

func ValidExpire(value string) bool {
	return config.ValidExpire(value)
}

func ParseExpire(value string, now time.Time) (*time.Time, error) {
	switch value {
	case "never":
		return nil, nil
	case "1d":
		t := now.AddDate(0, 0, 1)
		return &t, nil
	case "7d":
		t := now.AddDate(0, 0, 7)
		return &t, nil
	case "30d":
		t := now.AddDate(0, 0, 30)
		return &t, nil
	case "1mo":
		t := now.AddDate(0, 1, 0)
		return &t, nil
	case "3mo":
		t := now.AddDate(0, 3, 0)
		return &t, nil
	case "1y":
		t := now.AddDate(1, 0, 0)
		return &t, nil
	default:
		return nil, fmt.Errorf("invalid expire value: %s", value)
	}
}

func Slugify(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return "item"
	}

	var b strings.Builder
	lastSep := false
	for _, r := range input {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r >= 0x4e00 && r <= 0x9fff:
			b.WriteRune(r)
			lastSep = false
		case r == '-' || r == '_':
			b.WriteRune(r)
			lastSep = false
		default:
			if !lastSep {
				b.WriteByte('-')
				lastSep = true
			}
		}
	}

	result := strings.Trim(b.String(), "-_")
	if result == "" {
		return "item"
	}
	return result
}

func SummaryTwoLines(input string) string {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	lines := strings.Split(input, "\n")
	if len(lines) <= 2 {
		return strings.TrimSpace(input)
	}
	return strings.TrimSpace(strings.Join(lines[:2], "\n")) + "..."
}

func buildPrivateToken() string {
	buf := make([]byte, 18)
	_, _ = rand.Read(buf)
	return "p_" + base64.RawURLEncoding.EncodeToString(buf)
}

func buildPublicSlug(title string, now time.Time) string {
	base := Slugify(title)
	stamp := now.Format("20060102-150405")
	buf := make([]byte, 3)
	_, _ = rand.Read(buf)
	return fmt.Sprintf("%s-%s-%s", base, stamp, strings.ToLower(base64.RawURLEncoding.EncodeToString(buf)))
}

func fileTitle(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "file"
	}
	return name
}

func defaultTitle(kind, title, content string) string {
	if strings.TrimSpace(title) != "" {
		return strings.TrimSpace(title)
	}
	switch kind {
	case model.KindLink:
		u, err := url.Parse(strings.TrimSpace(content))
		if err == nil && u.Host != "" {
			return u.Host
		}
		return "link"
	case model.KindText:
		summary := SummaryTwoLines(content)
		if summary == "" {
			return "text"
		}
		if len([]rune(summary)) > 40 {
			return string([]rune(summary)[:40])
		}
		return summary
	default:
		return "item"
	}
}

func (s *Service) Create(ctx context.Context, params CreateParams) (*model.Share, error) {
	now := time.Now().UTC()
	expiresAt, err := ParseExpire(params.Expire, now)
	if err != nil {
		return nil, err
	}

	share := &model.Share{
		Kind:      params.Kind,
		IsPublic:  boolToInt(params.Public),
		IsPinned:  boolToInt(params.Pin),
		ExpiresAt: expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}

	switch params.Kind {
	case model.KindFile:
		if len(params.Files) == 0 {
			return nil, errors.New("at least one file is required")
		}
		share.Title = fileTitle(params.Title)
		if strings.TrimSpace(params.Title) == "" && len(params.Files) == 1 {
			share.Title = fileTitle(params.Files[0].Filename)
		} else if strings.TrimSpace(params.Title) == "" {
			share.Title = fmt.Sprintf("%d files", len(params.Files))
		}
		share.Slug = buildPrivateToken()
		if params.Public {
			share.Slug = buildPublicSlug(share.Title, now)
		}

		// Save all files
		for _, fh := range params.Files {
			storedPath, mimeType, sizeBytes, err := s.saveFile(fh, now)
			if err != nil {
				// Clean up already saved files
				for _, file := range share.Files {
					_ = os.Remove(filepath.Join(s.cfg.UploadsDir, file.StoredPath))
				}
				return nil, err
			}
			share.Files = append(share.Files, model.ShareFile{
				StoredPath:   storedPath,
				OriginalName: fh.Filename,
				MIMEType:     mimeType,
				SizeBytes:    sizeBytes,
				CreatedAt:    now,
			})
		}

		// Set legacy fields for backward compatibility (use first file)
		if len(share.Files) > 0 {
			share.StoredPath = share.Files[0].StoredPath
			share.OriginalName = share.Files[0].OriginalName
			share.MIMEType = share.Files[0].MIMEType
			share.SizeBytes = share.Files[0].SizeBytes
		}
	case model.KindText:
		if strings.TrimSpace(params.Text) == "" {
			return nil, errors.New("text is required")
		}
		share.Title = defaultTitle(model.KindText, params.Title, params.Text)
		share.ContentText = params.Text
		share.Slug = buildPrivateToken()
		if params.Public {
			share.Slug = buildPublicSlug(share.Title, now)
		}
	case model.KindLink:
		params.Link = strings.TrimSpace(params.Link)
		if params.Link == "" {
			return nil, errors.New("url is required")
		}
		if _, err := url.ParseRequestURI(params.Link); err != nil {
			return nil, errors.New("invalid url")
		}
		share.Title = defaultTitle(model.KindLink, params.Title, params.Link)
		share.ContentText = params.Link
		share.Slug = buildPrivateToken()
		if params.Public {
			share.Slug = buildPublicSlug(share.Title, now)
		}
	default:
		return nil, fmt.Errorf("unsupported share kind: %s", params.Kind)
	}

	if err := s.insertShare(share); err != nil {
		for _, file := range share.Files {
			_ = os.Remove(filepath.Join(s.cfg.UploadsDir, file.StoredPath))
		}
		return nil, err
	}

	return share, nil
}

func (s *Service) saveFile(fh *multipart.FileHeader, now time.Time) (string, string, int64, error) {
	opened, err := fh.Open()
	if err != nil {
		return "", "", 0, err
	}
	defer opened.Close()

	buffer := make([]byte, 512)
	n, err := opened.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", "", 0, err
	}
	mimeType := http.DetectContentType(buffer[:n])
	if _, err := opened.Seek(0, io.SeekStart); err != nil {
		return "", "", 0, err
	}

	storedName := buildPrivateToken()
	if ext := filepath.Ext(fh.Filename); ext != "" {
		storedName += ext
	}
	dstPath := filepath.Join(s.cfg.UploadsDir, storedName)
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", "", 0, err
	}
	defer dst.Close()

	size, err := io.Copy(dst, opened)
	if err != nil {
		_ = os.Remove(dstPath)
		return "", "", 0, err
	}

	return storedName, mimeType, size, nil
}

func (s *Service) insertShare(share *model.Share) error {
	return s.db.Create(share).Error
}

func (s *Service) ListPublic(_ context.Context) ([]model.Share, error) {
	var shares []model.Share
	err := s.db.
		Preload("Files").
		Where("is_public = ? AND (expires_at IS NULL OR expires_at > ?)", 1, time.Now().UTC()).
		Order("is_pinned DESC, created_at DESC").
		Find(&shares).Error
	return shares, err
}

func (s *Service) GetBySlug(_ context.Context, slug string) (*model.Share, error) {
	var share model.Share
	err := s.db.
		Preload("Files").
		Where("slug = ? AND (expires_at IS NULL OR expires_at > ?)", slug, time.Now().UTC()).
		First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

func (s *Service) TogglePin(_ context.Context, slug string) error {
	var share model.Share
	if err := s.db.Where("slug = ?", slug).First(&share).Error; err != nil {
		return err
	}
	newPinned := 0
	if share.IsPinned == 0 {
		newPinned = 1
	}
	return s.db.Model(&share).Updates(map[string]any{
		"is_pinned":  newPinned,
		"updated_at": time.Now().UTC(),
	}).Error
}

func (s *Service) GetShareFileByID(_ context.Context, fileID int64) (*model.ShareFile, error) {
	var file model.ShareFile
	if err := s.db.First(&file, fileID).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (s *Service) Delete(_ context.Context, slug string) error {
	var share model.Share
	if err := s.db.Preload("Files").Where("slug = ?", slug).First(&share).Error; err != nil {
		return err
	}
	if err := s.db.Delete(&share).Error; err != nil {
		return err
	}
	// Delete all files
	for _, file := range share.Files {
		_ = os.Remove(filepath.Join(s.cfg.UploadsDir, file.StoredPath))
	}
	// Also delete legacy single file if exists
	if share.StoredPath != "" {
		_ = os.Remove(filepath.Join(s.cfg.UploadsDir, share.StoredPath))
	}
	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
