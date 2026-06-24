# Rotina de Backup e Restauração do Banco de Dados

Esta página documenta o funcionamento, automação e restauração do banco de dados PostgreSQL 16 do projeto utilizando backups lógicos.

## Onde ficam os backups?
Os backups são gerados no formato de arquivo binário compactado do Postgres (`pg_dump -F c`) e salvos no diretório local do host em:
```
infra/backups/auth_YYYYMMDD_HHMMSS.dump
```
> [!NOTE]
> Esta pasta está configurada no `.gitignore` para nunca ser versionada.

---

## Como rodar manualmente?

### Windows (PowerShell)
Abra o PowerShell como Administrador, navegue até a pasta `infra/` e execute:
```powershell
./backup.ps1
```

### Linux / macOS / WSL (Bash)
Dê permissão de execução e execute o script na pasta `infra/`:
```bash
chmod +x backup.sh
./backup.sh
```

---

## Como agendar a automação?

### Windows (Agendador de Tarefas)
1. Abra o **Agendador de Tarefas** do Windows (`Task Scheduler`).
2. Clique em **Criar Tarefa Básica...** no menu lateral.
3. Nomeie a tarefa (ex.: `Secure Auth Database Backup`).
4. Escolha a frequência como **Diariamente** e defina o horário (ex.: `03:00 AM`).
5. Na ação, selecione **Iniciar um programa**.
6. No campo **Programa/script**, digite `powershell.exe`.
7. No campo **Adicionar argumentos (opcional)**, digite:
   `-ExecutionPolicy Bypass -File "C:\caminho\completo\para\projeto\secure-auth\infra\backup.ps1"`
8. Conclua a criação da tarefa.

### Linux / WSL (Crontab)
Para rodar todos os dias às 03:00 AM:
1. Abra o editor do crontab:
   ```bash
   crontab -e
   ```
2. Adicione a seguinte linha (ajustando o caminho absoluto do projeto):
   ```cron
   0 3 * * * /bin/bash /caminho/completo/para/projeto/secure-auth/infra/backup.sh >/dev/null 2>&1
   ```

---

## Procedimento de Restauração (pg_restore)

Para restaurar um backup existente sobre o banco de dados principal ou para fins de testes de integridade:

### 1. Testando a integridade de um arquivo de backup
Sempre teste a restauração em um banco temporário antes de aplicá-la em produção:

1. **Crie um banco de testes temporário**:
   ```bash
   docker exec -it infra-postgres-1 psql -U auth -c "CREATE DATABASE auth_backup_test;"
   ```
2. **Execute a restauração**:
   Substitua `[ARQUIVO_DE_BACKUP]` pelo nome do arquivo gerado (ex.: `auth_20260624_130000.dump`):
   ```bash
   docker exec -i infra-postgres-1 pg_restore -U auth -d auth_backup_test < infra/backups/[ARQUIVO_DE_BACKUP]
   ```
3. **Verifique se as tabelas foram criadas**:
   Conecte ao banco de testes e confira a existência das tabelas:
   ```bash
   docker exec -it infra-postgres-1 psql -U auth -d auth_backup_test -c "\dt"
   ```
4. **Remova o banco de testes após validar**:
   ```bash
   docker exec -it infra-postgres-1 psql -U auth -c "DROP DATABASE auth_backup_test;"
   ```

### 2. Sobrescrevendo o banco de dados ativo em caso de desastre
> [!CAUTION]
> Este procedimento apagará todos os dados atuais do banco de produção e os substituirá pelo backup.

1. **Desconecte conexões ativas e recrie o banco original**:
   ```bash
   docker exec -it infra-postgres-1 psql -U auth -c "REVOKE CONNECT ON DATABASE auth FROM public;"
   docker exec -it infra-postgres-1 psql -U auth -c "SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = 'auth' AND pid <> pg_backend_pid();"
   docker exec -it infra-postgres-1 psql -U auth -c "DROP DATABASE auth;"
   docker exec -it infra-postgres-1 psql -U auth -c "CREATE DATABASE auth;"
   ```
2. **Importe o backup**:
   ```bash
   docker exec -i infra-postgres-1 pg_restore -U auth -d auth < infra/backups/[ARQUIVO_DE_BACKUP]
   ```
3. **Reinicie o container da API** para restabelecer as conexões do pool:
   ```bash
   docker compose restart api
   ```
