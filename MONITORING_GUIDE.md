# 🚀 DashTrack - Stack de Monitoramento Exemplar

## 📊 Visão Geral

Esta é uma implementação completa e profissional de observabilidade para o sistema DashTrack, focando inicialmente no sistema de autenticação e usuários.

## 🛠️ Componentes da Stack

### 1. **Prometheus** - Coleta de Métricas
- **Porta**: 9090
- **Função**: Coleta métricas de time-series da aplicação
- **Configuração**: `monitoring/prometheus.yml`
- **Alertas**: `monitoring/alerts/dashtrack_rules.yml`

### 2. **Grafana** - Visualização
- **Porta**: 3000
- **Login**: admin/admin
- **Função**: Dashboards ricos e alertas visuais
- **Dashboards**:
  - API Overview (`dashtrack-api`)
  - Autenticação & Segurança (`dashtrack-auth`)
  - Infraestrutura & Performance (`dashtrack-infra`)
  - Logs & Debugging (`dashtrack-logs`)

### 3. **Jaeger** - Tracing Distribuído
- **Porta**: 16686
- **Função**: Rastreamento de requisições e performance
- **Integração**: Automática via OpenTelemetry

### 4. **Loki + Promtail** - Agregação de Logs
- **Loki Porta**: 3100
- **Função**: Coleta, indexação e pesquisa de logs
- **Integração**: Visualização via Grafana

## 📈 Métricas Implementadas

### HTTP & API
- `http_requests_total` - Total de requisições por endpoint
- `http_request_duration_seconds` - Latência das requisições
- Distribuição de status codes (2xx, 4xx, 5xx)

### Autenticação & Segurança
- `auth_success_total` - Logins bem-sucedidos por role
- `auth_failures_total` - Falhas de autenticação por motivo
- `password_reset_requests_total` - Solicitações de reset

### Banco de Dados
- `db_connections_active` - Conexões ativas
- `db_query_duration_seconds` - Performance das queries
- `db_queries_total` - Total de queries por operação

### Aplicação & Negócio
- `active_users_total` - Usuários ativos
- `user_sessions_total` - Sessões ativas
- `dashboard_views_total` - Visualizações por tipo
- `companies_total` - Total de empresas

### Sistema (Go Runtime)
- `go_memstats_alloc_bytes` - Uso de memória
- `go_goroutines` - Goroutines ativas
- `go_gc_duration_seconds` - Performance do GC

## 🚨 Alertas Configurados

### Críticos
- **APIDown**: API indisponível por >1min
- **HighAuthFailures**: >1 falha/seg (possível ataque)

### Warnings
- **HighErrorRate**: Taxa 5xx >5%
- **HighLatency**: P95 >500ms
- **HighDatabaseConnections**: >15 conexões
- **HighMemoryUsage**: >500MB
- **HighGoroutines**: >1000 goroutines

### Informativos
- **LowUserActivity**: <1 usuário ativo por 30min
- **HighUserRegistration**: >100 sessões/hora

## 🔧 Como Usar

### Iniciar a Stack Completa
```bash
docker-compose -f docker-compose.monitoring.yml up -d
```

### Gerar Dados de Teste
```powershell
# PowerShell
.\scripts\generate_test_data.ps1

# Bash
chmod +x scripts/generate_test_data.sh
./scripts/generate_test_data.sh
```

### Acessar Interfaces
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Jaeger**: http://localhost:16686
- **API**: http://localhost:8080
- **Métricas**: http://localhost:8080/metrics

## 📊 Dashboards Principais

### 1. DashTrack - API Overview
Visão geral da saúde da API:
- Taxa de requisições por endpoint
- Distribuição de status HTTP
- Latência (p50, p90, p95)
- Conexões do banco
- Memória e goroutines

### 2. DashTrack - Autenticação & Segurança
Foco em segurança:
- Taxa de login (sucesso vs falha)
- Falhas de autenticação acumuladas
- Usuários ativos
- Latência de autenticação
- Sessões ativas

### 3. DashTrack - Infraestrutura & Performance
Performance técnica:
- Requisições por endpoint
- Performance do banco de dados
- Uso de recursos (CPU, memória)
- Goroutines e GC

### 4. DashTrack - Logs & Debugging
Análise de logs:
- Logs da aplicação em tempo real
- Logs por nível (ERROR, WARN, INFO)
- Logs específicos de autenticação
- Correlação com métricas

## 🔍 Casos de Uso

### Para Desenvolvedores
1. **Debug de Performance**: Use Jaeger para trace de requests lentos
2. **Análise de Erros**: Correlacione logs no Grafana com métricas
3. **Monitoramento Local**: Prometheus + Grafana para desenvolvimento

### Para DevOps/SRE
1. **Alertas Proativos**: Configure notificações via Grafana
2. **Capacity Planning**: Monitore uso de recursos
3. **SLA Monitoring**: Dashboards para availability e latência

### Para Product/Negócio
1. **Métricas de Adoção**: Usuários ativos, sessões
2. **Análise de Uso**: Dashboard views por tipo
3. **Health do Produto**: Correlação entre métricas técnicas e de negócio

## 🚀 Próximos Passos (ESP32/IoT)

Quando implementarmos o sistema IoT, a stack está preparada para:

### Métricas IoT
- `esp32_devices_connected` - Dispositivos conectados
- `sensor_readings_total` - Leituras de sensores
- `data_processing_duration` - Latência de processamento
- `alerts_triggered_total` - Alertas de sensores

### Dashboards IoT
- Mapa de dispositivos em tempo real
- Gráficos de telemetria por sensor
- Alertas de threshold de sensores
- Analytics de padrões de dados

### Alertas IoT
- Dispositivos offline
- Valores anômalos de sensores
- Falhas de comunicação
- Thresholds de negócio

## 📝 Configuração de Produção

### Segurança
1. Configurar autenticação no Grafana
2. HTTPS para todas as interfaces
3. Firewall para portas de monitoramento
4. Backup das configurações

### Escalabilidade
1. Prometheus federation para múltiplos targets
2. Grafana clustering
3. Loki sharding para logs em volume
4. Alertmanager para notificações

### Retenção
1. Prometheus: 15 dias (configurável)
2. Loki: 30 dias (configurável)
3. Jaeger: 7 dias (configurável)
4. Backup automático de dashboards

## 🎯 Benefícios da Stack

### Observabilidade Completa
- **Métricas**: Performance e negócio
- **Logs**: Debug e auditoria
- **Traces**: Performance de requests

### Operacional
- **Alertas Proativos**: Problemas antes dos usuários
- **Troubleshooting Rápido**: Correlação entre dados
- **Capacity Planning**: Dados para escalar

### Desenvolvimento
- **Feedback Rápido**: Métricas em real-time
- **Debug Eficiente**: Traces detalhados
- **Quality Gates**: Métricas como CI/CD

## 📞 Troubleshooting

### Loki não inicia
```bash
# Verificar logs
docker logs dashtrack-loki

# Recriar com nova configuração
docker-compose -f docker-compose.monitoring.yml down
docker-compose -f docker-compose.monitoring.yml up -d loki
```

### Grafana sem dados
1. Verificar datasources em `/datasources`
2. Confirmar que Prometheus está coletando: http://localhost:9090/targets
3. Recarregar configurações no Grafana

### Métricas não aparecem
1. Verificar `/metrics` da aplicação
2. Confirmar middleware de métricas ativo
3. Verificar logs do Prometheus

---

**Esta stack é uma implementação de nível produção, preparada para escalar com o crescimento do DashTrack!** 🚀