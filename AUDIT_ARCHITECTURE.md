# 🔍 Audit Logs - Arquitetura Híbrida com Observabilidade

## 🎯 Estratégia: Três Camadas de Auditoria

### **Camada 1: Database Audit Logs** 📊
**Uso**: Auditoria de negócio, compliance, consultas históricas
- Armazena ações de usuários (CRUD de entidades)
- Queries complexas e relatórios
- Retenção longa (1-2 anos)
- Performance: Indexed queries

### **Camada 2: Prometheus Metrics** 📈
**Uso**: Métricas e alertas em tempo real
- Contadores de ações por tipo
- Taxa de erros
- Performance de operações
- Alertas automáticos

### **Camada 3: Jaeger Traces** 🔎
**Uso**: Debugging, performance tracing
- Rastreamento distribuído
- Latência de operações
- Correlação de requests
- Troubleshooting detalhado

---

## 🏗️ Arquitetura de Implementação

```
┌─────────────────────────────────────────────────────────────┐
│                     HTTP Request                            │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│              Audit Middleware (Interceptor)                 │
│  - Captura contexto (user, action, resource)               │
│  - Inicia Jaeger span                                       │
│  - Incrementa Prometheus metrics                            │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                   Business Handler                          │
│  - Executa lógica de negócio                               │
│  - Retorna resultado                                        │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│              Audit Service (Async)                          │
│  - Salva audit log no banco (goroutine)                    │
│  - Não bloqueia o request                                   │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
              ┌────────┴────────┐
              │                 │
              ▼                 ▼
    ┌─────────────┐   ┌─────────────────┐
    │  PostgreSQL │   │   Prometheus    │
    │  (DB Logs)  │   │   (Metrics)     │
    └─────────────┘   └─────────────────┘
              │                 │
              └────────┬────────┘
                       │
                       ▼
              ┌─────────────────┐
              │     Grafana     │
              │   (Dashboards)  │
              └─────────────────┘
                       │
                       ▼
              ┌─────────────────┐
              │     Jaeger      │
              │    (Traces)     │
              └─────────────────┘
```

---

## 📋 Implementação Detalhada

### **1. Modelo de Audit Log (Estendido)**

```go
// AuditLog representa um log de auditoria completo
type AuditLog struct {
    ID            uuid.UUID  `json:"id" db:"id"`
    UserID        uuid.UUID  `json:"user_id" db:"user_id"`
    UserEmail     string     `json:"user_email" db:"user_email"`
    CompanyID     *uuid.UUID `json:"company_id" db:"company_id"`
    
    // Ação
    Action        string     `json:"action" db:"action"`           // "CREATE", "UPDATE", "DELETE", "READ", "LOGIN"
    Resource      string     `json:"resource" db:"resource"`       // "user", "vehicle", "team", etc
    ResourceID    *uuid.UUID `json:"resource_id" db:"resource_id"`
    
    // Contexto
    Method        string     `json:"method" db:"method"`           // "GET", "POST", "PUT", "DELETE"
    Path          string     `json:"path" db:"path"`               // "/api/v1/users/123"
    IPAddress     string     `json:"ip_address" db:"ip_address"`
    UserAgent     string     `json:"user_agent" db:"user_agent"`
    
    // Dados
    Changes       *string    `json:"changes" db:"changes"`         // JSON com before/after
    Metadata      *string    `json:"metadata" db:"metadata"`       // JSON com info adicional
    
    // Resultado
    Success       bool       `json:"success" db:"success"`
    ErrorMessage  *string    `json:"error_message" db:"error_message"`
    StatusCode    int        `json:"status_code" db:"status_code"`
    Duration      int64      `json:"duration_ms" db:"duration_ms"` // milissegundos
    
    // Trace
    TraceID       *string    `json:"trace_id" db:"trace_id"`       // Jaeger trace ID
    SpanID        *string    `json:"span_id" db:"span_id"`         // Jaeger span ID
    
    CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}
```

### **2. Prometheus Metrics**

```go
// Métricas customizadas para audit
var (
    AuditActionsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "audit_actions_total",
            Help: "Total de ações auditadas por tipo",
        },
        []string{"action", "resource", "user_role", "success"},
    )
    
    AuditActionDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "audit_action_duration_seconds",
            Help:    "Duração das ações auditadas",
            Buckets: prometheus.DefBuckets,
        },
        []string{"action", "resource"},
    )
    
    AuditErrorsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "audit_errors_total",
            Help: "Total de erros nas ações auditadas",
        },
        []string{"action", "resource", "error_type"},
    )
)
```

