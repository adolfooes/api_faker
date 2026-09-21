
# API Faker

API Faker is a RESTful API built with Go that helps developers mock or fake API responses. It allows for configurable responses based on different HTTP statuses and provides flexibility in testing various scenarios. The project uses PostgreSQL as the database and features dynamic configuration using environment variables. It supports both local and production environments using Docker.

## Features

- CRUD operations for Accounts, Projects, URL Configs, HTTP Statuses and Response Models
- Public mock endpoints — no authentication required to call `/api/mock/:project_id/:path`
- PostgreSQL database with automatic migrations on startup
- Modular project structure
- Configurable using environment variables
- Docker-based local and production environments
- JWT authentication for administrative routes

## Requirements

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)
- Go is **not** required locally — `make build` and `make test` run inside a container

## Setup

### 1. Clone the Repository

```bash
git clone https://github.com/adolfooes/api_faker.git
cd api_faker
```

### 2. Configure Environment Variables

Copy or edit `config/.env`:

```env
FAKER_DATABASE_URL=postgres://postgres:password@db:5432/api_faker_dev?sslmode=disable
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=api_faker_dev
JWT_SECRET_KEY=change_me_to_a_secure_random_value
```

> **Warning:** If `JWT_SECRET_KEY` is empty or set to `your_secret_key`, the server will print a warning on startup.

### 3. Start the Project

```bash
make up
```

This builds and starts the app and PostgreSQL containers.

## Makefile Targets

| Target | Description |
|---|---|
| `make up` | Build and start all containers |
| `make down` | Stop and remove containers |
| `make restart` | Stop then start (equivalent to `down` + `up`) |
| `make logs` | Tail container logs |
| `make migrate` | Run pending database migrations inside the app container |
| `make shell` | Open a shell inside the app container |
| `make build` | Compile the project (runs inside a Go container — no local Go needed) |
| `make test` | Run all tests (runs inside a Go container — no local Go needed) |
| `make seed-kyc` | Seed KYC test data into the database |

## API Endpoints

### Public

- `POST /login` — authenticate and receive a JWT token
- `POST /account` — create a new account
- `GET|POST|PUT|DELETE|PATCH /api/mock/{project_id}/{path}` — serve a mock response for the configured path

### Protected (requires `Authorization: Bearer <token>`)

#### Accounts
- `GET /api/account/{id}` — get account
- `PUT /api/account/{id}` — update account
- `DELETE /api/account/{id}` — delete account

#### Projects
- `GET /api/project` — list all projects
- `GET /api/project/{id}` — get project
- `POST /api/project` — create project
- `PUT /api/project/{id}` — update project
- `DELETE /api/project/{id}` — delete project

#### URL Configs
- `GET /api/url_config` — list all URL configs
- `GET /api/url_config/{id}` — get URL config
- `POST /api/url_config` — create URL config
- `PUT /api/url_config/{id}` — update URL config
- `DELETE /api/url_config/{id}` — delete URL config

#### URL HTTP Statuses
- `GET /api/url_http_status` — list all HTTP statuses
- `POST /api/url_http_status` — create HTTP status
- `PUT /api/url_http_status/{id}` — update HTTP status
- `DELETE /api/url_http_status/{id}` — delete HTTP status

#### Response Models
- `GET /api/response_model` — list all response models
- `POST /api/response_model` — create response model
- `PUT /api/response_model/{id}` — update response model
- `DELETE /api/response_model/{id}` — delete response model

## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.
