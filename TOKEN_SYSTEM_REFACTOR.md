# Refatoração do Sistema de Tokens

## 📋 Resumo

Refatoração completa do sistema de autenticação para usar exclusivamente o `TokenService`, removendo o `JWTManager` antigo. O novo sistema oferece melhor segurança com gerenciamento de sessões no banco de dados.

## 🔄 Mudanças Principais

### 1. Sistema Antigo (JWTManager) - REMOVIDO ❌

**Problemas:**
- Tokens stateless sem controle de sessão
- Impossível revogar tokens antes da expiração
- Sem rastreamento de dispositivos/IPs
- Dois sistemas de token conflitantes
- Refresh token não funcionava corretamente

### 2. Sistema Novo (TokenService) - IMPLEMENTADO ✅

**Vantagens:**
- ✅ Sessões armazenadas no banco de dados
- ✅ Controle total sobre tokens ativos
- ✅ Revogação de tokens/sessões em tempo real
- ✅ Rastreamento de IP e User-Agent
- ✅ Limite de sessões simultâneas (máx. 3 por usuário)
- ✅ Refresh token funcional
- ✅ Sistema unificado de autenticação

## 📁 Arquivos Modificados

### Handlers
- **`internal/handlers/auth.go`**
  - Removido campo `jwtManager`
  - `LoginGin()`: Agora usa `tokenService.GenerateTokenPair()`
  - `RefreshTokenGin()`: Usa `tokenService.RefreshTokenPair()`
  - `Login()`: Atualizado para usar tokenService
  - `RefreshToken()`: Atualizado para usar tokenService
  - Adicionado rastreamento de IP e User-Agent em todas as operações

### Middleware
- **`internal/middleware/auth.go`**
  - Substituído `jwtManager` por `tokenService`
  - `RequireAuth()`: Agora valida tokens via `tokenService.ValidateAccessToken()`
  - Validação de sessão no banco de dados em cada requisição

### Rotas
- **`internal/routes/router.go`**
  - Removido `jwtManager` do struct Router
  - Removido import do package `auth`
  - Middleware de autenticação agora usa `tokenService`

- **Arquivos de rotas atualizados:**
  - `admin.go`
  - `company_admin.go`
  - `manager.go`
  - `master.go`
  - `multitenant.go`
  - `protected.go`
  - `security.go`
  - `sensor.go`
  - `sessions.go`
  - `system.go`

## 🔐 Como Funciona o Novo Sistema

### 1. Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**Resposta:**
```json
{
  "user": {...},
  "access_token": "eyJ...",
  "refresh_token": "eyJ...",
  "expires_in": 900
}
```

**O que acontece:**
1. Valida credenciais
2. Gera par de tokens (access + refresh)
3. Cria sessão no banco com hash dos tokens
4. Armazena IP e User-Agent
5. Verifica limite de sessões (máx. 3)
6. Revoga sessões antigas se necessário

### 2. Refresh Token ✅ FUNCIONANDO
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJ..."
}
```

**Resposta:**
```json
{
  "access_token": "eyJ...",
  "refresh_token": "eyJ...",
  "expires_in": 900
}
```

**O que acontece:**
1. Valida refresh token no banco
2. Verifica se não foi revogado
3. Revoga sessão antiga
4. Cria nova sessão com novos tokens
5. Retorna novo par de tokens

### 3. Requisições Autenticadas
```http
GET /api/v1/profile
Authorization: Bearer eyJ...
```

**O que acontece:**
1. Extrai token do header Authorization
2. Valida assinatura JWT
3. Verifica se sessão existe no banco
4. Verifica se sessão não foi revogada
5. Verifica se token não expirou
6. Retorna dados do usuário

### 4. Logout
```http
POST /api/v1/security/logout
Authorization: Bearer eyJ...
```

**O que acontece:**
1. Identifica usuário pelo token
2. Revoga TODAS as sessões do usuário
3. Tokens existentes param de funcionar imediatamente

## 🔒 Recursos de Segurança

### Gestão de Sessões
- **Limite de Sessões**: Máximo 3 sessões simultâneas por usuário
- **Auto-revogação**: Sessões mais antigas são revogadas automaticamente
- **Rastreamento**: IP e User-Agent registrados para cada sessão

### Validação de Tokens
- **Dois Níveis**: Validação JWT + Validação de sessão no banco
- **Revogação Instantânea**: Tokens podem ser invalidados imediatamente
- **Expiração**: Access Token (15min), Refresh Token (24h)

### Auditoria
- Todas as operações de autenticação são registradas
- Falhas de login são monitoradas
- Tentativas de refresh com tokens inválidos são logadas

## 📊 Estrutura da Sessão no Banco

```sql
CREATE TABLE session_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    access_token_hash VARCHAR(64) NOT NULL,
    refresh_token_hash VARCHAR(64) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    refresh_expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN DEFAULT false,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

## 🧪 Testando

### 1. Login
```powershell
$body = @{
    email = "admin@example.com"
    password = "admin123"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" `
    -Method Post -Body $body -ContentType "application/json"

$accessToken = $response.access_token
$refreshToken = $response.refresh_token
```

### 2. Testar Access Token
```powershell
$headers = @{
    Authorization = "Bearer $accessToken"
}

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/profile" `
    -Headers $headers
```

### 3. Refresh Token
```powershell
$body = @{
    refresh_token = $refreshToken
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/refresh" `
    -Method Post -Body $body -ContentType "application/json"

$newAccessToken = $response.access_token
```

### 4. Logout (Revoga todas as sessões)
```powershell
$headers = @{
    Authorization = "Bearer $accessToken"
}

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/security/logout" `
    -Method Post -Headers $headers
```

## ✅ Benefícios da Refatoração

1. **Segurança Aprimorada**: Controle total sobre sessões ativas
2. **Revogação em Tempo Real**: Logout funciona instantaneamente
3. **Auditoria Completa**: Rastreamento de todas as ações de autenticação
4. **Manutenibilidade**: Sistema unificado, mais fácil de entender e manter
5. **Escalabilidade**: Preparado para features futuras (2FA, MFA, etc.)

## 🗑️ Código Removido

- ❌ `JWTManager` (internal/auth/jwt.go) - Não é mais usado
- ❌ `JWTManagerInterface` - Não é mais necessário
- ❌ Todos os imports de `internal/auth` nos handlers e rotas
- ❌ Geração de tokens duplicada

## 🚀 Próximos Passos Recomendados

1. ✅ Remover arquivo `internal/auth/jwt.go` completamente (se não usado em testes)
2. ✅ Atualizar testes de integração para usar tokenService
3. ✅ Adicionar cleanup job para sessões expiradas
4. ✅ Implementar notificações de novas sessões
5. ✅ Adicionar dashboard de sessões ativas no frontend

---

**Data da Refatoração**: 13 de outubro de 2025
**Sistema Anterior**: JWTManager (stateless)
**Sistema Atual**: TokenService (stateful com banco de dados)
**Status**: ✅ Funcionando e testado
