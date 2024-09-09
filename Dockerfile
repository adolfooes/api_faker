# Step 1: Build the Go application in a builder container
FROM golang:1.18-alpine AS builder

# Set environment variables
ENV GO111MODULE=on

# Install golang-migrate for database migrations
RUN apk add --no-cache curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate.linux-amd64 /usr/local/bin/migrate

# Set the current working directory inside the container
WORKDIR /app

# Copy go mod and sum files first to cache dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application files
COPY . .

# Build the Go application
RUN go build -o api_faker ./cmd/main.go

# Step 2: Create a minimal final image to run the Go app
FROM alpine:latest

# Set the working directory
WORKDIR /root/

# Copy the Go application binary from the builder container
COPY --from=builder /app/api_faker .
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate

# Copy the migrations folder to the container
COPY ./internal/db/migrations /migrations

# Expose the port the application runs on
EXPOSE 8080

# Run the database migrations and then the application
CMD migrate -path /migrations -database $FAKER_DATABASE_URL up && ./api_faker
