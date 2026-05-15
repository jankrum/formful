# Formful

Google Forms clone. Portfolio project.

## Principles

Simplicity, security, accessibility, performance, progressive enhancement, semantic HTML, semantic HTTP, Locality of Behavior (LoB).

## Tech Stack

| Concern    | Choice                                                        |
| ---------- | ------------------------------------------------------------- |
| Language   | Go 1.26                                                       |
| Router     | chi v5                                                        |
| Templates  | a-h/templ                                                     |
| CSS        | @jankrum/style.css (customized Bulma), PurgeCSS at build time |
| JS         | htmx, surreal.js (inline scripts, LoB)                        |
| Database   | PostgreSQL 17                                                 |
| Migrations | goose                                                         |
| DB codegen | sqlc                                                          |
| Auth       | Magic link email + optional passkey (go-webauthn/webauthn)    |
| Email      | Resend                                                        |
| CLI args   | go-arg                                                        |
| Logging    | slog (stdlib)                                                 |
| Testing    | stdlib testing + testcontainers-go                            |

Client assets (htmx, surreal, style.css) are fetched via `make vendor-assets` and gitignored. Never committed.

## Scaffolding

Generated with [go-blueprint](https://github.com/melkeydev/go-blueprint):

```bash
go-blueprint create --name github.com/jankrum/formful --framework chi --driver postgres \
  --advanced --feature htmx --feature docker --feature githubaction --git commit
```

Then modified: removed tailwind/goreleaser artifacts, renamed env vars, restructured into feature packages, added 3-stage Dockerfile, and extended Makefile.

## Directory Structure

```
cmd/
  api/main.go               entry point, go-arg config, graceful shutdown (30s)
  web/
    assets/js/              htmx.min.js, surreal.js (gitignored, vendored)
    assets/css/             style.css (source, gitignored), app.css (purged output, gitignored)
    base.templ              root layout — package web
    index.templ             landing page
    auth/handler.go         Handler struct + auth page handlers — package auth
    dashboard/handler.go    Handler struct + dashboard handlers — package dashboard
    public/handler.go       Handler struct + public form handlers — package public
internal/
  server/
    server.go               http.Server setup, Server struct (holds db + sqlc queries)
    routes.go               chi router, middleware chain, route registration
  database/
    database.go             DB connection, Service interface (Health, Close, DB() *sql.DB)
    database_test.go        testcontainers integration tests
    *.sql.go                sqlc-generated query code (lands here via sqlc.yaml)
  auth/                     session tokens, magic links, passkey ceremony wrappers
  email/                    Resend API wrapper
  flash/                    AES-GCM encrypted flash cookie middleware
  sse/                      SSE broker — map[formID][]chan Event, fan-out on submit
  middleware/               nonce (CSP), CSRF, auth session lookup
migrations/                 goose SQL migration files
queries/                    sqlc .sql input files
```

## Handler Pattern (LoB)

Templates and their handlers live together in `cmd/web/{feature}/`. Each feature package defines a `Handler` struct with its dependencies:

```go
// cmd/web/auth/handler.go  (package auth)
type Handler struct {
    queries *database.Queries
    mailer  email.Sender
    wa      *webauthn.WebAuthn
}
func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) { ... }
```

`internal/server/routes.go` instantiates these and registers routes. Surreal inline scripts live in `.templ` files immediately adjacent to the elements they control — never in separate JS files.

## Key Middleware (outermost → innermost)

```
methodOverride → nonce → flash → Logger → Recoverer
  public group:  → handlers
  authed group:  → authSession → csrf → handlers
```

### Method Override

HTML forms can only GET/POST. Hidden input `_method` enables DELETE/PUT from no-JS forms. Middleware reads it and overwrites `r.Method` before routing:

```html
<input type="hidden" name="_method" value="DELETE" />
```

Implemented in `internal/server/routes.go` as `methodOverride` middleware. Must be outermost so chi routes on the corrected method.

### Content Security Policy

Per-request nonce generated in middleware, stored in context. Every `<script>` tag gets `nonce={ middleware.GetNonce(ctx) }`. CSP header:

```
default-src 'self'; script-src 'nonce-{nonce}'; style-src 'self';
img-src 'self' data:; connect-src 'self'; frame-ancestors 'none';
base-uri 'none'; form-action 'self'
```

Surreal's inline `<script>` tags require nonces — never use `'unsafe-inline'`.

### Flash Messages

AES-GCM encrypted `flash` cookie. Middleware reads and immediately clears it on every request, attaches to context. Templates render from context. Handler after POST: set flash cookie → `303 See Other`.

### CSRF

- Authenticated endpoints: `csrf_token` column in `sessions` table. Hidden `<input name="_csrf">` in every form. Validated on all non-GET requests.
- Anonymous form submissions: rely on `SameSite=Strict` + `Referer` check.

## Auth Flow

```
POST /login         → create magic_link row (15min TTL) → Resend sends link → 303
GET  /auth/verify   → validate token → create session (32-byte opaque token, 30-day TTL)
                    → httpOnly + Secure + SameSite=Strict cookie → 303 /dashboard

Passkey (JS-only):
POST /auth/passkey/register/begin   → BeginRegistration → store webauthn_session → JSON
POST /auth/passkey/register/finish  → FinishRegistration → store credential
POST /auth/passkey/login/begin      → BeginDiscoverableLogin → store webauthn_session → JSON
POST /auth/passkey/login/finish     → FinishDiscoverableLogin → create session → JSON {redirect}
```

WebAuthn ceremony sessions stored in `webauthn_sessions` DB table. Cleanup: lazy on read (`WHERE expires_at > NOW()`) + hourly background goroutine sweep.

## Database

sqlc generates into `internal/database/` (configured in `sqlc.yaml`). The `Service` interface exposes `DB() *sql.DB` so `server.go` can pass it to `database.New(db)` (sqlc entry point). Key tables: `users`, `sessions`, `magic_links`, `webauthn_credentials`, `webauthn_sessions`, `forms`, `questions`, `responses`, `answers`.

Question types (enum): `short_text`, `long_text`, `multiple_choice`, `checkboxes`, `dropdown`, `linear_scale`, `date`, `time`.

## SSE

`sse.Broker` holds `map[formID][]chan Event` behind a `sync.RWMutex`. On form submit → write DB → broker publishes to all channels for that form. Graceful shutdown sends `{"type":"close"}` and drains. No-JS fallback: static count + manual refresh.

## CSS Build Pipeline

```
make vendor-assets    # curls style.css, htmx.min.js, surreal.js into cmd/web/assets/
make css              # npx --yes purgecss scans cmd/web/**/*.templ + *.go → app.css
```

PurgeCSS reads `cmd/web/assets/css/style.css`, scans all `.templ` and `.go` files for class names, outputs `cmd/web/assets/css/app.css`. The purged file is embedded into the binary at build time via `cmd/web/efs.go`.

Run `make css` after adding new Bulma classes in templates — otherwise they'll be purged.

## Build Pipeline

```bash
make vendor-assets   # download client assets
make css             # purge CSS
templ generate       # compile .templ → _templ.go (gitignored)
go build -o main cmd/api/main.go   # embeds assets/, produces binary
```

`make build` does templ generate + go build. `make build-all` does the full pipeline including assets.

## Dockerfile (3 stages)

1. `node:alpine` — curl assets + npx purgecss → `app.css`
2. `golang:alpine` — copies `app.css` from stage 1, templ generate, go build
3. `alpine` — binary only

## Progressive Enhancement

| Feature           | No JS                  | With JS                               |
| ----------------- | ---------------------- | ------------------------------------- |
| Auth              | Email magic link only  | + Passkey option                      |
| Form submit       | Full page reload       | HTMX partial swap                     |
| Validation errors | Errors on reload       | HTMX swap in; surreal clears on input |
| Modals            | New page               | HTMX into `<dialog>`                  |
| Method override   | `_method` hidden input | Same (htmx uses it too)               |
| Response feed     | Static count + refresh | SSE live updates                      |

## Local Dev

```bash
cp .env.example .env      # fill in secrets
docker compose up postgres -d
make build-all
./main
```

**Port note:** a local postgres runs on 5432. Docker postgres is mapped to 5433 to avoid the conflict. `DB_PORT=5433` in `.env`.

`make watch` for live reload via air.

## Environment Variables

See `.env.example`. Required secrets (generate with `openssl rand -hex 32`): `FLASH_SECRET`, `SESSION_SECRET`.
