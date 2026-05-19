package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"go-bin/internal/config"
	"go-bin/internal/store"
	"go-bin/internal/web"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "serve" {
		fmt.Fprintf(os.Stderr, "usage: %s serve [flags]\n", os.Args[0])
		os.Exit(2)
	}

	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	cfg := config.Default()
	fs.StringVar(&cfg.Addr, "addr", cfg.Addr, "HTTP listen address")
	fs.StringVar(&cfg.DBPath, "db", cfg.DBPath, "SQLite database path")
	fs.StringVar(&cfg.UploadsDir, "uploads-dir", cfg.UploadsDir, "Uploads directory")
	fs.StringVar(&cfg.BaseURL, "base-url", cfg.BaseURL, "External base URL")
	fs.BoolVar(&cfg.DefaultPublic, "default-public", cfg.DefaultPublic, "Default public value for new shares")
	fs.BoolVar(&cfg.DefaultPin, "default-pin", cfg.DefaultPin, "Default pin value for new shares")
	fs.StringVar(&cfg.DefaultExpire, "default-expire", cfg.DefaultExpire, "Default expiration: never, 1d, 7d, 30d, 1mo, 3mo, 1y")
	fs.BoolVar(&cfg.SingleFile, "single-file", cfg.SingleFile, "Auto-submit when a single file is selected (disable for multi-file mode)")
	if err := fs.Parse(os.Args[2:]); err != nil {
		log.Fatal(err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	db, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := os.MkdirAll(cfg.UploadsDir, 0o755); err != nil {
		log.Fatal(err)
	}

	app, err := web.NewApp(cfg, db)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, app.Router()); err != nil {
		log.Fatal(err)
	}
}
