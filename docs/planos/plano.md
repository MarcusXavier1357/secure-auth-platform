# Plano de Implementação - Sistema de Autenticação e Autorização

## Objetivo

Implementar um sistema de autenticação e autorização seguro, simples de manter e preparado para crescimento futuro.

A arquitetura deverá utilizar:

* Frontend React + Vite
* Backend Go + Fiber
* PostgreSQL
* Redis
* Bun (ORM)
* golang-migrate (migrations)
* JWT
* Refresh Tokens
* bcrypt
* Audit Logs
* Docker
* Nginx

---

# Arquitetura Geral

```text
Frontend (React)
        ↓
     HTTPS
        ↓
      Nginx
        ↓
    API Go/Fiber
        ↓
   ┌────┴────┐
   ↓         ↓
 Redis    Bun (ORM)
(cache)       ↓
         PostgreSQL
   (fonte da verdade)
```

Responsabilidades:

### Frontend

* Login
* Logout
* Consumo da API
* Controle visual de permissões
* Renovação automática de sessão

### Backend

* Autenticação
* Autorização
* Regras de negócio
* Auditoria
* Emissão de tokens
* Cache e rate limiting (Redis)

### PostgreSQL

* Usuários
* Permissões
* Sessões
* Logs de auditoria

### Redis

* Cache de permissões por usuário
* Rate limiting (login / brute force)
* Estado efêmero — **não** substitui PostgreSQL como fonte da verdade

---

# Modelo de Permissões

## Conceito

Roles e permissões possuem responsabilidades diferentes.

### Roles

Representam a classificação organizacional do usuário.

Exemplos:

```text
Admin
Supervisor
Operador
Financeiro
Vendedor
```

As roles não concedem permissões automaticamente.

Servem apenas para:

* Organização
* Relatórios
* Classificação interna
* Regras de negócio específicas

---

### Permissões

Definem o que o usuário pode fazer.

São atribuídas diretamente ao usuário.

Exemplos:

```text
clients.read
clients.write
clients.delete

contracts.read
contracts.write
contracts.cancel

financial.read
financial.export

users.manage
permissions.manage
audit_logs.read
```

---

# Modelo de Dados

## Roles

```sql
CREATE TABLE roles (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## Users

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    role_id BIGINT REFERENCES roles(id),

    name VARCHAR(255) NOT NULL,

    email VARCHAR(255) NOT NULL UNIQUE,

    password_hash TEXT NOT NULL,

    active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## Permissions

```sql
CREATE TABLE permissions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    code VARCHAR(255) NOT NULL UNIQUE,

    description TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## User Permissions

```sql
CREATE TABLE user_permissions (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,

    PRIMARY KEY (
        user_id,
        permission_id
    )
);
```

---

## Sessions

```sql
CREATE TABLE sessions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    user_id BIGINT NOT NULL REFERENCES users(id),

    refresh_token_hash TEXT NOT NULL,

    expires_at TIMESTAMP NOT NULL,

    revoked BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## Audit Logs

```sql
CREATE TABLE audit_logs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    user_id BIGINT REFERENCES users(id),

    action VARCHAR(100) NOT NULL,

    entity VARCHAR(100) NOT NULL,

    entity_id BIGINT,

    old_data JSONB,

    new_data JSONB,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### Índices recomendados

```sql
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role_id ON users(role_id);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at) WHERE revoked = FALSE;
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity, entity_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
```

---

# Acesso a Dados (ORM)

## Escolha: Bun

