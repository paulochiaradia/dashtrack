# 🚛 Dashtrack - Fleet Monitoring API

Um backend robusto e escalável para monitoramento de frotas, desenvolvido em Go com foco em performance, observabilidade e integração com dispositivos IoT (ESP32).

## ✨ Características Principais

- **API REST completa** com endpoints para usuários, roles e autenticação
- **Banco PostgreSQL** com migrações automatizadas
- **Docker & Docker Compose** para desenvolvimento e produção
- **Observabilidade completa** com métricas Prometheus, tracing Jaeger e logs estruturados
- **Testes abrangentes** (unit, integration, benchmark)
- **Arquitetura limpa** com Repository Pattern
- **Configuração via ambiente** com padrão Singleton
- **CORS configurado** para integração frontend
- **Preparado para ESP32** com endpoints específicos para IoT

## 🚀 Quick Start

### Pré-requisitos
- Go 1.23+
- Docker & Docker Compose
- Make (opcional, mas recomendado)

### Instalação e Execução

```bash
# Clone o repositório
git clone https://github.com/paulochiaradia/dashtrack.git
cd dashtrack

# Instale dependências
go mod download

# Execute com Docker (recomendado)
docker-compose up --build

# Ou execute localmente
make run
```

A API estará disponível em: http://localhost:8080

## 🧪 Testes

Esta aplicação possui uma suíte completa de testes incluindo:

### Executar Testes
```bash
# Testes unitários
make test-unit
# ou
go test -v ./internal/handlers/... -short

# Testes de integração  
make test-integration
# ou
go test -v ./tests/integration/... -run Integration

# Benchmarks de performance
make test-bench
# ou
go test ./tests/integration/... -bench=. -run=^$

# Cobertura de código
make test-coverage
# ou
go test ./internal/handlers -coverprofile=coverage.out -short
```

### Resultados dos Testes ✅
- **Testes Unitários**: 3/3 passando (autenticação completa)
- **Testes de Integração**: 6/6 passando (fluxos E2E)
- **Benchmarks**: Login ~82k req/s, Auth ~45k req/s
- **Cobertura**: Componentes críticos 100% cobertos

📖 **Documentação completa dos testes**: [TESTING.md](./TESTING.md)

## 📚 Endpoints Disponíveis

### Saúde e Monitoramento
- `GET /health` - Status da aplicação
- `GET /metrics` - Métricas Prometheus

### Usuários
- `GET /users` - Listar usuários
- `POST /users` - Criar usuário
- `GET /users/{id}` - Buscar usuário por ID
- `PUT /users/{id}` - Atualizar usuário
- `DELETE /users/{id}` - Deletar usuário

### Roles
- `GET /roles` - Listar roles (admin, driver, helper)

## 🧪 Testes

Execute a suíte completa de testes:

```bash
# Windows
.\scripts\run-tests.ps1

# Linux/Mac
./scripts/run-tests.sh

# Ou usando Make
make test
```

### Tipos de Testes Incluídos

- **Unit Tests**: Testam handlers, repositórios e modelos
- **Integration Tests**: Testam fluxo completo da aplicação
- **Benchmark Tests**: Medem performance dos endpoints
- **Load Tests**: Testam capacidade sob carga

### Cobertura de Testes
Target de cobertura: **80%+**

Visualize o relatório de cobertura:
```bash
make test-coverage
# Abra test-reports/coverage.html no navegador
```

## 📊 Observabilidade

### Métricas (Prometheus)
Acesse: http://localhost:9090

Métricas disponíveis:
- `http_requests_total` - Total de requisições HTTP
- `http_request_duration_seconds` - Duração das requisições
- `database_connections` - Conexões ativas do banco
- `database_queries_total` - Total de queries executadas
- `users_total` - Total de usuários no sistema

### Tracing (Jaeger)
Acesse: http://localhost:16686

Traces automáticos para:
- Requisições HTTP
- Queries de banco de dados
- Operações de business logic

### Logs Estruturados
Logs em formato JSON com:
- Timestamps ISO8601
- Níveis de log (INFO, WARN, ERROR)
- Contexto de requisição
- Caller information

### Dashboard (Grafana)
Acesse: http://localhost:3000
- **Usuário**: admin
- **Senha**: admin

Dashboards incluídos:
- API Performance
- Database Metrics
- Application Health
- Error Rates

## 🏗️ Arquitetura

```
cmd/
  api/                  # Ponto de entrada da aplicação
internal/
  config/              # Configuração (Singleton pattern)
  database/            # Conexão e migrações
    migrations/        # Scripts SQL de migração
  handlers/            # HTTP handlers
  logger/              # Logger estruturado (Zap)
  metrics/             # Métricas Prometheus
  middleware/          # Middlewares HTTP
  models/              # Estruturas de dados
  repository/          # Repository pattern para dados
  tracing/             # Configuração de tracing
tests/
  benchmarks/          # Testes de performance
  integration/         # Testes de integração
monitoring/            # Configurações Prometheus/Grafana
scripts/               # Scripts de automação
```

