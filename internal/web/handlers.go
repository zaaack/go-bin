package web

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"go-bin/internal/model"
	"go-bin/internal/service"
)

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	shares, err := a.svc.ListPublic(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.render(w, r, "list.html", map[string]any{"Shares": shares})
}

func (a *App) handleNew(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	a.render(w, r, "new.html", map[string]any{"Defaults": a.cfg})
}

func (a *App) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(64 << 20); err != nil && err != http.ErrNotMultipart {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else if err == http.ErrNotMultipart {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if kind := strings.TrimSpace(r.FormValue("kind")); kind == model.KindFile && r.MultipartForm == nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}

	kind := strings.TrimSpace(r.FormValue("kind"))
	params := service.CreateParams{
		Kind:   kind,
		Title:  r.FormValue("title"),
		Text:   r.FormValue("text"),
		Link:   r.FormValue("link"),
		Public: formBool(r, "is_public", false),
		Pin:    formBool(r, "is_pinned", false),
		Expire: firstNonEmpty(r.FormValue("expire"), a.cfg.DefaultExpire),
	}
	if files, ok := fileHeaders(r, "files"); ok {
		params.Files = files
	}

	share, err := a.svc.Create(r.Context(), params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/s/"+share.Slug, http.StatusSeeOther)
}

func (a *App) handleDetail(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/s/")
	if slug == "" || strings.Contains(slug, "/") {
		http.NotFound(w, r)
		return
	}
	share, err := a.svc.GetBySlug(r.Context(), slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	a.render(w, r, "detail.html", map[string]any{"Share": share})
}

func (a *App) handleDownload(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/download/")
	share, err := a.svc.GetBySlug(r.Context(), slug)
	if err != nil || share.Kind != model.KindFile || share.StoredPath == "" {
		http.NotFound(w, r)
		return
	}
	path := filepath.Join(a.cfg.UploadsDir, share.StoredPath)
	f, err := os.Open(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	name := share.OriginalName
	if name == "" {
		name = share.Title
	}
	if name == "" {
		name = filepath.Base(path)
	}
	ctype := share.MIMEType
	if ctype == "" {
		ctype = contentTypeFromName(name)
	}
	if ctype == "" {
		buf := make([]byte, 512)
		n, _ := f.Read(buf)
		ctype = http.DetectContentType(buf[:n])
		_, _ = f.Seek(0, io.SeekStart)
	}
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
	_, _ = io.Copy(w, f)
}

func (a *App) handleDownloadFile(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/download-file/")
	if idStr == "" {
		http.NotFound(w, r)
		return
	}
	
	// Parse file ID
	var fileID int64
	if _, err := fmt.Sscanf(idStr, "%d", &fileID); err != nil {
		http.NotFound(w, r)
		return
	}
	
	// Get file from database
	file, err := a.svc.GetShareFileByID(r.Context(), fileID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	
	path := filepath.Join(a.cfg.UploadsDir, file.StoredPath)
	f, err := os.Open(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	name := file.OriginalName
	if name == "" {
		name = filepath.Base(path)
	}
	ctype := file.MIMEType
	if ctype == "" {
		ctype = contentTypeFromName(name)
	}
	if ctype == "" {
		buf := make([]byte, 512)
		n, _ := f.Read(buf)
		ctype = http.DetectContentType(buf[:n])
		_, _ = f.Seek(0, io.SeekStart)
	}
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
	_, _ = io.Copy(w, f)
}

func (a *App) handleTogglePin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	slug := strings.TrimPrefix(r.URL.Path, "/toggle-pin/")
	if err := a.svc.TogglePin(r.Context(), slug); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/s/"+slug, http.StatusSeeOther)
}

func (a *App) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	slug := strings.TrimPrefix(r.URL.Path, "/delete/")
	if err := a.svc.Delete(r.Context(), slug); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func formBool(r *http.Request, key string, def bool) bool {
	v := strings.TrimSpace(r.FormValue(key))
	if v == "" {
		return def
	}
	return v == "1" || strings.EqualFold(v, "true") || v == "on" || v == "yes"
}

func fileHeaders(r *http.Request, key string) ([]*multipart.FileHeader, bool) {
	if r.MultipartForm == nil {
		return nil, false
	}
	files := r.MultipartForm.File[key]
	if len(files) == 0 {
		return nil, false
	}
	return files, true
}

func contentTypeFromName(name string) string {
	if ext := filepath.Ext(name); ext != "" {
		return mime.TypeByExtension(ext)
	}
	return ""
}
