#!/bin/bash

# Script para gerar dados de teste para o sistema de monitoramento

echo "🚀 Gerando dados de teste para o DashTrack..."

API_URL="http://localhost:8080"
SLEEP_TIME=2

# Função para fazer login e obter token
do_login() {
    local email=$1
    local password=$2
    
    curl -s -X POST "${API_URL}/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"${email}\",\"password\":\"${password}\"}" \
        | grep -o '"token":"[^"]*"' | cut -d'"' -f4
}

# Função para fazer requisições autenticadas
authenticated_request() {
    local token=$1
    local endpoint=$2
    
    curl -s -H "Authorization: Bearer ${token}" "${API_URL}${endpoint}" > /dev/null
}

echo "📊 Simulando atividade de usuários..."

# Simular logins bem-sucedidos
for i in {1..10}; do
    echo "  Login attempt $i..."
    
    # Simular login bem-sucedido
    curl -s -X POST "${API_URL}/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"email":"master@dashtrack.com","password":"securepass"}' > /dev/null
    
    sleep $SLEEP_TIME
done

echo "🔒 Simulando tentativas de login falhadas..."

# Simular falhas de autenticação
for i in {1..5}; do
    echo "  Failed login attempt $i..."
    
    curl -s -X POST "${API_URL}/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"email":"hacker@evil.com","password":"wrongpass"}' > /dev/null
    
    sleep 1
done

echo "🔐 Simulando reset de senha..."

# Simular reset de senha
for i in {1..3}; do
    echo "  Password reset request $i..."
    
    curl -s -X POST "${API_URL}/api/v1/auth/forgot-password" \
        -H "Content-Type: application/json" \
        -d '{"email":"user@company.com"}' > /dev/null
    
    sleep 2
done

echo "📈 Simulando acesso ao dashboard..."

# Obter token válido
TOKEN=$(do_login "master@dashtrack.com" "securepass")

if [ ! -z "$TOKEN" ]; then
    # Simular acessos ao dashboard
    for i in {1..15}; do
        echo "  Dashboard access $i..."
        
        authenticated_request "$TOKEN" "/api/v1/dashboard"
        authenticated_request "$TOKEN" "/api/v1/dashboard/stats"
        authenticated_request "$TOKEN" "/health"
        
        sleep 1
    done
else
    echo "❌ Não foi possível obter token de autenticação"
fi

echo "🏥 Verificando saúde da aplicação..."

# Verificar endpoints de saúde
for i in {1..20}; do
    curl -s "${API_URL}/health" > /dev/null
    curl -s "${API_URL}/metrics" > /dev/null
    sleep 0.5
done

echo "✅ Geração de dados de teste concluída!"
echo ""
echo "🔍 Acesse os dashboards:"
echo "  📊 Grafana:    http://localhost:3000 (admin/admin)"
echo "  📈 Prometheus: http://localhost:9090"
echo "  🔍 Jaeger:     http://localhost:16686"
echo "  🏥 API Health: http://localhost:8080/health"
echo "  📊 Metrics:    http://localhost:8080/metrics"