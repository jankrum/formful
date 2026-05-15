all: build test

# ── tools ─────────────────────────────────────────────────────────────────────

templ-install:
	@if ! command -v templ > /dev/null; then \
		read -p "templ not installed. Install? [Y/n] " c; \
		if [ "$$c" != "n" ] && [ "$$c" != "N" ]; then \
			go install github.com/a-h/templ/cmd/templ@latest; \
		else \
			echo "templ required. Exiting."; exit 1; \
		fi; \
	fi

sqlc-install:
	@if ! command -v sqlc > /dev/null; then \
		read -p "sqlc not installed. Install? [Y/n] " c; \
		if [ "$$c" != "n" ] && [ "$$c" != "N" ]; then \
			go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest; \
		else \
			echo "sqlc required. Exiting."; exit 1; \
		fi; \
	fi

goose-install:
	@if ! command -v goose > /dev/null; then \
		read -p "goose not installed. Install? [Y/n] " c; \
		if [ "$$c" != "n" ] && [ "$$c" != "N" ]; then \
			go install github.com/pressly/goose/v3/cmd/goose@latest; \
		else \
			echo "goose required. Exiting."; exit 1; \
		fi; \
	fi

# ── client assets ─────────────────────────────────────────────────────────────

vendor-assets:
	@mkdir -p cmd/web/assets/js cmd/web/assets/css
	curl -sLo cmd/web/assets/js/htmx.min.js https://unpkg.com/htmx.org/dist/htmx.min.js
	curl -sLo cmd/web/assets/js/surreal.js https://cdn.jsdelivr.net/gh/gnat/surreal/surreal.js
	curl -sLo cmd/web/assets/css/style.css https://cdn.jsdelivr.net/npm/@jankrum/style.css/dist/style.css

css:
	npx --yes purgecss \
		--css cmd/web/assets/css/style.css \
		--content "cmd/web/**/*.templ" "cmd/web/**/*.go" \
		--output cmd/web/assets/css/app.css

# ── database ──────────────────────────────────────────────────────────────────

migrate-up: goose-install
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down: goose-install
	goose -dir migrations postgres "$(DATABASE_URL)" down

sqlc-gen: sqlc-install
	sqlc generate

# ── build ─────────────────────────────────────────────────────────────────────

build: templ-install
	@templ generate
	@go build -o main cmd/api/main.go

build-all: vendor-assets css sqlc-gen build

run:
	@go run cmd/api/main.go

docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

spell:
	pnpx cspell "**" ".air.toml"

# ── test ──────────────────────────────────────────────────────────────────────

test:
	@echo "Testing..."
	@go test ./... -v

itest:
	@echo "Running integration tests..."
	@go test ./internal/database -v

# ── dev ───────────────────────────────────────────────────────────────────────

watch:
	@if command -v air > /dev/null; then \
		air; \
	else \
		read -p "air not installed. Install? [Y/n] " c; \
		if [ "$$c" != "n" ] && [ "$$c" != "N" ]; then \
			go install github.com/air-verse/air@latest; \
			air; \
		else \
			echo "air required for watch. Exiting."; exit 1; \
		fi; \
	fi

clean:
	@echo "Cleaning..."
	@rm -f main
	@rm -f cmd/web/assets/js/htmx.min.js
	@rm -f cmd/web/assets/js/surreal.js
	@rm -f cmd/web/assets/css/style.css
	@rm -f cmd/web/assets/css/app.css

.PHONY: all build build-all run test clean watch docker-run docker-down itest \
	templ-install sqlc-install goose-install vendor-assets css sqlc-gen \
	migrate-up migrate-down spell
