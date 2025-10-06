# Script PowerShell para gerar dados de teste para o sistema de monitoramento

Write-Host "🚀 Gerando dados de teste para o DashTrack..." -ForegroundColor Green

$API_URL = "http://localhost:8080"
$SLEEP_TIME = 2

# Função para fazer login e obter token
function Do-Login {
    param($email, $password)
    
    $body = @{
        email = $email
        password = $password
    } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "$API_URL/api/v1/auth/login" -Method POST -ContentType "application/json" -Body $body
        return $response.token
    }
    catch {
        return $null
    }
}

# Função para fazer requisições autenticadas
function Invoke-AuthenticatedRequest {
    param($token, $endpoint)
    
    $headers = @{
        "Authorization" = "Bearer $token"
    }
    
    try {
        Invoke-RestMethod -Uri "$API_URL$endpoint" -Headers $headers | Out-Null
    }
    catch {
        # Ignorar erros silenciosamente para simulação
    }
}

Write-Host "📊 Simulando atividade de usuários..." -ForegroundColor Yellow

# Simular logins bem-sucedidos
for ($i = 1; $i -le 10; $i++) {
    Write-Host "  Login attempt $i..." -ForegroundColor Cyan
    
    $body = @{
        email = "master@dashtrack.com"
        password = "securepass"
    } | ConvertTo-Json
    
    try {
        Invoke-RestMethod -Uri "$API_URL/api/v1/auth/login" -Method POST -ContentType "application/json" -Body $body | Out-Null
    }
    catch {
        # Ignorar erros
    }
    
    Start-Sleep -Seconds $SLEEP_TIME
}

Write-Host "🔒 Simulando tentativas de login falhadas..." -ForegroundColor Red

# Simular falhas de autenticação
for ($i = 1; $i -le 5; $i++) {
    Write-Host "  Failed login attempt $i..." -ForegroundColor DarkRed
    
    $body = @{
        email = "hacker@evil.com"
        password = "wrongpass"
    } | ConvertTo-Json
    
    try {
        Invoke-RestMethod -Uri "$API_URL/api/v1/auth/login" -Method POST -ContentType "application/json" -Body $body | Out-Null
    }
    catch {
        # Erro esperado
    }
    
    Start-Sleep -Seconds 1
}

Write-Host "🔐 Simulando reset de senha..." -ForegroundColor Magenta

# Simular reset de senha
for ($i = 1; $i -le 3; $i++) {
    Write-Host "  Password reset request $i..." -ForegroundColor DarkMagenta
    
    $body = @{
        email = "user@company.com"
    } | ConvertTo-Json
    
    try {
        Invoke-RestMethod -Uri "$API_URL/api/v1/auth/forgot-password" -Method POST -ContentType "application/json" -Body $body | Out-Null
    }
    catch {
        # Ignorar erros
    }
    
    Start-Sleep -Seconds 2
}

Write-Host "📈 Simulando acesso ao dashboard..." -ForegroundColor Blue

# Obter token válido
$TOKEN = Do-Login "master@dashtrack.com" "securepass"

if ($TOKEN) {
    # Simular acessos ao dashboard
    for ($i = 1; $i -le 15; $i++) {
        Write-Host "  Dashboard access $i..." -ForegroundColor DarkBlue
        
        Invoke-AuthenticatedRequest $TOKEN "/api/v1/dashboard"
        Invoke-AuthenticatedRequest $TOKEN "/api/v1/dashboard/stats"
        Invoke-AuthenticatedRequest $TOKEN "/health"
        
        Start-Sleep -Seconds 1
    }
}
else {
    Write-Host "❌ Não foi possível obter token de autenticação" -ForegroundColor Red
}

Write-Host "🏥 Verificando saúde da aplicação..." -ForegroundColor White

# Verificar endpoints de saúde
for ($i = 1; $i -le 20; $i++) {
    try {
        Invoke-RestMethod -Uri "$API_URL/health" | Out-Null
        Invoke-RestMethod -Uri "$API_URL/metrics" | Out-Null
    }
    catch {
        # Ignorar erros
    }
    Start-Sleep -Milliseconds 500
}

Write-Host "✅ Geração de dados de teste concluída!" -ForegroundColor Green
Write-Host ""
Write-Host "🔍 Acesse os dashboards:" -ForegroundColor White
Write-Host "  📊 Grafana:    http://localhost:3000 (admin/admin)" -ForegroundColor Cyan
Write-Host "  📈 Prometheus: http://localhost:9090" -ForegroundColor Yellow
Write-Host "  🔍 Jaeger:     http://localhost:16686" -ForegroundColor Magenta
Write-Host "  🏥 API Health: http://localhost:8080/health" -ForegroundColor Green
Write-Host "  📊 Metrics:    http://localhost:8080/metrics" -ForegroundColor Blue