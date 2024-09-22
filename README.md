
# API Faker

API Faker is a RESTful API built with Go that helps developers mock or fake API responses. It allows for configurable responses based on different HTTP statuses and provides flexibility in testing various scenarios. The project uses PostgreSQL as the database and features dynamic configuration using environment variables. It supports both local and production environments using Docker.

## Features

- CRUD operations for Accounts and Projects
- PostgreSQL database connection
- Modular project structure
- Configurable using environment variables
- Local and production Docker configurations
- Secrets management for sensitive data using environment variables
- Simple mock endpoints for testing

## Requirements

- [Go](https://golang.org/doc/install) 1.18 or later
- [PostgreSQL](https://www.postgresql.org/download/) (running locally or on a remote server)
- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)

## Setup

### 1. Clone the Repository

Clone the repository to your local machine:

\`\`\`bash
git clone https://github.com/adolfooes/api_faker.git
cd api_faker
\`\`\`

### 2. Install Go Modules

Install the Go dependencies required for the project:

\`\`\`bash
go mod tidy
\`\`\`

This will ensure that all dependencies specified in \`go.mod\` are installed.

### 3. Setup PostgreSQL

#### Create a PostgreSQL Database

Make sure you have PostgreSQL installed and running. Create a new PostgreSQL database for the project:

\`\`\`bash
createdb api_faker_db
\`\`\`

You can replace \`api_faker_db\` with the name of your choice.

### 4. Configure Environment Variables

#### Local Development

For local development, environment variables are stored in a \`config.env\` file under the \`config/\` directory.

**Example \`config/config.env\`:**

\`\`\`env
FAKER_DATABASE_URL=postgres://postgres:password@db:5432/api_faker_dev?sslmode=disable
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=api_faker_dev
\`\`\`

### 5. Run the Project Locally with Docker

You can run the application in a local development environment using Docker Compose:

\`\`\`bash
make run-local
\`\`\`

This command will start the app and PostgreSQL database containers using the local Docker Compose configuration (\`docker-compose-local.yml\`).

To stop the containers:

\`\`\`bash
make stop-local
\`\`\`

### 6. Run the Project in Production

To run the project in a production-like environment using Docker Compose:

\`\`\`bash
make run-prod
\`\`\`

This command will use \`docker-compose.yml\` to start the app and database containers for production.

To stop the containers:

\`\`\`bash
make stop-prod
\`\`\`

### 7. Running Migrations

To apply the database migrations, use the following commands:

- For local development:

  \`\`\`bash
  make migrate-local
  \`\`\`

- For production:

  \`\`\`bash
  make migrate-prod
  \`\`\`

### 8. Running Tests

To run unit tests:

- For local development:

  \`\`\`bash
  make test-local
  \`\`\`

- For production:

  \`\`\`bash
  make test-prod
  \`\`\`

## API Endpoints

### Accounts

- \`GET /accounts\`: Retrieve all accounts
- \`GET /accounts/{id}\`: Retrieve a single account by ID
- \`POST /accounts\`: Create a new account
- \`PUT /accounts/{id}\`: Update an account by ID
- \`DELETE /accounts/{id}\`: Delete an account by ID

### Projects

- \`GET /projects\`: Retrieve all projects
- \`GET /projects/{id}\`: Retrieve a single project by ID
- \`POST /projects\`: Create a new project
- \`PUT /projects/{id}\`: Update a project by ID
- \`DELETE /projects/{id}\`: Delete a project by ID

## Testing the API

You can use \`curl\`, Postman, or any API client to test the API.

For example, to retrieve all accounts:

\`\`\`bash
curl -X GET http://localhost:8080/accounts
\`\`\`

## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.
