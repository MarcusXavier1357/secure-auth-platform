# AGENTS.md — Padrões do Projeto

Sistema de autenticação e autorização. Especificação completa em [plano.md](plano.md). Leia este arquivo antes de alterar código.

## Stack

- **Backend**: Go + Fiber, Bun (ORM), go-redis, golang-migrate, JWT, bcrypt
- **Frontend**: React 19 + Vite, Tailwind 4, TanStack Query, React Router
- **Infra**: Docker Compose (`infra/`), Nginx, PostgreSQL 16, Redis 7

## Estrutura

```
backend/
  cmd/api/           # main.go: apenas leitura de env + bootstrap. Sem lógica.
  internal/app/      # montagem da aplicação (wiring). Usado pelo main E pelos testes.
  internal/models/   # structs Bun (espelham as tabelas, nunca o contrário)
  internal/repository/  # acesso a dados. Única camada que toca o bun.DB.
  internal/service/  # regras de negócio. Recebe repositórios, expõe erros sentinela.
  internal/{auth,users,permissions,audit}/  # handlers HTTP (1 pacote por recurso)
  cache/             # client Redis, permission cache, rate limiter
  middleware/        # Auth (JWT) e RequirePermission
  routes/            # registro de rotas + injeção de dependências
  migrations/        # SQL versionado (golang-migrate)
  tests/             # testes ponta a ponta (TODOS os testes ficam aqui)
frontend/src/
  router/            # index.tsx (monta <Routes>) + paths.ts (fonte única de paths)
  app/               # convenção App Router: (public)/ e (protected)/ com layout.tsx + page.tsx
  services/api.ts    # único ponto de fetch. Token em memória + retry em 401.
  providers/         # AuthProvider (user, permissions, login/logout)
  hooks/ components/
infra/               # docker-compose.yml, nginx, .env (gitignored)
```

## Backend — regras de camadas

Fluxo obrigatório: `handler → service → repository`. Nunca pule camadas.

- **Handlers** só fazem: parse/validação de input, chamada de service, mapeamento de erro para status HTTP. Zero SQL, zero regra de negócio.
- **Services** contêm regras de negócio e auditoria. Retornam erros sentinela (`var ErrInvalidCredentials = errors.New(...)`), nunca `fiber.Error`.
- **Repositories** contêm todas as queries. Convertem `sql.ErrNoRows` em `repository.ErrNotFound`.
- Erros não mapeados sobem para o `ErrorHandler` global (`internal/app/app.go`), que responde `500` genérico sem vazar detalhes internos. Nunca coloque `err.Error()` de erro interno na resposta HTTP.
- Handlers traduzem sentinelas com `errors.Is`:

```go
if errors.Is(err, service.ErrInvalidCredentials) {
    return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
}
```

## Banco de dados

- Schema muda **apenas** via migration nova em `backend/migrations/` (`NNNNNN_nome.up.sql` + `.down.sql`). Nunca edite migrations já aplicadas. Nunca use AutoMigrate.
- Após criar migration, atualize a struct correspondente em `internal/models/`.
- Queries sempre parametrizadas (Bun `Where("x = ?", v)` ou placeholders em raw). Concatenação de strings em SQL é proibida.

## Segurança (inegociável)

- Frontend é não-confiável: toda autorização é revalidada no backend via `RequirePermission`.
- Permissões **nunca** entram no payload do JWT — sempre Redis/Postgres.
- Senhas: somente bcrypt. Refresh tokens: somente hash SHA-256 no banco, valor puro só no cookie `HttpOnly` + `SameSite=Strict`.
- Access token no frontend vive **apenas em memória** — nunca localStorage/sessionStorage.
- Rate limit roda **antes** do bcrypt no login. Redis indisponível no login → `503` (nunca pular o rate limit).
- Caminhos de falha do login sem usuário executam bcrypt dummy (`dummyPasswordHash` em `auth.go`) para igualar timing e impedir enumeração de emails. Mantenha esse padrão em qualquer novo caminho de falha.
- Usuário não pode desativar a própria conta (`ErrCannotDeactivateSelf` → `409`) — previne lockout do último admin.
- Secrets só via env vars (`JWT_SECRET`, `ADMIN_PASSWORD` não têm default). Nunca hardcode.
- Containers rodam como usuário não-root (`USER app` no Dockerfile do backend).

