APP     := oz-api
BIN     := ./bin/$(APP)
MAIN    := ./main.go
PKG     := ./...
DB_NAME ?= oz_db
DB_USER ?= oz
DB_ADDR ?= localhost:5432

# ─── Build ────────────────────────────────────────────────────────────────────

.PHONY: build
build:
	@mkdir -p bin
	go build -o $(BIN) $(MAIN)

.PHONY: run
run:
	go run $(MAIN)

.PHONY: watch
watch:
	@which air > /dev/null || go install github.com/air-verse/air@latest
	air

# ─── Tests ────────────────────────────────────────────────────────────────────

.PHONY: test
test:
	go test $(PKG) -v

.PHONY: test-cover
test-cover:
	go test $(PKG) -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Report: coverage.html"

# ─── Code quality ─────────────────────────────────────────────────────────────

.PHONY: lint
lint:
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run $(PKG)

.PHONY: vet
vet:
	go vet $(PKG)

.PHONY: fmt
fmt:
	gofmt -w .

.PHONY: tidy
tidy:
	go mod tidy

# ─── Swagger ──────────────────────────────────────────────────────────────────

.PHONY: swagger
swagger:
	@which swag > /dev/null || go install github.com/swaggo/swag/cmd/swag@latest
	swag init -g $(MAIN) -o ./docs

# ─── Docker ───────────────────────────────────────────────────────────────────

.PHONY: up
up:
	docker compose up --build -d

.PHONY: down
down:
	docker compose down

.PHONY: down-v
down-v:
	docker compose down -v

.PHONY: logs
logs:
	docker compose logs -f api

.PHONY: ps
ps:
	docker compose ps

# ─── Database ─────────────────────────────────────────────────────────────────

.PHONY: db-shell
db-shell:
	docker compose exec db psql -U $(DB_USER) -d $(DB_NAME)

.PHONY: db-migrate
db-migrate:
	docker compose exec -T db psql -U $(DB_USER) -d $(DB_NAME) < zorodb.sql

.PHONY: db-migrate-up
db-migrate-up:
	@for f in migrations/*.sql; do \
		echo "Applying $$f..."; \
		docker compose exec -T db psql -U $(DB_USER) -d $(DB_NAME) < $$f; \
	done

.PHONY: db-reset
db-reset: down-v up
	@echo "Waiting for DB to be ready..."
	@sleep 3
	$(MAKE) db-migrate

# ─── Helpers ──────────────────────────────────────────────────────────────────

.PHONY: clean
clean:
	rm -rf bin coverage.out coverage.html

.PHONY: env
env:
	@cp -n .env.example .env 2>/dev/null && echo ".env créé depuis .env.example" || echo ".env existe déjà"

.PHONY: help
help:
	@echo ""
	@echo "  build        Compile le binaire → bin/$(APP)"
	@echo "  run          Lance le serveur en local"
	@echo "  watch        Hot-reload avec air"
	@echo ""
	@echo "  test         Lance les tests"
	@echo "  test-cover   Tests + rapport de couverture HTML"
	@echo ""
	@echo "  lint         golangci-lint"
	@echo "  vet          go vet"
	@echo "  fmt          gofmt"
	@echo "  tidy         go mod tidy"
	@echo ""
	@echo "  swagger      Génère la doc Swagger (docs/)"
	@echo ""
	@echo "  up           docker compose up --build -d"
	@echo "  down         docker compose down"
	@echo "  down-v       docker compose down -v (supprime les volumes)"
	@echo "  logs         Logs du conteneur api"
	@echo "  ps           État des conteneurs"
	@echo ""
	@echo "  db-shell     psql interactif dans le conteneur"
	@echo "  db-migrate   Applique zorodb.sql dans le conteneur"
	@echo "  db-reset     Repart de zéro (down-v + up + migrate)"
	@echo ""
	@echo "  clean        Supprime bin/ et les rapports"
	@echo "  env          Crée .env depuis .env.example (si absent)"
	@echo ""
