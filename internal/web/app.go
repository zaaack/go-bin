package web

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"go-bin/internal/config"
	"go-bin/internal/service"

	"gorm.io/gorm"
)

type App struct {
	cfg  config.Config
	svc  *service.Service
	tmpl *template.Template
}

//go:embed assets/templates/*.html assets/static/*
var embeddedAssets embed.FS

func NewApp(cfg config.Config, db *gorm.DB) (*App, error) {
	templates, err := fs.Sub(embeddedAssets, "assets/templates")
	if err != nil {
		return nil, err
	}
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"summaryTwoLines": service.SummaryTwoLines,
		"isDefaultExpire": func(current, expected string) bool {
			return current == expected
		},
		"t": translate,
		"kindLabel": kindLabel,
		"isImage": isImage,
		"isVideo": isVideo,
		"minus": func(a, b int) int { return a - b },
	}).ParseFS(templates, "*.html")
	if err != nil {
		return nil, err
	}
	return &App{cfg: cfg, svc: service.New(db, cfg), tmpl: tmpl}, nil
}

func (a *App) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.handleIndex)
	mux.HandleFunc("/new", a.handleNew)
	mux.HandleFunc("/shares", a.handleCreate)
	mux.HandleFunc("/s/", a.handleDetail)
	mux.HandleFunc("/download/", a.handleDownload)
	mux.HandleFunc("/download-file/", a.handleDownloadFile)
	mux.HandleFunc("/view/", a.handleView)
	mux.HandleFunc("/toggle-pin/", a.handleTogglePin)
	mux.HandleFunc("/delete/", a.handleDelete)
	staticFS, err := fs.Sub(embeddedAssets, "assets/static")
	if err == nil {
		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	}
	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Expose-Headers", "Location")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *App) baseData() map[string]any {
	return map[string]any{
		"Now": time.Now().UTC(),
		"Cfg": a.cfg,
	}
}

func (a *App) render(w http.ResponseWriter, r *http.Request, name string, data map[string]any) {
	lang := detectLanguage(r.Header.Get("Accept-Language"))
	for k, v := range a.baseData() {
		data[k] = v
	}
	data["Lang"] = lang
	data["HTMLLang"] = strings.ReplaceAll(lang, "_", "-")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = a.tmpl.ExecuteTemplate(w, name, data)
}