## Auditoria e cache

- Toda mutação crítica (login, logout, CRUD de usuário, grant/revoke) registra em `audit_logs` via `AuditService.Log` com `old_data`/`new_data`.
- A escrita de audit usa `context.WithoutCancel` — a trilha é gravada mesmo se o cliente desconectar. Falha de audit não derruba a operação, mas gera `slog.Error`.
- Ao alterar permissões ou desativar usuário: **sempre** invalidar `permissions:user:{id}` no Redis. Esquecer a invalidação é bug de segurança.
- Desativar usuário também revoga todas as sessões (`RevokeAllByUser`).

## Testes

- Todos os testes ficam em `backend/tests/` (pacote `tests`), estilo ponta a ponta: app real via `app.New` + `Fiber.Test()`, Postgres e Redis reais.
- Pré-requisito: `docker compose up -d postgres redis` (Postgres publica em `127.0.0.1:55432` no host).
- A suite recria o banco `auth_test` e usa Redis db 1 — nunca aponte testes para o banco de dev.
- Rodar: `go test ./tests/ -v` em `backend/`.
- Novos testes devem ser **autossuficientes**: criem seus próprios usuários/dados e não dependam da ordem de execução de outros testes.
- Use os helpers existentes (`newClient`, `mustLogin`, `requireStatus`, `decodeJSON`). Testes que precisam de rate limit baixo usam `newRateLimitClient`.
- Toda feature nova de auth/autorização exige teste cobrindo o caso de sucesso **e** o de negação (401/403).

## Frontend

- Toda chamada HTTP passa por `services/api.ts` (`request()`), que renova o token em 401 automaticamente. Não use `fetch` direto em componentes.
- **Rotas seguem convenção App Router**: página nova = `src/app/(protected)/<segmento>/page.tsx` (export default) registrada em `src/router/index.tsx`. Route groups `(public)`/`(protected)` não entram na URL. Layouts compartilhados são `layout.tsx` com `<Outlet />`.
- Links e navegação **sempre** via `router/paths.ts` (`paths.users()`), nunca strings soltas.
- **Guard duplo obrigatório** em rota com permissão: `RequirePermission` protege a rota (redireciona) e `Can` esconde o link/card. Ambos cosméticos — o backend decide.
- Páginas admin entram com `lazy(() => import(...))` em `router/index.tsx`; o `Suspense` já existe no `ProtectedLayout`.
- Dados de servidor nas páginas: TanStack Query (`useQuery`/`useMutation` + invalidação). Estado de sessão fica no `AuthProvider`.
- Componentes funcionais, hooks para lógica reutilizável, Tailwind para estilo.

## Convenções gerais

- Erros e logs do backend em inglês; mensagens de UI em português.
- Logs sempre com `log/slog` estruturado (atributos chave-valor, ex.: `slog.Error("msg", "userId", id, "error", err)`). Nunca `log.Printf` nem interpolar dados na mensagem.
- Código simples e linear: early returns, sem abstrações especulativas, sem interfaces sem segundo implementador.
- Comentários explicam *por quê* (trade-offs, segurança), não *o quê*.
- Lint: `golangci-lint run ./...` no backend (config em `backend/.golangci.yml`; regra `exported` do revive desabilitada de propósito — não adicione comentários cerimoniais); `npm run lint` e `npm run format` no frontend.
- Antes de finalizar: `go build ./...`, `go vet ./...`, `golangci-lint run ./...`, `go test ./tests/` no backend; `npm run lint` e `npm run build` no frontend.

## Comandos úteis

```bash
# Stack completo
cd infra && docker compose up -d --build      # exige .env (veja env.example)

# Dev backend
cd backend && go run ./cmd/api                # exige DATABASE_URL, JWT_SECRET, ADMIN_PASSWORD

# Dev frontend (proxy /api → localhost:8080)
cd frontend && npm run dev

# Testes
cd backend && go test ./tests/ -v
```
