
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

O `config/.env` nao vai para o git. Parta do exemplo:

```bash
cp config/.env.example config/.env
```

Depois preencha `JWT_SECRET_KEY` (o servico avisa no boot se ficar vazio
ou com o default inseguro) e `SEED_PASSWORD`, usada pelo `make seed-kyc`:

```bash
openssl rand -hex 32
```

Os valores de Postgres ja vem preenchidos no exemplo: o banco sobe junto
pelo compose local e e efemero.

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

#### Webhooks
- `POST /api/webhook/dispatch` — fire one outbound webhook on demand
- `GET /api/webhook/dispatch?project_id={id}` — list recent webhook dispatches (any origin), optionally filtered by `project_id`

`POST /api/webhook/dispatch` body:

```json
{
  "project_id": 1,
  "target_url": "https://example.com/webhook",
  "method": "POST",
  "content_type": "application/x-www-form-urlencoded",
  "headers": { "X-Signature": "{{uuid}}" },
  "body": { "event": "invoice.status_changed", "data": { "id": "{{uuid}}", "status": "paid" } }
}
```

Returns `{ status, response_headers, response_body, duration_ms, error }`. `method` defaults to `POST`, `content_type` defaults to `application/json`. Supported `content_type` values:
- `application/json` — `body` is JSON-encoded as-is
- `application/x-www-form-urlencoded` — `body` as a nested object is flattened to Iugu's `data[field]=value` style (two levels of nesting supported); `body` as a string is sent raw, unchanged

Both `headers` values and `body` fields support the same template tokens as response models (`{{uuid}}`, `{{now}}`, etc. — see `internal/api/handler/template.go`).

Outbound request timeout defaults to 10s, configurable via `WEBHOOK_TIMEOUT_SECONDS`. No automatic retry on failure or timeout — the attempt is logged either way.

**Allowed targets.** Loopback and private-network targets are blocked unless the host is listed in `WEBHOOK_ALLOWED_HOSTS` (comma-separated hostnames/IPs, default `localhost,127.0.0.1,host.docker.internal`) — list the local services you want to receive webhooks (e.g. the infrapay container name). Link-local addresses (cloud metadata, `169.254.169.254`) are always blocked. The check runs on the IP actually being connected to, so DNS rebinding cannot bypass it. Target response bodies are truncated at 1 MiB, and `POST /api/webhook/dispatch` is rate limited per account (5 req/s, burst 20).

##### Scenario-chained webhooks

A scenario endpoint (`POST /scenario`) can declare a `webhooks` array, fired asynchronously after its mocked response has already been served:

```json
{
  "path": "/invoices/pay",
  "method": "POST",
  "statuses": [...],
  "webhooks": [
    {
      "delay_ms": 500,
      "target_url": "https://example.com/webhook",
      "content_type": "application/x-www-form-urlencoded",
      "headers": { "X-Signature": "{{uuid}}" },
      "body": { "event": "invoice.status_changed", "data": { "id": "{{response.data.id}}", "cpf": "{{request.body.cpf}}" } }
    }
  ]
}
```

Templates can reach both the request that triggered the mock (`{{request.body.<field>}}`) and the mocked response just served (`{{response.<field>}}`). Endpoints without `webhooks` behave exactly as before. Every dispatch — admin-triggered or scenario-chained — is logged and visible via `GET /api/webhook/dispatch`.

## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.
