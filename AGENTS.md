# AGENTS.md

Minecraft SMP application system ("tyforms"): Go backend serving a JSON API plus the built Vue 3 SPA from `wwwroot/`, fronted by Caddy. Satellite apps in this repo: `vote/` and `rules/` (plain static HTML), `wordle/` (standalone Node/Express service on :3000).

## Commands

- Backend: `make build-backend` (binary `./tyforms`), `make run-backend` (uses `air` hot-reload per `.air.toml` if installed, else runs the binary directly).
- Frontend: `make dev-frontend` (Vite on :5173, proxies `/api` to :5099 — run the backend alongside it). `make build-frontend` outputs to `../wwwroot` (gitignored).
- Format: `make fmt`. Tests: `make test` (`go test ./...`) — no test files exist yet.
- Wordle: `make run-wordle` / `make dev-wordle` (nodemon).

## Architecture

- Entry `cmd/webserver/main.go` → routes registered on the default mux → `internal/handlers/application_handler.go` → `internal/database/sqlite_store.go` (schema created at startup). Tables: `applications`, `admins`, `admin_sessions`, `change_log` (admin/session logic in `internal/database/admin_store.go`, change tree in `internal/database/change_log.go`; auth + admin-account endpoints in `internal/handlers/admin_handler.go`).
- All API endpoints are POST-only, under `/api/application/*` and `/api/auth/*`. Admin endpoints accept either a session token or the legacy config password in the JSON body; frontend helper is `FrontEnd/src/services/api.js`, tokens kept in `localStorage` (`adminToken` + `adminRefreshToken`, auto-refreshed on 401/proactively).
- Admin accounts live in the `admins` table (bcrypt hashes); a seed "admin" account is created from the config password on first startup. Sessions are persisted in `admin_sessions` (access 1h + refresh 30d, refresh rotates) and survive restarts. Every mutation is recorded in the `change_log` tree with the acting admin; per-application history: `POST /api/application/history`, global feed: `POST /api/application/changes`.
- `import_csv.go` is a second `package main` at the repo root — a one-off CSV importer expecting `applications.csv` in cwd. Build the server explicitly (`go build -o tyforms ./cmd/webserver`); `go build ./...` builds both mains.

## Gotchas

- Port `:5099` is hardcoded in `main.go` (`http.ListenAndServe`); `server.port` in config is loaded but ignored.
- Config: `config.yaml` in cwd (gitignored, same dir the binary runs from); defaults are db `applications.db` + root admin `admin`/`admin123`. `admin.username` sets the root admin name used for seeding and for legacy password-only logins. Air runs with `APP_ENV=dev APP_USER=air`.
- `mattn/go-sqlite3` requires cgo (a C compiler must be installed).
- The 13MB `tyforms` binary is tracked in git — stale build output, not a source file.
- `AdminPanel.vue` (desktop) and `MobileAdminPanel.vue` are separate near-duplicate implementations; admin UI changes usually need both.
- Caddyfile maps: apply.tysmp.com → :5099 + `wwwroot/` SPA; vote/rules.tysmp.com → static dirs; wordle.tysmp.com → :3000 API + `wordle/public`.
- `howtousepteroapi.md` is reference notes for the Pterodactyl API (controlling the Minecraft server), not part of the app.
