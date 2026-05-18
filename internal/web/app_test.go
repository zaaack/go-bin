package web

import (
	"bytes"
	"database/sql"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-bin/internal/config"
	appstore "go-bin/internal/store"
)

func TestHTTPFlows(t *testing.T) {
	tmp := t.TempDir()
	uploads := filepath.Join(tmp, "uploads")
	if err := os.MkdirAll(uploads, 0o755); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(tmp, "test.db")

	db, err := appstore.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	cfg := config.Default()
	cfg.DBPath = dbPath
	cfg.UploadsDir = uploads
	cfg.BaseURL = "http://example.com"

	app, err := NewApp(cfg, db)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(app.Router())
	defer server.Close()

	fileSlug := createFileShare(t, server.URL, "测试报告.pdf", "hello file content\nsecond line\nthird line")
	textSlug := createTextShare(t, server.URL, "第一行\n第二行\n第三行")
	linkSlug := createLinkShare(t, server.URL, "https://example.com/path?a=1")
	privateSlug := createPrivateTextShare(t, server.URL, "secret body")

	t.Run("list shows public shares", func(t *testing.T) {
		resp, body := mustGet(t, server.URL+"/")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d", resp.StatusCode)
		}
		if !strings.Contains(body, "测试报告.pdf") || !strings.Contains(body, "第一行") || !strings.Contains(body, "example.com") {
			t.Fatalf("unexpected list body: %s", body)
		}
		if strings.Contains(body, privateSlug) {
			t.Fatalf("private slug leaked in list")
		}
	})

	t.Run("detail page works", func(t *testing.T) {
		resp, body := mustGet(t, server.URL+"/s/"+fileSlug)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d", resp.StatusCode)
		}
		if !strings.Contains(body, "下载文件") {
			t.Fatalf("detail page missing download action")
		}
	})

	t.Run("download returns file", func(t *testing.T) {
		resp, body := mustGet(t, server.URL+"/download/"+fileSlug)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d", resp.StatusCode)
		}
		if got := resp.Header.Get("Content-Disposition"); !strings.Contains(got, "测试报告.pdf") {
			t.Fatalf("Content-Disposition = %q", got)
		}
		if body != "hello file content\nsecond line\nthird line" {
			t.Fatalf("download body = %q", body)
		}
	})

	t.Run("toggle pin redirects", func(t *testing.T) {
		resp := mustPostForm(t, server.URL+"/toggle-pin/"+textSlug, nil)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusSeeOther {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})

	t.Run("private share detail works", func(t *testing.T) {
		resp, _ := mustGet(t, server.URL+"/s/"+privateSlug)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("private detail status = %d", resp.StatusCode)
		}
	})

	t.Run("delete removes share", func(t *testing.T) {
		resp := mustPostForm(t, server.URL+"/delete/"+linkSlug, nil)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusSeeOther {
			t.Fatalf("status = %d", resp.StatusCode)
		}
		resp2, _ := mustGet(t, server.URL+"/s/"+linkSlug)
		defer resp2.Body.Close()
		if resp2.StatusCode != http.StatusNotFound {
			t.Fatalf("deleted detail status = %d", resp2.StatusCode)
		}
	})
}

func createFileShare(t *testing.T, baseURL, filename, content string) string {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	mustWriteField(t, writer, "kind", "file")
	mustWriteField(t, writer, "title", filename)
	mustWriteField(t, writer, "is_public", "1")
	mustWriteField(t, writer, "expire", "3mo")
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(part, strings.NewReader(content)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	resp := mustPostMultipart(t, baseURL+"/shares", &body, writer.FormDataContentType())
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	return extractSlug(resp.Header.Get("Location"))
}

func createTextShare(t *testing.T, baseURL, text string) string {
	t.Helper()
	form := url.Values{}
	form.Set("kind", "text")
	form.Set("title", "文本分享")
	form.Set("text", text)
	form.Set("is_public", "1")
	form.Set("expire", "3mo")
	resp := mustPostForm(t, baseURL+"/shares", strings.NewReader(form.Encode()))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	return extractSlug(resp.Header.Get("Location"))
}

func createLinkShare(t *testing.T, baseURL, link string) string {
	t.Helper()
	form := url.Values{}
	form.Set("kind", "link")
	form.Set("title", "链接分享")
	form.Set("link", link)
	form.Set("is_public", "1")
	form.Set("expire", "3mo")
	resp := mustPostForm(t, baseURL+"/shares", strings.NewReader(form.Encode()))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	return extractSlug(resp.Header.Get("Location"))
}

func createPrivateTextShare(t *testing.T, baseURL, text string) string {
	t.Helper()
	form := url.Values{}
	form.Set("kind", "text")
	form.Set("title", "private")
	form.Set("text", text)
	form.Set("expire", "3mo")
	resp := mustPostForm(t, baseURL+"/shares", strings.NewReader(form.Encode()))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	return extractSlug(resp.Header.Get("Location"))
}

func mustGet(t *testing.T, url string) (*http.Response, string) {
	t.Helper()
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		t.Fatal(err)
	}
	return resp, string(body)
}

func mustPostForm(t *testing.T, url string, body io.Reader) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func mustPostMultipart(t *testing.T, url string, body io.Reader, contentType string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", contentType)
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func mustWriteField(t *testing.T, writer *multipart.Writer, key, value string) {
	t.Helper()
	if err := writer.WriteField(key, value); err != nil {
		t.Fatal(err)
	}
}

func extractSlug(location string) string {
	return strings.TrimPrefix(location, "/s/")
}

var _ *sql.DB
