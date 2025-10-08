# DashTrack - Documentação Completa do Sistema

## 📋 Índice
1. [Visão Geral do Sistema](#visão-geral-do-sistema)
2. [Arquitetura Multi-tenant](#arquitetura-multi-tenant)
3. [Hierarquia de Usuários](#hierarquia-de-usuários)
4. [Funcionalidades por Perfil](#funcionalidades-por-perfil)
5. [Implementação Técnica](#implementação-técnica)
6. [API Endpoints](#api-endpoints)
7. [Testes Manuais com Postman](#testes-manuais-com-postman)

---

## 🎯 Visão Geral do Sistema

O **DashTrack** é um sistema de gestão de frotas e rastreamento veicular desenvolvido em Go, utilizando arquitetura REST API com autenticação JWT. O sistema foi projetado para suportar múltiplas empresas (multi-tenant) com diferentes níveis de acesso hierárquico.

### Principais Características:
- **Multi-tenant**: Suporte a múltiplas empresas isoladas
- **Autenticação JWT**: Sistema seguro de tokens de acesso
- **RBAC**: Controle de acesso baseado em funções (Role-Based Access Control)
- **Audit Logs**: Registro completo de ações do sistema
- **Rate Limiting**: Proteção contra abuso de API
- **2FA**: Autenticação de dois fatores (implementado)
- **Session Management**: Gerenciamento avançado de sessões

---

## 🏢 Arquitetura Multi-tenant

### Conceito de Empresas (Companies)
Cada empresa é uma entidade isolada com seus próprios:
- Usuários
- Veículos
- Equipes (Teams)
- Sensores
- Dados de rastreamento

### Estrutura da Empresa:
```json
{
  "id": "uuid",
  "name": "Nome da Empresa",
  "slug": "nome-empresa",
  "email": "contato@empresa.com",
  "phone": "+5511999999999",
  "address": "Endereço completo",
  "settings": {
    "timezone": "America/Sao_Paulo",
    "language": "pt-BR"
  }
}
```

---

## 👥 Hierarquia de Usuários

### 1. **Master** (Superusuário do Sistema)
- **Acesso**: Total ao sistema, todas as empresas
- **Responsabilidades**:
  - Criar e gerenciar empresas
  - Criar usuários company_admin
  - Monitorar todo o sistema
  - Configurações globais
  - Auditoria completa

### 2. **Company Admin** (Administrador da Empresa)
- **Acesso**: Limitado à sua empresa
- **Responsabilidades**:
  - Gerenciar usuários da empresa
  - Configurar veículos e equipes
  - Visualizar relatórios da empresa
  - Gerenciar sensores e dispositivos
  - Configurações da empresa

### 3. **Admin** (Administrador Geral)
- **Acesso**: Amplo, mas limitado por empresa
- **Responsabilidades**:
  - Gerenciar usuários
  - Configurar sistema
  - Visualizar relatórios
  - Suporte técnico

### 4. **Driver** (Motorista)
- **Acesso**: Limitado aos veículos atribuídos
- **Responsabilidades**:
  - Visualizar informações do veículo
  - Atualizar status de viagem
  - Reportar problemas
  - Acessar histórico de viagens

### 5. **Helper** (Ajudante)
- **Acesso**: Muito limitado, apenas veículos específicos
- **Responsabilidades**:
  - Visualizar informações básicas
  - Reportar status de entregas
  - Acesso limitado ao histórico

---

## 🔧 Funcionalidades por Perfil

### Master
```
✅ CRUD Empresas (companies)
✅ Dashboard global do sistema
✅ Gestão completa de usuários
✅ Auditoria completa
✅ Configurações globais
✅ Monitoramento de sistema
```

### Company Admin
```
✅ Dashboard da empresa
✅ CRUD Usuários da empresa
✅ CRUD Veículos
✅ CRUD Equipes
✅ CRUD Sensores
✅ Relatórios da empresa
✅ Configurações da empresa
```

### Admin
```
✅ Gestão de usuários
✅ Visualização de relatórios
✅ Configurações do sistema
✅ Suporte técnico
```

### Driver/Helper
```
✅ Visualização de veículos atribuídos
✅ Atualização de status
✅ Histórico limitado
✅ Perfil pessoal
```

---

## 🏗️ Implementação Técnica

### Autenticação JWT
```go
// Token Structure
{
  "user_id": "uuid",
  "email": "user@example.com",
  "role": "driver",
  "company_id": "uuid",
  "exp": timestamp,
  "iat": timestamp
}
```

### Middleware de Autenticação
- **RequireAuth()**: Valida token JWT
- **RequireRole(role)**: Exige papel específico
- **RequireAnyRole(roles...)**: Aceita múltiplos papéis

### Controle Multi-tenant
```go
// Cada query é filtrada por company_id
WHERE company_id = $1 AND user_id = $2
```

### Segurança Implementada
1. **Rate Limiting**: Limite de requisições por IP/usuário
2. **2FA**: Autenticação de dois fatores
3. **Session Management**: Controle de sessões ativas
4. **Audit Logs**: Log de todas as ações
5. **Password Security**: Hash bcrypt + políticas

### Banco de Dados
```sql
-- Principais tabelas
- companies (empresas)
- users (usuários)
- roles (perfis)
- user_sessions (sessões)
- auth_logs (logs de autenticação)
- rate_limit_rules (regras de rate limiting)
- two_factor_auth (2FA)
- audit_logs (auditoria)
```

---

## 🌐 API Endpoints

### Públicos (Sem Autenticação)
```
GET    /health                    # Status da API
GET    /metrics                   # Métricas Prometheus
POST   /api/v1/auth/login         # Login
POST   /api/v1/auth/refresh       # Renovar token
POST   /api/v1/auth/forgot-password  # Esqueci senha
POST   /api/v1/auth/reset-password   # Resetar senha
```

### Autenticados (Token JWT Obrigatório)
```
GET    /api/v1/profile            # Perfil do usuário
POST   /api/v1/profile/change-password  # Alterar senha
GET    /api/v1/roles              # Listar perfis
POST   /api/v1/auth/logout        # Logout
```

### Master Only
```
POST   /api/v1/master/companies        # Criar empresa
GET    /api/v1/master/companies        # Listar empresas
DELETE /api/v1/master/companies/:id    # Deletar empresa
GET    /api/v1/master/dashboard        # Dashboard master
GET    /api/v1/master/users            # Todos os usuários
POST   /api/v1/master/users            # Criar usuário
```

### Admin/Company Admin
```
GET    /api/v1/admin/users         # Listar usuários
POST   /api/v1/admin/users         # Criar usuário
GET    /api/v1/admin/users/:id     # Obter usuário
PUT    /api/v1/admin/users/:id     # Atualizar usuário
DELETE /api/v1/admin/users/:id     # Deletar usuário
```

### Sistema (Admin+)
```
GET    /api/v1/system/users        # Usuários do sistema
GET    /api/v1/system/roles        # Perfis do sistema
GET    /api/v1/audit/logs          # Logs de auditoria
```

### Manager
```
GET    /api/v1/manager/users       # Usuários gerenciados
```

### Segurança (2FA)
```
POST   /api/v1/security/2fa/enable    # Habilitar 2FA
POST   /api/v1/security/2fa/disable   # Desabilitar 2FA
POST   /api/v1/security/2fa/verify    # Verificar código 2FA
POST   /api/v1/security/2fa/backup-codes  # Gerar códigos backup
```

### Sessões
```
GET    /api/v1/sessions/dashboard      # Dashboard de sessões
GET    /api/v1/sessions/active         # Sessões ativas
DELETE /api/v1/sessions/:sessionId     # Revogar sessão
GET    /api/v1/sessions/metrics        # Métricas de sessão
GET    /api/v1/sessions/security-alerts # Alertas de segurança
```

---

## 🧪 Testes Manuais com Postman

### Configuração Inicial
1. **Base URL**: `http://localhost:8080`
2. **Headers Globais**:
   ```
   Content-Type: application/json
   Accept: application/json
   ```

### 1. Teste de Health Check
```http
GET {{base_url}}/health
```
**Resposta Esperada (200)**:
```json
{
  "status": "ok",
  "message": "API is running with HOT RELOAD and GIN! 🚀",
  "database": "connected",
  "version": "1.0.0",
  "timestamp": "2025-10-08T16:27:38Z"
}
```

### 2. Criação de Usuário Master (Primeira execução)
**⚠️ Nota**: Como o banco começa limpo, você precisa criar a primeira empresa e usuário master:

```sql
-- Execute no PostgreSQL
-- 1. Criar primeira empresa
INSERT INTO companies (name, slug, email, subscription_plan, max_users) 
VALUES ('Master Company', 'master', 'master@dashtrack.com', 'enterprise', 1000);

-- 2. Criar usuário master
INSERT INTO users (name, email, password, role_id, company_id) 
SELECT 
    'Master Admin',
    'master@dashtrack.com',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: "password"
    r.id,
    c.id
FROM roles r, companies c 
WHERE r.name = 'master' AND c.slug = 'master';
```

### 3. Teste de Login
```http
POST {{base_url}}/api/v1/auth/login
Content-Type: application/json

{
  "email": "master@dashtrack.com",
  "password": "password"
}
```

**Resposta Esperada (200)**:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "user": {
    "id": "uuid",
    "name": "Master Admin",
    "email": "master@dashtrack.com",
    "role": {
      "name": "master",
      "description": "Master user with full system access"
    }
  }
}
```

### 4. Teste de Perfil (Com Token)
```http
GET {{base_url}}/api/v1/profile
Authorization: Bearer {{access_token}}
```

**Resposta Esperada (200)**:
```json
{
  "id": "uuid",
  "name": "Master Admin",
  "email": "master@dashtrack.com",
  "phone": null,
  "cpf": null,
  "avatar": null,
  "role": {
    "id": "uuid",
    "name": "master",
    "description": "Master user with full system access"
  },
  "company_id": "uuid",
  "active": true,
  "last_login": "2025-10-08T16:30:00Z"
}
```

### 5. Criação de Nova Empresa (Master Only)
```http
POST {{base_url}}/api/v1/master/companies
Authorization: Bearer {{access_token}}
Content-Type: application/json

{
  "name": "Transportadora ABC",
  "slug": "transportadora-abc",
  "email": "contato@transportadoraabc.com",
  "phone": "+5511999999999",
  "address": "Rua das Empresas, 123, São Paulo, SP"
}
```

**Resposta Esperada (201)**:
```json
{
  "id": "uuid",
  "name": "Transportadora ABC",
  "slug": "transportadora-abc",
  "email": "contato@transportadoraabc.com",
  "phone": "+5511999999999",
  "address": "Rua das Empresas, 123, São Paulo, SP",
  "active": true,
  "created_at": "2025-10-08T16:30:00Z"
}
```

### 6. Criação de Company Admin
```http
POST {{base_url}}/api/v1/master/users
Authorization: Bearer {{access_token}}
Content-Type: application/json

{
  "name": "João Silva",
  "email": "joao@transportadoraabc.com",
  "password": "senhaSegura123",
  "phone": "+5511888888888",
  "cpf": "12345678901",
  "role_name": "company_admin",
  "company_id": "{{company_id_from_previous_request}}"
}
```

**Resposta Esperada (201)**:
```json
{
  "id": "uuid",
  "name": "João Silva",
  "email": "joao@transportadoraabc.com",
  "phone": "+5511888888888",
  "cpf": "12345678901",
  "role": {
    "name": "company_admin",
    "description": "Company administrator with company-wide access"
  },
  "company_id": "uuid",
  "active": true,
  "created_at": "2025-10-08T16:30:00Z"
}
```

### 7. Login como Company Admin
```http
POST {{base_url}}/api/v1/auth/login
Content-Type: application/json

{
  "email": "joao@transportadoraabc.com",
  "password": "senhaSegura123"
}
```

### 8. Teste de Autorização - Tentativa de Acesso Negado
```http
GET {{base_url}}/api/v1/master/companies
Authorization: Bearer {{company_admin_token}}
```

**Resposta Esperada (403)**:
```json
{
  "error": "Insufficient permissions"
}
```

### 9. Criação de Motorista (Company Admin)
```http
POST {{base_url}}/api/v1/admin/users
Authorization: Bearer {{company_admin_token}}
Content-Type: application/json

{
  "name": "Carlos Motorista",
  "email": "carlos@transportadoraabc.com",
  "password": "senha123",
  "phone": "+5511777777777",
  "role_name": "driver"
}
```

### 10. Teste de Refresh Token
```http
POST {{base_url}}/api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "{{refresh_token}}"
}
```

### 11. Teste de Logout
```http
POST {{base_url}}/api/v1/auth/logout
Authorization: Bearer {{access_token}}
```

**Resposta Esperada (200)**:
```json
{
  "message": "Successfully logged out"
}
```

### 12. Teste de Token Inválido (Após Logout)
```http
GET {{base_url}}/api/v1/profile
Authorization: Bearer {{expired_token}}
```

**Resposta Esperada (401)**:
```json
{
  "error": "Invalid or expired token"
}
```

---

## 🔍 Casos de Teste Específicos

### Teste de Rate Limiting
Faça múltiplas requisições rapidamente para o endpoint de login:
```http
POST {{base_url}}/api/v1/auth/login
(repetir várias vezes em sequência)
```

**Resposta Esperada (429)**:
```json
{
  "error": "Rate limit exceeded",
  "retry_after": 300
}
```

### Teste de Validação de Email
```http
POST {{base_url}}/api/v1/auth/login
Content-Type: application/json

{
  "email": "email-invalido",
  "password": "senha123"
}
```

**Resposta Esperada (400)**:
```json
{
  "error": "Invalid request format"
}
```

### Teste de Senha Incorreta
```http
POST {{base_url}}/api/v1/auth/login
Content-Type: application/json

{
  "email": "joao@transportadoraabc.com",
  "password": "senhaErrada"
}
```

**Resposta Esperada (401)**:
```json
{
  "error": "Invalid credentials"
}
```

---

## 📊 Collection Postman Sugerida

Crie uma collection no Postman com os seguintes folders:

1. **01 - Health & Setup**
   - Health Check
   - Database Check

2. **02 - Authentication**
   - Login Master
   - Login Company Admin
   - Login Driver
   - Refresh Token
   - Logout

3. **03 - Master Operations**
   - Create Company
   - List Companies
   - Create Master User
   - Master Dashboard

4. **04 - Company Admin Operations**
   - Create Users
   - List Company Users
   - Update User
   - Delete User

5. **05 - General User Operations**
   - Get Profile
   - Update Profile
   - Change Password
   - List Roles

6. **06 - Security Tests**
   - Invalid Token
   - Expired Token
   - Insufficient Permissions
   - Rate Limiting

7. **07 - Error Cases**
   - Invalid Email Format
   - Wrong Password
   - Missing Fields
   - Invalid JSON

---

## 🔐 Variáveis de Ambiente Postman

Configure estas variáveis na sua collection:
```
base_url: http://localhost:8080
master_token: {{obtido_no_login_master}}
company_admin_token: {{obtido_no_login_company_admin}}
driver_token: {{obtido_no_login_driver}}
company_id: {{id_da_empresa_criada}}
```

---

## ✅ Checklist de Testes

### Funcionalidades Básicas
- [ ] Health check responde
- [ ] Login master funciona
- [ ] Criação de empresa funciona
- [ ] Criação de usuário company_admin funciona
- [ ] Login company_admin funciona
- [ ] Criação de motorista funciona
- [ ] Refresh token funciona
- [ ] Logout funciona

### Segurança
- [ ] Token inválido é rejeitado
- [ ] Acesso sem permissão é negado
- [ ] Rate limiting funciona
- [ ] Validações de input funcionam

### Multi-tenant
- [ ] Company admin só vê sua empresa
- [ ] Motorista só vê seus dados
- [ ] Master vê todas as empresas

Este guia completo deve cobrir todos os aspectos do sistema DashTrack e permitir testes abrangentes de todas as funcionalidades implementadas até o momento! 🚀