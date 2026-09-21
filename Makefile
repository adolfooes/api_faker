# Define the docker-compose file for local environment
DOCKER_COMPOSE_LOCAL = docker-compose-local.yml

TEST_DB_ENV = $(if $(TEST_DATABASE_URL),-e TEST_DATABASE_URL="$(TEST_DATABASE_URL)",)
GO_IMAGE = golang:1.23-alpine
DOCKER_GO = docker run --rm --network host -v "$(PWD)":/app -w /app -e GOFLAGS=-buildvcs=false
GO_CONTAINER = $(DOCKER_GO) $(GO_IMAGE)

# Target to start the services (build and run)
up:
	@echo "Starting local services..."
	docker-compose --env-file ./config/.env -f docker-compose-local.yml up --build

# Target to stop the services
down:
	@echo "Stopping local services..."
	docker-compose -f $(DOCKER_COMPOSE_LOCAL) down

# Target to rebuild and restart the services
restart: down up

# Target to view logs
logs:
	@echo "Displaying logs..."
	docker-compose -f $(DOCKER_COMPOSE_LOCAL) logs -f

# Target to run migrations
migrate:
	@echo "Running database migrations..."
	docker-compose -f $(DOCKER_COMPOSE_LOCAL) exec app migrate -path /migrations -database $$FAKER_DATABASE_URL up

# Target to open shell into the app container
shell:
	@echo "Opening shell in the app container..."
	docker-compose -f $(DOCKER_COMPOSE_LOCAL) exec app sh

# Build the Go binary (no local Go installation required)
build:
	$(GO_CONTAINER) go build ./...

# Run tests (no local Go installation required)
# Roda com -v e conta os PULADOS. `go test` imprime "ok" mesmo quando um
# teste foi pulado, e o skip so aparece com -v: suite verde escondendo teste
# que nunca executou e exatamente o que nao se quer descobrir tarde.
# TEST_DATABASE_URL habilita os testes de integracao; sem ela eles pulam, e o
# aviso diz quantos e quais.
test:
	@$(DOCKER_GO) $(TEST_DB_ENV) $(GO_IMAGE) go test ./... -v > .test-output.log 2>&1; \
	status=$$?; \
	grep -E "^(ok|FAIL|--- FAIL)" .test-output.log || true; \
	passed=`grep -c '^--- PASS' .test-output.log || true`; \
	skipped=`grep -c '^--- SKIP' .test-output.log || true`; \
	echo "passaram: $$passed | pulados: $$skipped"; \
	if [ "$$skipped" -gt 0 ]; then \
		echo "AVISO: $$skipped teste(s) NAO executaram:"; \
		grep '^--- SKIP' .test-output.log | sed 's/^/  /'; \
		echo "  defina TEST_DATABASE_URL para roda-los"; \
	fi; \
	rm -f .test-output.log; \
	exit $$status

# Seed KYC test data (requires stack running via make up and local Go installation)
seed-kyc:
	@echo "Seeding KYC scenario..."
	go run ./scripts/seed_kyc/main.go

.PHONY: up down restart logs migrate shell build test seed-kyc
