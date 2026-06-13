# Roadmap de Evolução da Autenticação

## Objetivo

Evoluir o sistema de autenticação e autorização para um padrão próximo ao utilizado em aplicações corporativas modernas, mantendo simplicidade arquitetural e foco em aprendizado.

---

# Fase 1 - Rotação de Refresh Tokens

## Problema

Atualmente o mesmo Refresh Token pode ser reutilizado diversas vezes durante toda sua validade.

Caso ele seja comprometido, um invasor poderá continuar renovando sessões enquanto o token permanecer válido.

---

## Objetivo

Implementar Refresh Token Rotation.

A cada renovação:

```text
Refresh Token A
↓
Refresh Token B
↓
Refresh Token C
↓
Refresh Token D
```

O token anterior é imediatamente invalidado.

---

## Alterações no Banco

### Sessions

Adicionar controle de revogação e rastreamento.

```sql
ALTER TABLE sessions
ADD COLUMN revoked_at TIMESTAMP NULL;
```

---

## Fluxo

### Login

```text
Usuário faz login
↓
Cria sessão
↓
Gera Refresh Token
↓
Armazena hash do token
```

### Refresh

```text
Recebe Refresh Token
↓
Valida hash
↓
Revoga sessão anterior
↓
Cria nova sessão
↓
Gera novo Refresh Token
```

---

## Benefícios

* Reduz impacto de vazamento de Refresh Tokens.
* Aproxima o sistema dos padrões modernos de autenticação.
* Permite detecção de reutilização indevida.

---

# Fase 2 - Fingerprint de Sessão

## Objetivo

Registrar informações do dispositivo utilizado para autenticação.

---

## Alterações no Banco

### Sessions

```sql
ALTER TABLE sessions
ADD COLUMN ip_address VARCHAR(255);

ALTER TABLE sessions
ADD COLUMN user_agent TEXT;

ALTER TABLE sessions
ADD COLUMN last_activity_at TIMESTAMP;
```

---

## Informações Registradas

### IP

```text
189.xxx.xxx.xxx
```

### User Agent

```text
Chrome 138
Windows 11
```

### Última Atividade

```text
2026-06-10 15:42:10
```

---

## Atualização de Atividade

Sempre que uma requisição autenticada for realizada:

```text
Atualizar last_activity_at
```

---

## Benefícios

* Auditoria mais rica.
* Rastreamento de sessões.
* Base para futuras funcionalidades de segurança.

---

# Fase 3 - Auditoria Completa

## Objetivo

Expandir o sistema atual de auditoria para registrar todos os eventos relevantes.

---

## Eventos de Autenticação

```text
login.success
login.failed
logout
```

---

## Eventos de Sessão

```text
session.created
session.refreshed
session.revoked
session.expired
```

---

## Eventos de Usuário

```text
user.created
user.updated
user.deleted
user.activated
user.deactivated
```

---

## Eventos de Permissões

```text
permission.granted
permission.revoked
permission.created
permission.updated
permission.deleted
```

---

## Dados Registrados

```json
{
  "user_id": 15,
  "action": "permission.granted",
  "entity": "user_permissions",
  "entity_id": 8
}
```

---

## Benefícios

* Rastreabilidade completa.
* Facilidade de investigação.
* Histórico de alterações.

---

# Fase 4 - Rate Limiting Inteligente

## Objetivo

Proteger o sistema contra ataques de força bruta.

---

## Tecnologia

Redis.

---

## Estratégia

### Por IP

```text
login:ip:{ip}
```

### Por Email

```text
login:email:{email}
```

---

## Escalonamento

### Primeira Faixa

```text
5 tentativas
↓
1 minuto
```

### Segunda Faixa

```text
10 tentativas
↓
15 minutos
```

### Terceira Faixa

```text
20 tentativas
↓
24 horas
```

---

## Benefícios

* Proteção contra força bruta.
* Redução de abuso.
* Aprendizado prático de Redis.

---

# Fase 5 - Permissões Hierárquicas

## Objetivo

Simplificar a administração de permissões.

---

## Situação Atual

```text
users.read
users.write
users.delete
```

---

## Nova Estrutura

```text
users.*
contracts.*
financial.*

*
```

---

## Regras

### Exemplo

Usuário possui:

```text
users.*
```

Valida automaticamente:

```text
users.read
users.write
users.delete
```

---

### Super Administrador

```text
*
```

Possui acesso total.

---

## Implementação

Adaptar o método:

```go
HasPermission()
```

para suportar:

```text
Permissão exata
Wildcard de módulo
Wildcard global
```

---

## Benefícios

* Menos permissões atribuídas manualmente.
* Administração simplificada.
* Melhor experiência operacional.

---

# Fase 6 - JWT Assimétrico (RS256)

## Objetivo

Migrar do modelo simétrico (HS256) para assinatura assimétrica (RS256).

---

## Estrutura

```text
keys/
├── private.pem
└── public.pem
```

---

## Responsabilidades

### Private Key

```text
Assina JWTs
```

### Public Key

```text
Valida JWTs
```

---

## Fluxo

```text
Private Key
↓
Assina Token
↓
JWT
↓
Public Key
↓
Valida Token
```

---

## Segurança

### Nunca versionar

```text
private.pem
```

Adicionar ao:

```text
.gitignore
```

---

## Benefícios

* Arquitetura compatível com sistemas distribuídos.
* Melhor separação de responsabilidades.
* Aprendizado de padrões utilizados por provedores de identidade modernos.

---

# Estrutura Final Esperada

## Segurança

* JWT RS256
* Refresh Token Rotation
* Rate Limiting Inteligente
* Auditoria Completa

---

## Sessões

* Sessões Persistidas
* Fingerprint de Sessão
* Rastreamento de Atividade

---

## Permissões

* Permissões Granulares
* Permissões Hierárquicas
* Wildcards

---

## Infraestrutura

* PostgreSQL
* Redis
* Docker
* Docker Compose
* Nginx

---

# Ordem Recomendada

```text
1. Refresh Token Rotation
2. Fingerprint de Sessão
3. Auditoria Completa
4. Rate Limiting Inteligente
5. Permissões Hierárquicas
6. JWT RS256
```

Essa ordem maximiza o ganho de segurança e aprendizado enquanto mantém a implementação incremental e fácil de testar.
