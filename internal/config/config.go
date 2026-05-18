package config

import (
	"fmt"
	"strings"
)

type Config struct {
	Addr           string
	DBPath         string
	UploadsDir     string
	BaseURL        string
	DefaultPublic  bool
	DefaultPin     bool
	DefaultExpire  string
}

func Default() Config {
	return Config{
		Addr:          ":8080",
		DBPath:        "data.db",
		UploadsDir:    "uploads",
		BaseURL:       "",
		DefaultPublic: true,
		DefaultPin:    false,
		DefaultExpire: "3mo",
	}
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Addr) == "" {
		return fmt.Errorf("addr is required")
	}
	if strings.TrimSpace(c.DBPath) == "" {
		return fmt.Errorf("db path is required")
	}
	if strings.TrimSpace(c.UploadsDir) == "" {
		return fmt.Errorf("uploads-dir is required")
	}
	if !ValidExpire(c.DefaultExpire) {
		return fmt.Errorf("invalid default-expire: %s", c.DefaultExpire)
	}
	return nil
}

func ValidExpire(v string) bool {
	switch v {
	case "never", "1d", "7d", "30d", "1mo", "3mo", "1y":
		return true
	default:
		return false
	}
}
