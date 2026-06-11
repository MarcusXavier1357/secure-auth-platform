1. Detecção de Reutilização de Refresh Token

Essa é provavelmente a melhoria mais importante.

Imagine:

Refresh A

é roubado.

Usuário legítimo:

Refresh A
↓
Refresh B

Agora o invasor tenta usar:

Refresh A

novamente.

O sistema detecta:

Refresh Token revogado sendo reutilizado

Resultado:

Revogar TODAS as sessões do usuário

Isso é uma técnica muito usada atualmente.


3. Security Headers

Via Nginx.

Exemplo:

X-Frame-Options: DENY
X-Content-Type-Options: nosniff
Referrer-Policy: strict-origin

e principalmente:

Content-Security-Policy

4. CSP (Content Security Policy)

Essa sozinha reduz muito o risco de XSS.

Exemplo:

Content-Security-Policy:
default-src 'self'

Mesmo que alguém injete:

<script>
malicioso()
</script>

o navegador bloqueia.


6. Detecção de Comportamento Suspeito

Exemplo:

Login Fortaleza
↓
2 minutos
↓
Login Moscou

ou

20 falhas seguidas

Gerar evento:

security.alert


8. Hardening de Cookies

Refresh Token:

HttpOnly
Secure
SameSite=Strict

Você já deve estar planejando isso.

9. Hash Mais Moderno

Hoje:

bcrypt

Muito bom.

Mas atualmente muitos especialistas preferem:

Argon2

porque é mais resistente a ataques em GPU.

Eu ainda usaria bcrypt num projeto pessoal porque é simples e amplamente aceito.

10. Proteção contra Enumeração de Usuários

Evitar respostas como:

Usuário não encontrado

ou

Senha incorreta

Sempre:

Credenciais inválidas