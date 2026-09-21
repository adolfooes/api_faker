# Define the docker-compose file for local environment
DOCKER_COMPOSE_LOCAL = docker-compose-local.yml

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

# Default target if no target is provided
.PHONY: swagger up down restart logs migrate shell

# ---- Go via container ----
# O toolchain Go nao esta instalado na maquina; usa a MESMA imagem do
# builder do Dockerfile para que build e teste rodem contra a versao que
# de fato vai a producao.
GO_CONTAINER = docker run --rm -v "$(PWD)":/app -w /app -e GOFLAGS=-buildvcs=false golang:1.23-alpine

build:
	$(GO_CONTAINER) go build ./...

test:
	$(GO_CONTAINER) go test ./...

.PHONY: build test
