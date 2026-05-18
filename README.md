# Formful

Google Forms clone built with Go, htmx, and PostgreSQL.

## Setup

```bash
cp .env.example .env       # fill in DB_PASSWORD, FLASH_SECRET, SESSION_SECRET, RESEND_API_KEY
docker compose up postgres -d
make build-all             # vendor assets, purge CSS, generate templ, build binary
./main
```

Generate secrets with `openssl rand -hex 32`.

**Port note:** if you have a local postgres on 5432, the Docker instance is mapped to 5433. `DB_PORT=5433` is the default in `.env.example`.

## Make targets

| Target               | Description                                                            |
| -------------------- | ---------------------------------------------------------------------- |
| `make build-all`     | Full build: vendor assets → purge CSS → templ generate → go build      |
| `make build`         | templ generate + go build (assets must already exist)                  |
| `make run`           | `go run` without rebuilding assets                                     |
| `make watch`         | Live reload via air                                                    |
| `make test`          | Unit + integration tests                                               |
| `make itest`         | Integration tests only (spins up Postgres via testcontainers)          |
| `make vendor-assets` | Curl htmx, surreal.js, and style.css into `cmd/web/assets/`            |
| `make css`           | Run PurgeCSS on style.css → app.css (run after adding new CSS classes) |
| `make sqlc-gen`      | Regenerate sqlc query code from `queries/`                             |
| `make migrate-up`    | Run pending goose migrations (`DATABASE_URL` must be set)              |
| `make migrate-down`  | Roll back last goose migration                                         |
| `make clean`         | Remove binary and vendored assets                                      |
| `make docker-run`    | Build and start all services via docker compose                        |
| `make docker-down`   | Stop docker compose services                                           |
| `make spell`         | Spell-check all files via cspell                                       |
