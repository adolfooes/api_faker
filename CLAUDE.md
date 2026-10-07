# CLAUDE.md — api_faker

## Tipo de Projeto
**Tipo:** backend

## Stack
- **Linguagem:** Go 1.23.1
- **Framework:** gorilla/mux v1.8.1
- **Banco de dados:** PostgreSQL 16
- **Gerenciador de pacotes:** Go modules (`go mod`)
- **Auth:** golang-jwt/v5 (JWT Bearer)
- **Migrations:** golang-migrate/v4
- **Password hashing:** bcrypt (golang.org/x/crypto)

## Estrutura de Diretórios
```
api_faker/
├── cmd/
│   └── main.go                  # Entry point — inicia DB, migrations e servidor HTTP :8080
├── config/
│   ├── .env / .env.example      # FAKER_DATABASE_URL, JWT_SECRET_KEY, POSTGRES_*, WEBHOOK_*, SEED_*
│   ├── config.go                # Lê env vars de DB e JWT
│   └── constants.go             # ContextKey para injetar account_id via JWT no contexto HTTP
├── internal/
│   ├── api/
│   │   ├── handler/             # Um arquivo por recurso: account, project, url_config, url_http_status,
│   │   │                        # response_model, mock, scenario, template, login, webhook (+ dispatch/encode/
│   │   │                        # persistence/security), cada um com *_test.go companheiro
│   │   ├── middleware/
│   │   │   └── middleware.go    # JWTMiddleware — valida Bearer token e injeta account_id no contexto
│   │   └── router/
│   │       └── router.go        # Monta rotas públicas (/login, /account POST, /api/mock/*) e protegidas (/api/*)
│   └── db/
│       ├── db.go                # InitDB, RunMigrations, GetDB
│       ├── enums.go             # TranslateEnumValue — converte []byte de colunas ENUM do Postgres
│       └── migrations/          # SQL migrations numeradas (000001 a 000004)
├── pkg/
│   └── utils/
│       ├── crud/
│       │   └── crud.go          # CRUD genérico sobre qualquer tabela (Create, Read, List, Update, Delete, Raw)
│       └── response/
│           └── response.go      # SendResponse — serialização padronizada de respostas HTTP
├── scripts/
│   ├── fetch_secrets.sh         # Busca segredos (uso em CI/prod)
│   ├── kyc-scenario.json        # Cenário de mock: fluxo KYC (iugu-kyc)
│   ├── partner1-scenario.json   # Cenário de mock: endpoints partner1 (infrapay e2e)
│   └── seed_kyc/main.go         # Seed de dados de teste KYC no banco
├── app/                         # Vazio — sem uso atual
├── Dockerfile
├── docker-compose.yml           # Produção
├── docker-compose-local.yml     # Desenvolvimento local
└── Makefile                     # Targets: up, down, restart, logs, migrate, shell, build, test, seed-kyc
```

## Módulos Principais

| Módulo | Responsabilidade | Arquivos-chave |
|---|---|---|
| Router | Monta todas as rotas, aplica JWTMiddleware nas rotas `/api/*` | `internal/api/router/router.go` |
| Handlers | Um handler por recurso — decodifica body, valida, chama crud, devolve response | `internal/api/handler/*.go` |
| Middleware | Valida JWT Bearer e injeta `account_id` no context HTTP | `internal/api/middleware/middleware.go` |
| CRUD genérico | Executa SQL dinâmico sobre qualquer tabela usando `RETURNING *` | `pkg/utils/crud/crud.go` |
| DB | Singleton `*sql.DB`, inicialização, migrations automáticas na subida | `internal/db/db.go` |
| Mock engine | Recebe `GET/POST/PUT/DELETE/PATCH /api/mock/{project_id}/{path}`, sorteia HTTP status por percentual e retorna response_model configurado | `internal/api/handler/mock.go` |
| Scenario | Cria/lê cenários completos de mock (project + url_config + response_model em lote) a partir de JSON | `internal/api/handler/scenario.go` |
| Template | Suporte a templating de responses | `internal/api/handler/template.go` |
| Webhook | Dispara, assina, persiste e lista webhooks de saída configurados por cenário | `internal/api/handler/webhook*.go` |

## Comandos
```bash
# Instalar dependências
go mod tidy

# Rodar localmente com Docker (build + up)
make up

# Parar containers / reiniciar
make down
make restart

# Logs / shell no container
make logs
make shell

# Aplicar migrations (dentro do container)
make migrate

# Build e test (rodam dentro de container Go — não precisa Go local)
make build
make test

# Seed de dados KYC
make seed-kyc

# Build/run direto (requer Go e Postgres locais)
go build ./cmd/main.go
FAKER_DATABASE_URL="postgres://user:pass@localhost:5432/api_faker_dev?sslmode=disable" \
JWT_SECRET_KEY="secret" \
go run ./cmd/main.go
```

## Repos Relacionados
Nenhum — projeto standalone (API pura, sem frontend separado).

## Convenções
- Conventional Commits: `feat:`, `fix:`, `chore:`, `feat(scope):`, `fix(scope):`
- Todas as respostas HTTP passam por `response.SendResponse(w, statusCode, message, errDetail, data, isRaw)`
- CRUD genérico via `pkg/utils/crud` — handlers não escrevem SQL diretamente (exceção: `crud.Raw` para queries ad-hoc)
- Enum PostgreSQL traduzido em runtime em `internal/db/enums.go`; adicionar novo ENUM = editar `enumMappings`
- Rotas protegidas ficam sob prefixo `/api` com `JWTMiddleware`; rota de mock (`/api/mock/*`) é pública
- `account_id` do JWT é injetado via `context.WithValue` usando `config.JWTAccountIDKey`
- Variáveis de ambiente obrigatórias: `FAKER_DATABASE_URL`, `JWT_SECRET_KEY`; opcionais: `WEBHOOK_TIMEOUT_SECONDS`, `WEBHOOK_ALLOWED_HOSTS`, `SEED_EMAIL`, `SEED_PASSWORD`
- Cada handler de recurso tem `*_test.go` companheiro — seguir esse padrão para handlers novos
- Cenários de mock completos (projeto + rotas + responses) ficam versionados em `scripts/*-scenario.json`

## Escopo dos Agentes SDLC
- **@pm:** `.sdlc/specs/`, `.sdlc/knowledge/`
- **@backend-eng:** `cmd/`, `internal/`, `pkg/`, `config/`
- **@qa:** lê todos, escreve em `.sdlc/reports/`
- **@security:** lê todos, escreve em `.sdlc/reports/`
- **@devops:** `scripts/`, `Dockerfile`, `docker-compose*.yml`, `Makefile`
