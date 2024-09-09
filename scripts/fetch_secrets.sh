#!/bin/bash

# Fetch secrets from Google Cloud Secret Manager
export FAKER_DATABASE_URL=$(gcloud secrets versions access latest --secret="faker-database-url-secret")
export POSTGRES_USER=$(gcloud secrets versions access latest --secret="postgres-user-secret")
export POSTGRES_PASSWORD=$(gcloud secrets versions access latest --secret="postgres-password-secret")

# Validate that secrets were retrieved successfully
if [ -z "$FAKER_DATABASE_URL" ] || [ -z "$POSTGRES_USER" ] || [ -z "$POSTGRES_PASSWORD" ]; then
    echo "Error: One or more secrets could not be retrieved from Google Cloud Secret Manager."
    exit 1
fi

# Display the fetched secrets (for debugging, remove this in production)
echo "FAKER_DATABASE_URL=$FAKER_DATABASE_URL"
echo "POSTGRES_USER=$POSTGRES_USER"
echo "POSTGRES_PASSWORD=$POSTGRES_PASSWORD"

# Run Docker Compose with the environment variables
docker-compose up
