# Step 1: Build the Go application in a builder container
FROM golang:1.23-alpine AS builder

# Install necessary tools for building and migrations
RUN apk add --no-cache curl postgresql-client

# Set the working directory
WORKDIR /app

# Copy go mod and sum files to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application files
COPY . .

# Build the Go application
RUN go build -o api_faker ./cmd/main.go

# Step 2: Create a minimal image to run the Go app
FROM alpine:latest

# Set the working directory
WORKDIR /root/

# Install PostgreSQL client for health checks
RUN apk add --no-cache postgresql-client

# Copy the built Go application binary from the builder
COPY --from=builder /app/api_faker .

# Copy the migrations folder to the container
COPY ./internal/db/migrations /migrations

# Expose the port the application runs on
EXPOSE 8080

# Download and extract migrate binary correctly
RUN apk add --no-cache wget \
    && wget https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz \
    && tar -xzf migrate.linux-amd64.tar.gz \
    && mv migrate /usr/local/bin/migrate \
    && chmod +x /usr/local/bin/migrate

# Set the default command to run the migrations first, then start the app
CMD echo "Waiting for database to be ready..."; \
    until pg_isready -h db -p 5432 -U postgres; do \
        echo "Waiting for PostgreSQL to start..."; \
        sleep 2; \
    done; \
    echo "Creating database api_faker_dev if it doesn't exist..."; \
    PGPASSWORD=dev123 psql -h db -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = 'api_faker_dev'" | grep -q 1 || \
    PGPASSWORD=dev123 psql -h db -U postgres -c "CREATE DATABASE api_faker_dev;" || { echo "Database creation failed"; exit 1; }; \
    echo "Database is ready! Running migrations..."; \
    migrate -path /migrations -database postgres://postgres:dev123@db:5432/api_faker_dev?sslmode=disable up || { echo "Migrations failed"; exit 1; }; \
    echo "Migrations completed. Starting the app..."; \
    ./api_faker
