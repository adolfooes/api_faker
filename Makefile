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
.PHONY: up down restart logs migrate shell