Utilizar [Bun](https://bun.uptrace.dev/) como camada de acesso ao PostgreSQL.

Motivos para este projeto:

* **SQL-first** — o schema SQL deste plano continua sendo a fonte da verdade; o ORM mapeia tabelas, não substitui o desenho do banco.
* **PostgreSQL nativo** — suporte a `JSONB` (audit logs), arrays e tipos do Postgres.
* **Type-safe** — structs Go com tags `bun` alinhadas às tabelas existentes.
* **Menos magia que GORM** — queries explícitas, mais previsíveis em auth/sessões.
* **Migrations integradas** — Bun Migrate ou `golang-migrate` para versionar o schema.

### Alternativas consideradas

| Opção | Quando usar |
|-------|-------------|
| **GORM** | Equipe já familiar; prioriza velocidade de bootstrap e hooks (ex.: `BeforeCreate` para audit). |
| **sqlc** | Preferência por SQL puro + código gerado; não é ORM, mas é muito idiomático em Go. |
| **Ent** | Domínio complexo com muitos relacionamentos; provavelmente excesso para auth inicial. |

**Recomendação:** Bun + `golang-migrate`. Se a equipe já domina GORM, trocar é aceitável — o modelo de dados não muda.

---

## Migrations

Não usar `AutoMigrate` em produção.

Fluxo:

1. Schema SQL versionado em `backend/migrations/` (ex.: `000001_init.up.sql`).
2. Aplicar via `golang-migrate` no startup do Docker ou comando dedicado.
3. Structs Bun em `internal/models/` espelham as tabelas após migration.

```text
backend/migrations/
├── 000001_init.up.sql
├── 000001_init.down.sql
├── 000002_add_indexes.up.sql
└── 000002_add_indexes.down.sql
```

---

## Modelos Bun (exemplo)

```go
type User struct {
    bun.BaseModel `bun:"table:users,alias:u"`

    ID           int64     `bun:"id,pk,autoincrement"`
    RoleID       *int64    `bun:"role_id"`
    Name         string    `bun:"name,notnull"`
    Email        string    `bun:"email,notnull,unique"`
    PasswordHash string    `bun:"password_hash,notnull"`
    Active       bool      `bun:"active,notnull,default:true"`
    CreatedAt    time.Time `bun:"created_at,notnull,default:current_timestamp"`
    UpdatedAt    time.Time `bun:"updated_at,notnull,default:current_timestamp"`

    Role        *Role         `bun:"rel:belongs-to,join:role_id=id"`
    Permissions []Permission  `bun:"m2m:user_permissions,join:User=Permission"`
}

type AuditLog struct {
    bun.BaseModel `bun:"table:audit_logs,alias:al"`

    ID        int64                  `bun:"id,pk,autoincrement"`
    UserID    *int64                 `bun:"user_id"`
    Action    string                 `bun:"action,notnull"`
    Entity    string                 `bun:"entity,notnull"`
    EntityID  *int64                 `bun:"entity_id"`
    OldData   map[string]interface{} `bun:"old_data,type:jsonb"`
    NewData   map[string]interface{} `bun:"new_data,type:jsonb"`
    CreatedAt time.Time              `bun:"created_at,notnull,default:current_timestamp"`
}
```

---

## Padrão de repositório

Manter queries fora dos handlers HTTP:

```text
internal/
├── models/       # structs Bun
├── repository/   # UserRepository, SessionRepository, AuditRepository
└── service/      # regras de negócio (auth, permissions)
```

Exemplo de query parametrizada via Bun:

```go
err := db.NewSelect().
    Model(&user).
    Relation("Permissions").
    Where("u.email = ?", email).
    Where("u.active = ?", true).
    Scan(ctx)
```

Bun gera SQL parametrizado — mantém proteção contra SQL injection.

---

## Permissões e cache (Redis)

No middleware `RequirePermission`:

1. Consultar cache Redis: `permissions:user:{userId}`.
2. Cache miss → carregar do PostgreSQL via Bun e gravar no Redis.
3. Verificar se o usuário possui a permissão requerida.
4. **Não** colocar permissões no JWT.

TTL padrão do cache:

```text
5 minutos
```

Invalidar cache quando:

* Permissão concedida ou revogada (`DEL permissions:user:{userId}`)
* Usuário desativado ou excluído
* Logout global / revogação de todas as sessões do usuário

Se Redis estiver indisponível: consultar PostgreSQL diretamente (degraded mode) e registrar alerta — auth continua funcionando, só mais lento.

---

# Redis (Cache e Rate Limiting)

## Escolha: go-redis

Utilizar [go-redis](https://github.com/redis/go-redis) (`github.com/redis/go-redis/v9`) como client Redis no backend.

Redis entra **desde o início** — não como otimização futura. Docker Compose sobe Postgres + Redis juntos; a API só fica healthy com ambos disponíveis.

---

## O que fica no Redis vs PostgreSQL

| Dado | Onde | Motivo |
|------|------|--------|
| Usuários, roles, permissões | PostgreSQL | Persistência, auditoria |
| Sessões / refresh token hash | PostgreSQL | Revogação durável, auditável |
| Audit logs | PostgreSQL | Histórico permanente |
| Cache de permissões | Redis | Reduzir joins a cada request |
| Contadores de login (IP/email) | Redis | Rate limit distribuído entre instâncias |

**Regra:** Redis pode ser perdido/reiniciado sem perda de dados críticos. Sessões e tokens **não** migram para Redis.

---

## Convenção de chaves

```text
permissions:user:{userId}          → SET de codes (ex.: clients.read)
ratelimit:login:ip:{ip}            → contador INT, TTL 15 min
ratelimit:login:email:{email}      → contador INT, TTL 15 min
```

Usar prefixo configurável via env (`REDIS_KEY_PREFIX=auth:`) para ambientes compartilhados.

---

## Cache de permissões

### Formato

```text
Key:   permissions:user:123
Type:  SET
Value: clients.read, clients.write, users.manage
TTL:   5m
```

### Fluxo (cache-aside)

```text
RequirePermission("clients.read")
        ↓
Redis SMISMEMBER / SISMEMBER
        ↓
Hit  → verifica permissão
Miss → SELECT no Postgres → SADD + EXPIRE → verifica permissão
```

### Exemplo Go

```go
key := fmt.Sprintf("permissions:user:%d", userID)

codes, err := rdb.SMembers(ctx, key).Result()
if err == redis.Nil || len(codes) == 0 {
    codes, err = permRepo.ListCodesByUserID(ctx, userID)
    if err != nil { return err }

    pipe := rdb.Pipeline()
    pipe.Del(ctx, key)
    if len(codes) > 0 {
        pipe.SAdd(ctx, key, codesToAny(codes)...)
    }
    pipe.Expire(ctx, key, 5*time.Minute)
    _, err = pipe.Exec(ctx)
}
```

---

## Rate limiting (login)

Aplicar **antes** da comparação bcrypt — evita custo de hash em ataques de força bruta.

Limites (já definidos na seção Segurança):

```text
5 tentativas / 15 minutos — por IP
5 tentativas / 15 minutos — por email
```

Bloquear se **qualquer** dos dois limites for atingido.

### Algoritmo

Fixed window com `INCR` + `EXPIRE` na primeira tentativa:

```go
func (s *RateLimiter) Increment(ctx context.Context, key string, window time.Duration) (int64, error) {
    count, err := s.rdb.Incr(ctx, key).Result()
    if err != nil {
        return 0, err
    }
    if count == 1 {
        s.rdb.Expire(ctx, key, window)
    }
    return count, nil
}
```

Resposta ao exceder limite:

```http
HTTP/1.1 429 Too Many Requests
Retry-After: 900
```

Registrar `login.failed` no audit log com motivo `rate_limited`.

### Fluxo no login

```text
POST /auth/login
        ↓
Increment ratelimit:login:ip:{ip}
Increment ratelimit:login:email:{email}
        ↓
count > 5 em algum? → 429
        ↓
Buscar usuário + bcrypt
        ↓
Sucesso → DEL ratelimit:login:email:{email} (opcional, limpa após login ok)
Falha  → continua contador
```

---

## Configuração

Variáveis de ambiente:

```text
REDIS_URL=redis://redis:6379/0
REDIS_KEY_PREFIX=auth:
PERMISSIONS_CACHE_TTL=5m
LOGIN_RATE_LIMIT=5
LOGIN_RATE_WINDOW=15m
```

---

## Docker Compose (Redis)

```yaml
services:
  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

  api:
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
```

`appendonly yes` persiste cache/rate limits entre restarts — aceitável para dev; em produção o cache reconstrói sozinho após restart.

---

## Comportamento em falha

| Cenário | Comportamento |
|---------|---------------|
| Redis down no middleware de permissões | Fallback para PostgreSQL |
| Redis down no rate limit do login | Rejeitar login com `503` ou fallback conservador — **preferir 503** para não abrir brecha de brute force |
| Redis reiniciado | Cache miss automático; rate limits resetam (aceitável) |

---

# Fluxo de Autenticação

## Login

### Endpoint

```http
POST /auth/login
```

### Request

```json
{
  "email": "usuario@email.com",
  "password": "senha"
}
```

### Processo

1. Verificar rate limit no Redis (IP + email).
2. Buscar usuário pelo email.
3. Verificar se está ativo.
4. Comparar senha utilizando bcrypt.
5. Criar Access Token.
6. Criar Refresh Token.
7. Registrar sessão no PostgreSQL.
8. Retornar Access Token.

---

## Access Token

### Características

* JWT
* Curta duração
* Utilizado em todas as chamadas da API

### Validade

```text
15 minutos
```

### Payload

```json
{
  "sub": 123,
  "email": "usuario@email.com",
  "iat": 123456,
  "exp": 123456
}
```

### Observação

Não armazenar permissões dentro do JWT.

---

## Refresh Token

### Características

* Longa duração
* Utilizado apenas para renovação de sessão

### Validade

```text
30 dias
```

### Armazenamento

Cookie:

```text
HttpOnly
Secure
SameSite=Strict
```

Banco:

```text
Apenas hash do token.
```

Nunca armazenar o token puro.

---

## Renovação de Sessão

### Endpoint

```http
POST /auth/refresh
```

### Processo

1. Ler Refresh Token do cookie.
2. Validar sessão.
3. Validar expiração.
4. Validar hash.
5. Gerar novo Access Token.
6. (Opcional, recomendado) Rotacionar Refresh Token — invalidar o anterior e emitir novo hash na mesma sessão.

---

## Logout

### Endpoint

```http
POST /auth/logout
```

### Processo

1. Revogar sessão.
2. Invalidar Refresh Token.
3. Remover cookie.

---

# Middleware de Autenticação

## Responsabilidades

* Validar JWT.
* Validar assinatura.
* Validar expiração.
* Verificar usuário ativo.

Após validação:

```go
ctx.Locals("userId", userId)
```

---

# Middleware de Permissões

## Objetivo

Garantir que apenas usuários autorizados executem determinadas ações.

### Exemplo

```go
app.Get(
    "/clients",
    AuthMiddleware,
    RequirePermission("clients.read"),
    ListClients,
)
```

### Fluxo

```text
Usuário autenticado
        ↓
Redis: permissions:user:{id}
        ↓
Cache miss? → PostgreSQL → popula Redis
        ↓
Possui permissão?
        ↓
SIM → Continua
NÃO → 403 Forbidden
```

---

# Auditoria

Todas as operações críticas devem gerar registros.

## Exemplos

```text
user.created
user.updated
user.deleted

permission.granted
permission.revoked

contract.created
contract.updated
contract.cancelled

login.success
login.failed
logout
```

## Informações Registradas

* Usuário responsável
* Tipo da ação
* Entidade afetada
* ID da entidade
* Dados anteriores
* Dados novos
* Data e hora

---

# Segurança

## Senhas

Utilizar:

```text
bcrypt
```

Nunca armazenar:

```text
senha em texto puro
MD5
SHA1
```

---

## SQL Injection

Utilizar apenas queries parametrizadas — via Bun ou SQL raw com placeholders (`$1`, `?`).

Evitar concatenação de strings em SQL.

Exemplo com Bun:

```go
db.NewSelect().
    Model(&user).
    Where("email = ?", email).
    Scan(ctx)
```

Queries raw (se necessário) devem usar placeholders:

```go
db.QueryContext(ctx, "SELECT * FROM users WHERE email = $1", email)
```

---

## Força Bruta

Implementar via **Redis** (desde o início):

```text
5 tentativas / 15 minutos — por IP
5 tentativas / 15 minutos — por email
```

Chaves:

```text
ratelimit:login:ip:{ip}
ratelimit:login:email:{email}
```

Resposta: `429 Too Many Requests` com header `Retry-After`.

Detalhes de implementação na seção **Redis (Cache e Rate Limiting)**.

---

## CSRF

Configurar cookies:

```text
SameSite=Strict
```

---

## HTTPS

Obrigatório em produção.

Todo tráfego deverá utilizar TLS.

---

## Tokens

Não armazenar Access Token em:

```text
localStorage
```

Preferencialmente manter em memória na aplicação React.

---

## Backend como Fonte da Verdade

Nunca confiar em validações do frontend.

Toda autorização deve ser validada novamente no backend.

---

# Estrutura Inicial do Projeto

```text
frontend/
├── src/
├── components/
├── pages/
├── services/
├── hooks/
└── providers/

backend/
├── cmd/
│   └── api/
├── internal/
│   ├── auth/
│   ├── users/
│   ├── permissions/
│   ├── sessions/
│   ├── audit/
│   ├── models/       # structs Bun
│   ├── repository/   # acesso a dados
│   └── service/      # regras de negócio
├── cache/            # Redis client, permission cache, rate limiter
├── middleware/
├── database/         # conexão Bun + pool
├── migrations/       # SQL versionado (golang-migrate)
└── routes/

infra/
├── nginx/
├── postgres/
├── redis/
└── docker-compose.yml
```

---

# Stack Final

## Frontend

* React
* Vite
* Tailwind CSS
* shadcn/ui
* TanStack Query
* React Router

## Backend

* Go
* Fiber
* PostgreSQL
* Redis
* go-redis
* Bun (ORM)
* golang-migrate
* JWT
* bcrypt

## Infraestrutura

* Docker
* Docker Compose
* Nginx
* Redis 7

## Funcionalidades de Segurança

* JWT de curta duração
* Refresh Token
* Sessões revogáveis
* Controle granular de permissões
* Cache de permissões (Redis)
* Rate limiting distribuído (Redis)
* Audit Logs
* Proteção contra SQL Injection
* Proteção contra força bruta
* HTTPS obrigatório
