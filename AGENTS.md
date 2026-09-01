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
- Login: `POST /api/auth/login` with `{username, password}`; an empty username validates the legacy config password and logs in as the root admin. Session endpoints: `refresh` (rotates), `logout`, `validate`. Admin-account management: `POST /api/auth/admins/{list,create,delete,password}`.
- Admin accounts live in the `admins` table (bcrypt hashes); a seed root admin account is created from the config credentials on first startup (`admin.username` + `admin.password`, defaults `admin`/`admin123`). Sessions are persisted in `admin_sessions` (access 1h + refresh 30d, refresh rotates) and survive restarts. Every mutation is recorded in the `change_log` tree with the acting admin; per-application history: `POST /api/application/history`, global feed: `POST /api/application/changes`.
- Admin UI routes: `/admin` (AdminPanel.vue, desktop), `/admin/mobile` (MobileAdminPanel.vue), `/admin/stats` (AdminStats.vue), `/admin/accounts` (AdminAccounts.vue — admin account management). Shared: `ProfileMenu.vue` (top-right avatar + logout dropdown) and `components/NotesModal.vue`.
- `import_csv.go` is a second `package main` at the repo root — a one-off CSV importer expecting `applications.csv` in cwd. Build the server explicitly (`go build -o tyforms ./cmd/webserver`); `go build ./...` builds both mains.

## Gotchas

- Port `:5099` is hardcoded in `main.go` (`http.ListenAndServe`); `server.port` in config is loaded but ignored.
- Config: `config.yaml` in cwd (gitignored, same dir the binary runs from); defaults are db `applications.db` + root admin `admin`/`admin123`. `admin.username` sets the root admin name used for seeding and for legacy password-only logins. Air runs with `APP_ENV=dev APP_USER=air`.
- `mattn/go-sqlite3` requires cgo (a C compiler must be installed).
- The `tyforms` binary is tracked in git — stale build output, not a source file. Rebuild (`make build-backend`) after backend changes and restart: an outdated running binary has no new `/api/*` routes, so requests fall through to the SPA handler and return `index.html` HTML instead of JSON.
- `AdminPanel.vue` (desktop) and `MobileAdminPanel.vue` are separate near-duplicate implementations; admin UI changes usually need both.
- Caddyfile maps: apply.tysmp.com → :5099 + `wwwroot/` SPA; vote/rules.tysmp.com → static dirs; wordle.tysmp.com → :3000 API + `wordle/public`.
- `howtousepteroapi.md` is reference notes for the Pterodactyl API (controlling the Minecraft server), not part of the app.
