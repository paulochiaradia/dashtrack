#!/bin/bash

# Script para executar migrações de banco de dados
# Usage: ./run-migrations.sh

echo "🚀 Executando migrações do banco de dados..."

# Verificar se o Docker está rodando
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker não está rodando. Por favor, inicie o Docker Desktop."
    exit 1
fi

# Verificar se o container do banco está rodando
if ! docker-compose ps db | grep -q "Up"; then
    echo "🔄 Iniciando container do banco de dados..."
    docker-compose up -d db
    sleep 5
fi

# Copiar arquivo de migração para o container
echo "📋 Copiando arquivos de migração..."
docker-compose cp scripts/migrations/001_create_sensor_tables.sql db:/tmp/

# Executar migração
echo "⚡ Executando migração 001_create_sensor_tables.sql..."
docker-compose exec -T db psql -U dashtrack_user -d dashtrack_db -f /tmp/001_create_sensor_tables.sql

if [ $? -eq 0 ]; then
    echo "✅ Migração executada com sucesso!"
else
    echo "❌ Erro ao executar migração"
    exit 1
fi

# Verificar se as tabelas foram criadas
echo "🔍 Verificando tabelas criadas..."
docker-compose exec -T db psql -U dashtrack_user -d dashtrack_db -c "\dt sensors*"

echo "🎉 Migrações concluídas!"