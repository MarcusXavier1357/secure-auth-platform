# Plano de Implementação — Política de Senhas Seguras e Verificação de Vazamentos

## Objetivo

Implementar uma política moderna de senhas que:

* Aumente a segurança real das contas.
* Reduza o uso de senhas previsíveis.
* Bloqueie senhas já expostas em vazamentos públicos.
* Mantenha uma boa experiência para o usuário.
* Esteja alinhada às recomendações atuais da indústria.

---

# Requisitos Funcionais

## Criação de Senha

O usuário deve informar uma senha que atenda aos critérios mínimos:

### Comprimento

* Mínimo: 12 caracteres
* Máximo: 128 caracteres

### Complexidade

Obrigatório possuir:

* Pelo menos 1 letra minúscula
* Pelo menos 1 letra maiúscula
* Pelo menos 1 número

Opcional:

* Símbolos especiais

Exemplos válidos:

```text
MinhaSenha2026
Fortaleza123ABC
K9m2Q7x4L8z1
```

Exemplos inválidos:

```text
abcdefg
123456789
minhasenha
MINHASENHA
```

---

# Regras de Bloqueio

## Senhas Comuns

Bloquear senhas presentes em listas conhecidas de senhas fracas.

Exemplos:

```text
123456
123456789
password
senha123
qwerty
admin
```

---

## Sequências

Bloquear sequências previsíveis.

Exemplos:

```text
12345678
87654321
abcdefgh
qwertyui
```

---

## Repetições Excessivas

Bloquear padrões repetitivos.

Exemplos:

```text
aaaaaaaaaaaa
111111111111
abababababab
xyzxyzxyzxyz
```

---

## Dados do Próprio Usuário

Bloquear senhas contendo:

* Nome
* Sobrenome
* E-mail
* Login

Exemplo:

Usuário:

```text
Marcus Xavier
```

Senhas bloqueadas:

```text
Marcus123
Marcus2026
Xavier123
MarcusXavier@
```

---

# Verificação Contra Vazamentos

## Objetivo

Impedir o uso de senhas já comprometidas em incidentes de segurança públicos.

---

## Fonte de Dados

Utilizar a API de Passwords do Have I Been Pwned (HIBP).

Método recomendado:

* k-Anonymity

Benefícios:

* A senha nunca é enviada ao serviço.
* Apenas os 5 primeiros caracteres do hash SHA-1 são transmitidos.
* Preserva privacidade do usuário.

---

## Fluxo

### Passo 1

Receber senha.

### Passo 2

Gerar SHA-1.

### Passo 3

Separar:

```text
Prefixo = primeiros 5 caracteres
Sufixo = restante do hash
```

### Passo 4

Consultar:

```http
GET https://api.pwnedpasswords.com/range/{prefix}
```

### Passo 5

Verificar localmente se o sufixo existe na resposta.

### Passo 6

Caso exista:

```text
Esta senha já apareceu em vazamentos públicos.
Escolha uma senha diferente.
```

---

# Fluxo de Cadastro

```text
Usuário digita senha
        ↓
Validação de tamanho
        ↓
Validação de composição
        ↓
Validação de padrões proibidos
        ↓
Consulta HIBP
        ↓
Senha aprovada
        ↓
Hash Argon2id
        ↓
Persistência
```

---

# Mensagens de Erro

## Comprimento

```text
A senha deve possuir pelo menos 12 caracteres.
```

## Complexidade

```text
A senha deve conter ao menos uma letra maiúscula, uma minúscula e um número.
```

## Vazamento

```text
Esta senha já apareceu em vazamentos públicos e não pode ser utilizada.
```

## Força

```text
A senha é considerada fraca. Escolha uma senha mais forte.
```

---

# Critérios de Aceite

## Segurança

* Senhas vazadas são rejeitadas.
* Senhas comuns são rejeitadas.
* Senhas previsíveis são rejeitadas.
* Senhas são armazenadas com Argon2id.

## Experiência do Usuário

* Feedback em tempo real.
* Mensagens claras.
* Validações executadas antes do envio do formulário.

## Performance

* Validações locais executam em menos de 100 ms.
* Consulta HIBP executa de forma assíncrona.
* Hash Argon2id executa apenas após todas as validações serem aprovadas.
