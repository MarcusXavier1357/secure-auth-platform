# Script de Backup do PostgreSQL para Windows
# Requisitos: Docker rodando e container 'infra-postgres-1' ativo.

$ErrorActionPreference = "Stop"

# 1. Verificar se o container do banco está rodando
$containerStatus = docker inspect -f '{{.State.Running}}' infra-postgres-1 2>$null
if ($LASTEXITCODE -ne 0 -or $containerStatus -ne "true") {
    Write-Host "ERRO: O container 'infra-postgres-1' nao esta rodando." -ForegroundColor Red
    exit 1
}

# 2. Garantir que a pasta de backups existe
$backupDir = Join-Path $PSScriptRoot "backups"
if (-not (Test-Path $backupDir)) {
    New-Item -ItemType Directory -Path $backupDir | Out-Null
    Write-Host "Diretorio de backups criado em: $backupDir" -ForegroundColor Green
}

# 3. Definir nome do arquivo com timestamp completo
$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$fileName = "auth_$timestamp.dump"
$backupPath = Join-Path $backupDir $fileName

Write-Host "Iniciando backup do banco de dados 'auth'..." -ForegroundColor Cyan

# 4. Executar o pg_dump (formato Custom -F c que ja possui compressao nativa)
# Importante: executamos via cmd.exe /c para evitar corrupcao de codificacao binaria que o redirection do PowerShell causa.
cmd.exe /c "docker exec -i infra-postgres-1 pg_dump -U auth -F c auth > `"$backupPath`""

# 5. Verificar o sucesso do backup
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERRO: Falha ao gerar o backup com pg_dump." -ForegroundColor Red
    if (Test-Path $backupPath) {
        Remove-Item $backupPath -Force
    }
    exit 1
}

Write-Host "Backup gerado com sucesso em: $backupPath" -ForegroundColor Green

# 6. Rotacionar backups antigos (manter apenas os ultimos 7 dias)
$limitDate = (Get-Date).AddDays(-7)
$oldBackups = Get-ChildItem -Path $backupDir -Filter "auth_*.dump" | Where-Object { $_.CreationTime -lt $limitDate }

if ($oldBackups) {
    Write-Host "Removendo backups com mais de 7 dias..." -ForegroundColor Yellow
    foreach ($file in $oldBackups) {
        Remove-Item $file.FullName -Force
        Write-Host "Removido: $($file.Name)" -ForegroundColor Gray
    }
}

Write-Host "Rotina de backup concluida." -ForegroundColor Green
