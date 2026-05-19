# AGENTS.md

## Project

Go-based lightweight file/text/link sharing web service. Chinese-first UI with English fallback via `Accept-Language`. SQLite metadata + local `uploads/` directory for files.

## Commands

```powershell
# Run tests
go test ./...

# Run dev server
go run ./cmd/go-bin serve

# Build executable (outputs go-bin.exe)
go generate .
# or
go build -o go-bin.exe ./cmd/go-bin
```

No Makefile, no linter config, no CI. Keep it simple.

## Architecture

```
cmd/go-bin/main.go     → entry point, CLI flags, wires everything
internal/config/       → Config struct, validation, defaults
internal/model/        → data types (Share, ShareFile)
internal/store/        → SQLite open + schema (auto-creates tables, no migrations)
internal/service/      → business logic (create, list, delete, slug generation)
internal/web/          → HTTP handlers, templates, static, i18n
internal/web/assets/   → embedded via //go:embed (templates + static + favicon)
```

Flow: `main → config → store.Open → web.NewApp → http.ListenAndServe`

## Key Quirks

- **Embedded assets**: Templates and static files in `internal/web/assets/` are compiled into the binary via `//go:embed`. Any changes require rebuild.
- **No migrations**: SQLite schema is created in `store.Open()` with `CREATE TABLE IF NOT EXISTS`. Schema changes go in that function.
- **Slug generation**: Public shares get `{slugified-title}-{timestamp}-{random}` slugs. Private shares get `p_{random}` tokens.
- **Multi-file support**: Shares can contain multiple files via `share_files` table. Legacy single-file fields (`stored_path`, `original_name` etc.) kept on `shares` table for backward compat.
- **i18n**: Translation keys in `internal/web/i18n.go`. Default is `zh-CN`. English only if `Accept-Language` starts with `en`.
- **Form parsing**: `handleCreate` tries multipart first, falls back to form. File kind requires multipart.
- **No auth**: All shares are accessible by slug. Private shares are just unguessable URLs.
- **Expiration**: Parsed in `service.ParseExpire`. Valid values: `never`, `1d`, `7d`, `30d`, `1mo`, `3mo`, `1y`.

## Testing

```powershell
# All tests
go test ./...

# Specific package
go test ./internal/web
go test ./internal/service
```

Tests use `httptest.NewServer` with real SQLite in temp dirs. No mocks, no fixtures files. The `internal/web/app_test.go` covers full HTTP flows (create, list, detail, download, toggle-pin, delete).

## Config Defaults

| Flag | Default |
|------|---------|
| `--addr` | `:8080` |
| `--db` | `data.db` |
| `--uploads-dir` | `uploads` |
| `--default-public` | `true` |
| `--default-pin` | `false` |
| `--default-expire` | `3mo` |
| `--single-file` | `true` |

## Routes

| Path | Method | Handler |
|------|--------|---------|
| `/` | GET | Public share list |
| `/new` | GET | Create form |
| `/shares` | POST | Create share |
| `/s/{slug}` | GET | Share detail |
| `/download/{slug}` | GET | Download file (legacy single-file) |
| `/download-file/{id}` | GET | Download file by ID (multi-file) |
| `/toggle-pin/{slug}` | POST | Toggle pin |
| `/delete/{slug}` | POST | Delete share |
