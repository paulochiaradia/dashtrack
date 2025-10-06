# Script PowerShell para executar migrações de banco de dados
# Usage: .\run-migrations.ps1

Write-Host "🚀 Executando migrações do banco de dados..." -ForegroundColor Green

# Verificar se o Docker está rodando
try {
    docker info | Out-Null
} catch {
    Write-Host "❌ Docker não está rodando. Por favor, inicie o Docker Desktop." -ForegroundColor Red
    exit 1
}

# Verificar se o container do banco está rodando
$dbStatus = docker-compose ps db
if ($dbStatus -notmatch "Up") {
    Write-Host "🔄 Iniciando container do banco de dados..." -ForegroundColor Yellow
    docker-compose up -d db
    Start-Sleep -Seconds 5
}

# Copiar arquivo de migração para o container
Write-Host "📋 Copiando arquivos de migração..." -ForegroundColor Blue
docker-compose cp scripts/migrations/001_create_sensor_tables.sql db:/tmp/

# Executar migração
Write-Host "⚡ Executando migração 001_create_sensor_tables.sql..." -ForegroundColor Yellow
$result = docker-compose exec -T db psql -U dashtrack_user -d dashtrack_db -f /tmp/001_create_sensor_tables.sql

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Migração executada com sucesso!" -ForegroundColor Green
} else {
    Write-Host "❌ Erro ao executar migração" -ForegroundColor Red
    exit 1
}

# Verificar se as tabelas foram criadas
Write-Host "🔍 Verificando tabelas criadas..." -ForegroundColor Blue
docker-compose exec -T db psql -U dashtrack_user -d dashtrack_db -c "\dt sensors*"

Write-Host "🎉 Migrações concluídas!" -ForegroundColor Green