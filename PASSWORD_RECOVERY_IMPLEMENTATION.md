# Sistema de Recuperação de Senha - Implementação Completa

## 📋 Resumo

Implementação completa de recuperação de senha por email usando **SMTP Umbler** com código de 6 dígitos.

## ✅ Arquivos Criados/Modificados

### 1. **Email Service** (`internal/services/email_service.go`)
- ✅ Conexão SMTP com Umbler (smtp.umbler.com:587)
- ✅ Suporte a TLS opcional
- ✅ Templates HTML profissionais
- ✅ Logging com padrão zap.Logger
- ✅ Dois templates:
  - Código de recuperação (com timer de 15min)
  - Confirmação de senha alterada

### 2. **Password Reset Handler** (`internal/handlers/password_reset.go`)
- ✅ 3 endpoints implementados:
  - `POST /api/v1/auth/forgot-password` - Solicita código
  - `POST /api/v1/auth/verify-reset-code` - Verifica código
  - `POST /api/v1/auth/reset-password` - Redefine senha
- ✅ Rate limiting: máx 3 tentativas em 15min
- ✅ Código expira em 15 minutos
- ✅ Uso único do código
- ✅ Invalidação de sessões ao resetar
- ✅ Logging com padrão logger.Info/Error/Warn

### 3. **Database Migration** (`migrations/014_create_password_reset_tokens.up.sql`)
- ✅ Tabela `password_reset_tokens`
- ✅ Campos: token_code (6 dígitos), user_id, expires_at, used_at
- ✅ Constraints: formato do código, datas válidas
- ✅ Índices para performance
- ✅ Rollback (.down.sql)

### 4. **Config** (`internal/config/config.go` e `.env`)
- ✅ Estrutura SMTPConfig
- ✅ Variáveis de ambiente:
  ```
  SMTP_HOST=smtp.umbler.com
  SMTP_PORT=587
  SMTP_USERNAME=seu-email@seudominio.com
  SMTP_PASSWORD=sua-senha
  SMTP_FROM=noreply@seudominio.com
  SMTP_FROM_NAME=DashTrack
  SMTP_USE_TLS=true
  ```

### 5. **Routes** (`internal/routes/router.go`)
- ✅ Rotas públicas registradas
- ✅ Handler injetado no router
- ✅ EmailService inicializado

## 🔐 Fluxo de Segurança

1. **Solicitação** → Gera código aleatório de 6 dígitos
2. **Rate Limiting** → Max 3 tentativas/15min por usuário
3. **Email** → Envia código via SMTP Umbler
4. **Verificação** → Valida código + email
5. **Reset** → Marca código como usado + invalida sessões
6. **Confirmação** → Email de confirmação assíncrono

## 📝 Próximos Passos

### 1. Aplicar Migration
```bash
docker-compose exec -T dashtrack-db psql -U postgres -d dashtrack < migrations/014_create_password_reset_tokens.up.sql
```

### 2. Configurar Credenciais SMTP no .env
```env
SMTP_HOST=smtp.umbler.com
SMTP_PORT=587
SMTP_USERNAME=seu-email-real@seudominio.com
SMTP_PASSWORD=sua-senha-real
SMTP_FROM=noreply@seudominio.com
SMTP_FROM_NAME=DashTrack
SMTP_USE_TLS=true
```

### 3. Reiniciar Aplicação
```bash
docker-compose restart dashtrack-api
```

### 4. Testar Endpoints

**A) Solicitar Código:**
```powershell
$body = @{
    email = "admin@dashtrack.com"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/forgot-password" `
    -Method POST `
    -Body $body `
    -ContentType "application/json"
```

**B) Verificar Código (recebido no email):**
```powershell
$body = @{
    email = "admin@dashtrack.com"
    code = "123456"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/verify-reset-code" `
    -Method POST `
    -Body $body `
    -ContentType "application/json"
```

**C) Resetar Senha:**
```powershell
$body = @{
    email = "admin@dashtrack.com"
    code = "123456"
    new_password = "NovaSenha@2024"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/reset-password" `
    -Method POST `
    -Body $body `
    -ContentType "application/json"
```

## 🎨 Templates de Email

### Template de Código
- Header verde com ícone de cadeado
- Código em destaque (fonte grande, espaçado)
- Avisos de segurança (timer, uso único)
- Design responsivo

### Template de Confirmação
- Header azul com checkmark
- Mensagem de sucesso
- Link para suporte se não foi você
- Design profissional

## 🔍 Observabilidade

Todos os logs seguem o padrão da aplicação:
- `logger.Info()` - Operações bem-sucedidas
- `logger.Error()` - Erros com contexto
- `logger.Warn()` - Tentativas suspeitas
- Campos zap: `zap.String()`, `zap.Error()`

## ✨ Características

- ✅ Código de 6 dígitos aleatório
- ✅ Expiração em 15 minutos
- ✅ Uso único
- ✅ Rate limiting (3/15min)
- ✅ Não revela se email existe (segurança)
- ✅ Invalidação de sessões ativas
- ✅ Email assíncrono (não bloqueia)
- ✅ Templates HTML profissionais
- ✅ Logging completo
- ✅ Transações atômicas
- ✅ Audit trail (IP, user agent)

## 📊 Banco de Dados

```sql
password_reset_tokens:
  - token_id (UUID PK)
  - user_id (UUID FK → users)
  - token_code (VARCHAR(6))
  - expires_at (TIMESTAMP)
  - used_at (TIMESTAMP nullable)
  - created_at (TIMESTAMP)
  - ip_address (VARCHAR(45))
  - user_agent (TEXT)
```

## 🚀 Status: **PRONTO PARA TESTE**

Aguardando:
1. Credenciais SMTP reais do Umbler
2. Aplicação da migration
3. Testes com email real
