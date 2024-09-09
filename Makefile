# Variables
DOCKER_COMPOSE_LOCAL = docker-compose -f docker-compose.yml -f docker-compose-local.yml
DOCKER_COMPOSE_PROD = docker-compose -f docker-compose.yml

# Run the Go application locally (with docker-compose-local.yml)
run-local:
	$(DOCKER_COMPOSE_LOCAL) up --build

# Run the Go application in production (with docker-compose.yml)
run-prod:
	$(DOCKER_COMPOSE_PROD) up --build

# Stop and remove local containers
stop-local:
	$(DOCKER_COMPOSE_LOCAL) down

# Stop and remove production containers
stop-prod:
	$(DOCKER_COMPOSE_PROD) down

# Build the Go application (local or production)
build-local:
	$(DOCKER_COMPOSE_LOCAL) build

build-prod:
	$(DOCKER_COMPOSE_PROD) build

# Run tests in the local environment
test-local:
	$(DOCKER_COMPOSE_LOCAL) run app go test ./...

# Run tests in the production environment
test-prod:
	$(DOCKER_COMPOSE_PROD) run app go test ./...

# Format the Go code
fmt:
	go fmt ./...

# Clean up the project (removes built binary)
clean:
	rm -rf bin/

# Run Go mod tidy to clean up go.mod file
tidy:
	go mod tidy

# Apply database migrations locally
migrate-local:
	$(DOCKER_COMPOSE_LOCAL) run app migrate -path ./internal/db/migrations -database ${FAKER_DATABASE_URL} up

# Apply database migrations in production
migrate-prod:
	$(DOCKER_COMPOSE_PROD) run app migrate -path ./internal/db/migrations -database ${FAKER_DATABASE_URL} up

# Rebuild the Docker containers without cache (local and production)
rebuild-local:
	$(DOCKER_COMPOSE_LOCAL) build --no-cache

rebuild-prod:
	$(DOCKER_COMPOSE_PROD) build --no-cache

# Stop and remove all local containers, volumes, and networks
clean-local:
	$(DOCKER_COMPOSE_LOCAL) down -v

# Stop and remove all production containers, volumes, and networks
clean-prod:
	$(DOCKER_COMPOSE_PROD) down -v
