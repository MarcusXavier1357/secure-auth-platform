# Sistema de Autenticação e Autorização

Implementação do plano descrito em [plano.md](plano.md): autenticação com JWT + refresh token, permissões granulares por usuário, cache e rate limiting em Redis e auditoria em PostgreSQL.

Documentação técnica detalhada em [docs/](docs/README.md) — backend, banco de dados, Redis, frontend e infra.

## Stack

| Camada | Tecnologias |
|--------|-------------|
| Frontend | React 19, Vite, Tailwind CSS 4, TanStack Query, React Router |
| Backend | Go, Fiber, Bun (ORM), go-redis, golang-migrate, JWT, bcrypt |
| Infra | Docker Compose, Nginx, PostgreSQL 16, Redis 7 |

## Clonar em outra máquina

Pré-requisitos: [Git](https://git-scm.com/) e [Docker Desktop](https://www.docker.com/products/docker-desktop/) (ou Docker Engine + Compose).

```bash
git clone https://github.com/MarcusXavier1357/secure-auth-platform.git
cd secure-auth-platform/infra
```

Copie o arquivo de ambiente e ajuste os valores:

```bash
# Windows (PowerShell / CMD)
copy env.example .env

# Linux / macOS
cp env.example .env
```

Defina ao menos `ADMIN_PASSWORD` no `.env` (obrigatório, sem default) e suba o stack (JWT RS256: chaves geradas na imagem Docker):

```bash
docker compose up -d --build
```

A aplicação fica em `http://localhost`. Login padrão após o primeiro boot: `ADMIN_EMAIL` / `ADMIN_PASSWORD` do `.env` (default `admin@local.dev`).

## Como rodar (Docker)

Se o repositório já está na sua máquina, siga os mesmos passos da seção [Clonar em outra máquina](#clonar-em-outra-máquina) a partir de `infra/`.

A API responde em `http://localhost/api`. No primeiro boot, as migrations criam o schema, populam roles e permissões, e a API cria o usuário admin com todas as permissões.

## Desenvolvimento local (sem Docker para app)

Suba apenas Postgres e Redis:

```bash
cd infra
docker compose up -d postgres redis
```

Backend (precisa das envs `DATABASE_URL`, chaves JWT RS256, `ADMIN_PASSWORD`):

```bash
cd backend
go run ./cmd/api
```

Frontend (proxy de `/api` para `localhost:8080` já configurado no Vite):

```bash
cd frontend
npm install
npm run dev
```

## Testes

A suite de testes ponta a ponta (`backend/tests/`) exercita a API completa — handlers, middlewares, services, Postgres e Redis reais — cobrindo login, refresh com rotação, logout, rate limit, autorização (401/403), grant/revoke com invalidação de cache, desativação de usuário e auditoria.

Pré-requisito: Postgres e Redis do compose rodando. A suite cria um banco isolado `auth_test` (recriado a cada execução) e usa o Redis db 1 — não toca nos dados de desenvolvimento.

```bash
cd infra
docker compose up -d postgres redis

cd ../backend
go test ./tests/ -v
```

Conexões padrão dos testes (sobrescreva com `TEST_PG_ADMIN_URL`, `TEST_PG_URL` e `TEST_REDIS_URL` se necessário): Postgres em `127.0.0.1:55432` e Redis em `127.0.0.1:6379`.

Nota: o Postgres do compose publica no host a porta `55432` (não `5432`) para evitar conflito com instalações nativas do PostgreSQL no Windows. Dentro da rede Docker a API continua usando `postgres:5432`.

## Endpoints principais

| Método | Rota | Proteção |
|--------|------|----------|
| POST | `/auth/login` | Rate limit (5/15min por IP e email) |
| POST | `/auth/refresh` | Cookie HttpOnly (rotação de token) |
| POST | `/auth/logout` | Cookie HttpOnly |
| GET | `/me` | Autenticado |
| GET/POST | `/users` | `users.manage` |
| GET/PATCH | `/users/:id` | `users.manage` |
| GET | `/permissions` | `permissions.manage` |
| POST | `/users/:id/permissions` | `permissions.manage` |
| DELETE | `/users/:id/permissions/:permissionId` | `permissions.manage` |
| GET | `/audit-logs` | `audit_logs.read` |

## Decisões de segurança

- Access token (15 min) vive apenas em memória no frontend; refresh token (30 dias) em cookie `HttpOnly` + `SameSite=Strict`, com apenas o hash SHA-256 persistido no banco.
- Permissões nunca entram no JWT — são consultadas via cache Redis (TTL 5 min) com fallback para PostgreSQL.
- Se o Redis cair durante o login, a API responde `503` em vez de pular o rate limit.
- Login com bcrypt dummy nos caminhos de falha sem usuário — iguala o tempo de resposta e impede enumeração de emails por timing.
- Error handler global: erros internos nunca vazam na resposta HTTP (`500` genérico + log estruturado).
- Usuário não pode desativar a própria conta (previne lockout do último admin).
- Toda operação crítica (login, logout, CRUD de usuários, grant/revoke) gera registro em `audit_logs`; a escrita usa contexto não-cancelável.
- API roda como usuário não-root no container, com graceful shutdown em SIGTERM e limpeza periódica de sessões expiradas.
- Em produção, configure TLS no Nginx (`infra/nginx/nginx.conf`) e `COOKIE_SECURE=true`.

## Qualidade

```bash
# Backend
cd backend
go vet ./...
golangci-lint run ./...   # config em .golangci.yml
go test ./tests/ -v

# Frontend
cd frontend
npm run lint
npm run format:check
```
