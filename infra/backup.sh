#!/bin/bash
# Script de Backup do PostgreSQL para Linux/macOS/WSL
# Requisitos: Docker rodando e container 'infra-postgres-1' ativo.

set -e

# Obter diretório onde o script está localizado
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="$DIR/backups"

# 1. Verificar se o container do banco está rodando
RUNNING=$(docker inspect -f '{{.State.Running}}' infra-postgres-1 2>/dev/null || echo "false")
if [ "$RUNNING" != "true" ]; then
    echo "ERRO: O container 'infra-postgres-1' nao esta rodando." >&2
    exit 1
fi

# 2. Garantir que a pasta de backups existe
mkdir -p "$BACKUP_DIR"

# 3. Definir nome do arquivo com timestamp completo
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
FILE_NAME="auth_$TIMESTAMP.dump"
BACKUP_PATH="$BACKUP_DIR/$FILE_NAME"

echo "Iniciando backup do banco de dados 'auth'..."

# 4. Executar o pg_dump (formato Custom -F c que ja possui compressao nativa)
# Desativar saída imediata por erro para capturar o status de saída do pg_dump
set +e
docker exec -i infra-postgres-1 pg_dump -U auth -F c auth > "$BACKUP_PATH"
STATUS=$?
set -e

# 5. Verificar o sucesso do backup
if [ $STATUS -ne 0 ]; then
    echo "ERRO: Falha ao gerar o backup com pg_dump." >&2
    rm -f "$BACKUP_PATH"
    exit 1
fi

echo "Backup gerado com sucesso em: $BACKUP_PATH"

# 6. Rotacionar backups antigos (manter apenas os ultimos 7 dias)
echo "Verificando expiracao de backups antigos..."
# Apenas remove se a pasta contiver backups antigos
find "$BACKUP_DIR" -name "auth_*.dump" -type f -mtime +7 -delete

echo "Rotina de backup concluida."