### **3. Audit Middleware**

```go
// Middleware que captura todas as ações
func AuditMiddleware(auditService *AuditService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Skip health checks e metrics
        if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/metrics" {
            c.Next()
            return
        }
        
        start := time.Now()
        
        // Inicia Jaeger span
        ctx, span := tracing.StartSpan(c.Request.Context(), "audit:"+c.Request.Method+" "+c.Request.URL.Path)
        defer span.End()
        
        // Captura user context
        userID, _ := c.Get("user_id")
        userEmail, _ := c.Get("user_email")
        userRole, _ := c.Get("user_role")
        
        // Processa request
        c.Next()
        
        duration := time.Since(start)
        
        // Incrementa Prometheus metrics
        labels := prometheus.Labels{
            "action":    inferAction(c.Request.Method),
            "resource":  inferResource(c.Request.URL.Path),
            "user_role": fmt.Sprintf("%v", userRole),
            "success":   fmt.Sprintf("%t", c.Writer.Status() < 400),
        }
        AuditActionsTotal.With(labels).Inc()
        AuditActionDuration.With(prometheus.Labels{
            "action":   inferAction(c.Request.Method),
            "resource": inferResource(c.Request.URL.Path),
        }).Observe(duration.Seconds())
        
        // Salva no banco (async, não bloqueia)
        go auditService.Log(ctx, AuditLog{
            UserID:      userID.(uuid.UUID),
            UserEmail:   userEmail.(string),
            Action:      inferAction(c.Request.Method),
            Resource:    inferResource(c.Request.URL.Path),
            Method:      c.Request.Method,
            Path:        c.Request.URL.Path,
            IPAddress:   c.ClientIP(),
            UserAgent:   c.Request.UserAgent(),
            Success:     c.Writer.Status() < 400,
            StatusCode:  c.Writer.Status(),
            Duration:    duration.Milliseconds(),
            TraceID:     getTraceID(span),
            SpanID:      getSpanID(span),
        })
    }
}
```

---

## 📊 Dashboards Grafana

### **Dashboard 1: Audit Overview**
- Total de ações por hora/dia
- Taxa de sucesso vs erro
- Ações por usuário
- Ações por resource type
- Top 10 usuários mais ativos

### **Dashboard 2: Security Monitoring**
- Tentativas de login falhadas
- Acessos suspeitos (múltiplos IPs)
- Mudanças de permissões
- Exclusões de dados
- Acessos fora de horário

### **Dashboard 3: Performance**
- Latência por tipo de ação
- Ações mais lentas (p99)
- Volume de requests por endpoint
- Correlação com Jaeger traces

---

## 🚀 Benefícios da Abordagem Híbrida

### **Banco de Dados**
✅ Queries complexas (filtros, agregações)
✅ Relatórios históricos
✅ Compliance e auditoria legal
✅ Backup e restore

### **Prometheus**
✅ Alertas em tempo real
✅ Métricas de performance
✅ Integração com Grafana
✅ Baixa latência

### **Jaeger**
✅ Troubleshooting detalhado
✅ Correlação de traces
✅ Performance analysis
✅ Debugging distribuído

---

## 📝 Endpoints de Audit Logs

```
GET  /api/v1/audit/logs              - Lista logs (filtros: user, action, resource, date)
GET  /api/v1/audit/logs/:id          - Detalhes de um log
GET  /api/v1/audit/stats             - Estatísticas agregadas
GET  /api/v1/audit/timeline          - Timeline de ações
GET  /api/v1/audit/users/:id/logs    - Logs de um usuário
GET  /api/v1/audit/resources/:type   - Logs por tipo de recurso
GET  /api/v1/audit/export            - Exportar logs (CSV/JSON)
```

---

## 🎯 Próximos Passos

1. ✅ Criar migration para tabela `audit_logs`
2. ✅ Implementar modelo `AuditLog`
3. ✅ Criar `AuditRepository`
4. ✅ Criar `AuditService` com métodos async
5. ✅ Implementar `AuditMiddleware`
6. ✅ Adicionar métricas Prometheus
7. ✅ Criar handlers para consulta
8. ✅ Criar dashboards Grafana
9. ✅ Documentar e testar

**Pronto para começar a implementação?** 🚀
