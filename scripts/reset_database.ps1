# DashTrack - Reset Database Script
# Este script reseta completamente o banco de dados e configura dados iniciais

param(
    [switch]$SkipConfirmation
)

$ErrorActionPreference = "Stop"

Write-Host "🚀 DashTrack Database Reset Script" -ForegroundColor Cyan
Write-Host "=================================" -ForegroundColor Cyan

if (-not $SkipConfirmation) {
    Write-Host "⚠️  ATENÇÃO: Este script irá:" -ForegroundColor Yellow
    Write-Host "   - Parar todos os containers" -ForegroundColor Red
    Write-Host "   - APAGAR TODOS OS DADOS do banco" -ForegroundColor Red
    Write-Host "   - Recriar banco limpo" -ForegroundColor Green
    Write-Host "   - Configurar dados iniciais" -ForegroundColor Green
    
    $confirmation = Read-Host "`n🤔 Deseja continuar? (y/N)"
    if ($confirmation -ne 'y' -and $confirmation -ne 'Y') {
        Write-Host "❌ Operação cancelada pelo usuário." -ForegroundColor Red
        exit 0
    }
}

try {
    # 1. Parar containers
    Write-Host "`n🛑 Parando containers..." -ForegroundColor Yellow
    docker-compose down
    
    # 2. Remover volume do PostgreSQL
    Write-Host "🗑️  Removendo dados do banco..." -ForegroundColor Red
    docker volume rm dashtrack_postgres_data -f 2>$null
    
    # 3. Subir aplicação
    Write-Host "🚀 Subindo aplicação com banco limpo..." -ForegroundColor Green
    docker-compose up -d
    
    # 4. Aguardar inicialização
    Write-Host "⏳ Aguardando aplicação inicializar..." -ForegroundColor Cyan
    Start-Sleep 15
    
    # 5. Verificar se aplicação está rodando
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -TimeoutSec 10
        if ($response.StatusCode -eq 200) {
            Write-Host "✅ Aplicação rodando!" -ForegroundColor Green
        }
    }
    catch {
        Write-Host "⚠️  Aplicação ainda inicializando, aguarde mais alguns segundos..." -ForegroundColor Yellow
    }
    
    # 6. Executar setup inicial do banco
    Write-Host "📊 Configurando dados iniciais..." -ForegroundColor Cyan
    docker-compose exec -T db psql -U user -d dashtrack -f - < scripts/setup_initial_data.sql
    
    # 7. Verificações finais
    Write-Host "`n🔍 Verificações finais..." -ForegroundColor Cyan
    
    Write-Host "   📋 Tabelas criadas:" -ForegroundColor White
    docker-compose exec -T db psql -U user -d dashtrack -c "\dt" -q
    
    Write-Host "`n   🏢 Empresas:" -ForegroundColor White
    docker-compose exec -T db psql -U user -d dashtrack -c "SELECT name, slug, email FROM companies;" -q
    
    Write-Host "`n   👤 Usuários:" -ForegroundColor White
    docker-compose exec -T db psql -U user -d dashtrack -c "SELECT u.name, u.email, r.name as role FROM users u JOIN roles r ON u.role_id = r.id;" -q
    
    # 8. Teste de login
    Write-Host "`n🧪 Testando login..." -ForegroundColor Cyan
    $loginBody = @{
        email = "master@dashtrack.com"
        password = "password"
    } | ConvertTo-Json
    
    try {
        $loginResponse = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/auth/login" -Method POST -Body $loginBody -ContentType "application/json" -UseBasicParsing
        if ($loginResponse.StatusCode -eq 200) {
            Write-Host "✅ Login funcionando!" -ForegroundColor Green
        }
    }
    catch {
        Write-Host "⚠️  Erro no teste de login: $($_.Exception.Message)" -ForegroundColor Yellow
    }
    
    # 9. Sucesso
    Write-Host "`n🎉 SETUP CONCLUÍDO COM SUCESSO!" -ForegroundColor Green
    Write-Host "=================================" -ForegroundColor Green
    Write-Host "📧 Email: master@dashtrack.com" -ForegroundColor White
    Write-Host "🔑 Senha: password" -ForegroundColor White
    Write-Host "🌐 URL: http://localhost:8080" -ForegroundColor White
    Write-Host "📚 Health: http://localhost:8080/health" -ForegroundColor White
    Write-Host "📊 Metrics: http://localhost:8080/metrics" -ForegroundColor White
    
    Write-Host "`n🔧 Próximos passos:" -ForegroundColor Cyan
    Write-Host "   1. Teste o health check" -ForegroundColor White
    Write-Host "   2. Faça login com as credenciais acima" -ForegroundColor White
    Write-Host "   3. Crie empresas e usuários conforme necessário" -ForegroundColor White
    Write-Host "   4. Execute seus testes com Postman" -ForegroundColor White

}
catch {
    Write-Host "`n❌ ERRO durante a execução:" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    Write-Host "`n🔧 Comandos para diagnóstico:" -ForegroundColor Yellow
    Write-Host "   docker-compose logs api" -ForegroundColor Gray
    Write-Host "   docker-compose ps" -ForegroundColor Gray
    exit 1
}