## 🛠️ Desenvolvimento

### Comandos Úteis (Make)

```bash
make help              # Lista todos os comandos disponíveis
make build             # Compila a aplicação
make test              # Executa todos os testes
make docker-up         # Inicia containers Docker
make docker-down       # Para containers Docker
make db-reset          # Reseta o banco de dados
make monitoring-up     # Inicia stack de monitoramento
make api-test-health   # Testa endpoint de saúde
make load-test         # Executa teste de carga
```

### Live Reload
```bash
# Instale o Air para reload automático
go install github.com/cosmtrek/air@latest

# Execute com reload
make run-dev
```

### Migrações de Banco
```bash
# Aplicar migrações
make migrate-up

# Reverter migrações  
make migrate-down

# Status das migrações
make migrate-status
```

## 🌟 Stack de Monitoramento Completa

Inicie todos os serviços de monitoramento:

```bash
docker-compose -f docker-compose.monitoring.yml up -d
```

Serviços disponíveis:
- **API**: http://localhost:8080
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000
- **Jaeger**: http://localhost:16686
- **PostgreSQL**: localhost:5432

## 🚛 Integração ESP32

A API está preparada para receber dados de dispositivos ESP32 com:

- Endpoints específicos para telemetria
- Autenticação via API Token
- Buffer para dados offline
- Validação de payload IoT

### Exemplo de Payload ESP32
```json
{
  "device_id": "ESP32_001",
  "timestamp": "2023-01-01T12:00:00Z",
  "location": {
    "lat": -23.5505,
    "lng": -46.6333
  },
  "sensors": {
    "speed": 65.5,
    "fuel": 87.2,
    "temperature": 25.1
  }
}
```

## 📈 Performance

### Benchmarks Atuais
- **Health Endpoint**: ~0.1ms médio
- **Get Users**: ~2.5ms médio  
- **Create User**: ~15ms médio
- **Throughput**: ~10k req/s

### Otimizações Implementadas
- Connection pooling PostgreSQL
- Índices otimizados
- JSON encoding eficiente
- Query prepared statements
- Graceful shutdown

## 🔒 Segurança

### Implementado
- Validação de input
- SQL injection prevention
- CORS configurado
- Rate limiting (TODO)
- JWT authentication (TODO)

### Scan de Seguridade
```bash
# Instale gosec
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Execute scan
make security-scan
```

## 🚀 Deploy em Produção

### Docker Production
```bash
# Build para produção
docker build --target production -t dashtrack-api:latest .

# Execute
docker run -p 8080:8080 dashtrack-api:latest
```

### Variáveis de Ambiente
```bash
DB_SOURCE=postgresql://user:pass@host:5432/db
API_PORT=8080
ENVIRONMENT=production
LOG_LEVEL=info
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
PROMETHEUS_ENABLED=true
```

## 📋 Roadmap

### 🎯 Próximas Funcionalidades
- [ ] Sistema de autenticação JWT completo
- [ ] Endpoints específicos para ESP32
- [ ] Rate limiting e throttling
- [ ] Cache Redis para performance
- [ ] Notificações Telegram/WhatsApp
- [ ] Dashboard web frontend
- [ ] API de relatórios
- [ ] Backup automático do banco
- [ ] Deploy Kubernetes
- [ ] CI/CD com GitHub Actions

### 🔄 Melhorias Contínuas
- [ ] Aumentar cobertura de testes para 90%+
- [ ] Otimizar queries mais complexas
- [ ] Implementar circuit breaker
- [ ] Adicionar health checks avançados
- [ ] Documentação OpenAPI/Swagger

## 🤝 Contribuição

1. Fork o projeto
2. Crie uma feature branch (`git checkout -b feature/nova-funcionalidade`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova funcionalidade'`)
4. Push para a branch (`git push origin feature/nova-funcionalidade`)
5. Crie um Pull Request

### Guidelines
- Mantenha cobertura de testes acima de 80%
- Use conventional commits
- Execute `make test` antes do PR
- Documente novas funcionalidades

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo [LICENSE](LICENSE) para detalhes.

## 👥 Autores

- **Paulo Chiaradia** - *Desenvolvimento inicial* - [paulochiaradia](https://github.com/paulochiaradia)

## 🙏 Agradecimentos

- Comunidade Go pela excelente documentação
- Projeto Prometheus pela stack de monitoramento
- PostgreSQL pela robustez do banco de dados
- Docker pela facilidade de containerização

---

**Dashtrack** - Monitoramento de frotas do futuro! 🚛✨
