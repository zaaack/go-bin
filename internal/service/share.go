package service

import (
	"context"
	"crypto/rand"
	"database/sql"
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
)

type CreateParams struct {
	Kind    string
	Title   string
	Text    string
	Link    string
	Public  bool
	Pin     bool
	Expire  string
	File    *multipart.FileHeader
}

type Service struct {
	db  *sql.DB
	cfg config.Config
}

func New(db *sql.DB, cfg config.Config) *Service {
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
		IsPublic:  params.Public,
		IsPinned:  params.Pin,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}

	switch params.Kind {
	case model.KindFile:
		if params.File == nil {
			return nil, errors.New("file is required")
		}
		share.Title = fileTitle(params.Title)
		if strings.TrimSpace(params.Title) == "" {
			share.Title = fileTitle(params.File.Filename)
		}
		share.OriginalName = params.File.Filename
		share.Slug = buildPrivateToken()
		if params.Public {
			share.Slug = buildPublicSlug(share.Title, now)
		}
		storedPath, mimeType, sizeBytes, err := s.saveFile(params.File, now)
		if err != nil {
			return nil, err
		}
		share.StoredPath = storedPath
		share.MIMEType = mimeType
		share.SizeBytes = sizeBytes
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

	if err := s.insertShare(ctx, share); err != nil {
		if share.StoredPath != "" {
			_ = os.Remove(filepath.Join(s.cfg.UploadsDir, share.StoredPath))
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

func (s *Service) insertShare(ctx context.Context, share *model.Share) error {
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO shares (
			kind, slug, title, content_text, stored_path, original_name, mime_type,
			size_bytes, is_public, is_pinned, expires_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, share.Kind, share.Slug, share.Title, share.ContentText, share.StoredPath, share.OriginalName, share.MIMEType, share.SizeBytes, boolToInt(share.IsPublic), boolToInt(share.IsPinned), share.ExpiresAt, share.CreatedAt, share.UpdatedAt)
	if err != nil {
		return err
	}
	share.ID, _ = result.LastInsertId()
	return nil
}

func (s *Service) ListPublic(ctx context.Context) ([]model.Share, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, kind, slug, title, content_text, stored_path, original_name, mime_type,
			size_bytes, is_public, is_pinned, expires_at, created_at, updated_at
		FROM shares
		WHERE is_public = 1 AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
		ORDER BY is_pinned DESC, created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shares []model.Share
	for rows.Next() {
		share, err := scanShare(rows)
		if err != nil {
			return nil, err
		}
		shares = append(shares, share)
	}
	return shares, rows.Err()
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (*model.Share, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, kind, slug, title, content_text, stored_path, original_name, mime_type,
			size_bytes, is_public, is_pinned, expires_at, created_at, updated_at
		FROM shares
		WHERE slug = ? AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
	`, slug)
	share, err := scanShare(row)
	if err != nil {
		return nil, err
	}
	return &share, nil
}

func (s *Service) TogglePin(ctx context.Context, slug string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE shares
		SET is_pinned = CASE WHEN is_pinned = 1 THEN 0 ELSE 1 END,
			updated_at = ?
		WHERE slug = ?
	`, time.Now().UTC(), slug)
	return err
}

func (s *Service) Delete(ctx context.Context, slug string) error {
	share, err := s.GetBySlug(ctx, slug)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `DELETE FROM shares WHERE slug = ?`, slug)
	if err != nil {
		return err
	}
	if share.StoredPath != "" {
		_ = os.Remove(filepath.Join(s.cfg.UploadsDir, share.StoredPath))
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanShare(s scanner) (model.Share, error) {
	var share model.Share
	var expiresAt sql.NullTime
	var isPublic int
	var isPinned int
	if err := s.Scan(
		&share.ID,
		&share.Kind,
		&share.Slug,
		&share.Title,
		&share.ContentText,
		&share.StoredPath,
		&share.OriginalName,
		&share.MIMEType,
		&share.SizeBytes,
		&isPublic,
		&isPinned,
		&expiresAt,
		&share.CreatedAt,
		&share.UpdatedAt,
	); err != nil {
		return model.Share{}, err
	}
	share.IsPublic = isPublic == 1
	share.IsPinned = isPinned == 1
	if expiresAt.Valid {
		share.ExpiresAt = &expiresAt.Time
	}
	return share, nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
