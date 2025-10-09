# 🚀 Guia de Testes da API DashTrack - Atualizado

## 📝 Mudanças Implementadas

### ✅ Padronização de Rotas
Todas as rotas agora usam o prefixo `/api/v1/` para melhor versionamento da API.

### ✅ Sistema de Sessões Avançado
- Limite de 3 sessões simultâneas por usuário
- Dashboard de sessões com métricas detalhadas
- Detecção de atividade suspeita
- Controle total de sessões ativas

---

## 🔐 **1. AUTENTICAÇÃO**

### 1.1 Login
```http
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json

{
  "email": "master@dashtrack.com",
  "password": "Master123!"
}
```

**Resposta de Sucesso:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900,
  "user": {
    "id": "9a10df3b-9ee1-4ce8-bc09-20a586b56aa5",
    "name": "Master Admin",
    "email": "master@dashtrack.com",
    "role": "master"
  }
}
```

### 1.2 Refresh Token
```http
POST http://localhost:8080/api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "SEU_REFRESH_TOKEN_AQUI"
}
```

### 1.3 Logout
```http
POST http://localhost:8080/api/v1/security/logout
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
```

---

## 👥 **2. GESTÃO DE SESSÕES** (Novo!)

### 2.1 Dashboard de Sessões
```http
GET http://localhost:8080/api/v1/sessions/dashboard
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
```

**Resposta inclui:**
- Métricas completas (total de sessões, sessões ativas, tempo gasto)
- Lista de sessões ativas com IP, User-Agent e duração
- Histórico de sessões recentes
- Alertas de segurança
- Avisos sobre limite de sessões

### 2.2 Sessões Ativas
```http
GET http://localhost:8080/api/v1/sessions/active
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
```

### 2.3 Métricas de Sessão
```http
GET http://localhost:8080/api/v1/sessions/metrics
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
```

### 2.4 Alertas de Segurança
```http
GET http://localhost:8080/api/v1/sessions/security-alerts
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
```

### 2.5 Revogar Sessão Específica
```http
DELETE http://localhost:8080/api/v1/sessions/ID_DA_SESSAO
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
```

---

## 👤 **3. PERFIL DO USUÁRIO**

### 3.1 Obter Perfil
```http
GET http://localhost:8080/api/v1/profile
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
```

### 3.2 Alterar Senha
```http
POST http://localhost:8080/api/v1/profile/change-password
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
Content-Type: application/json

{
  "current_password": "Master123!",
  "new_password": "NovaSenh@123!"
}
```

### 3.3 Listar Roles
```http
GET http://localhost:8080/api/v1/roles
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
```

---

## 🔒 **4. SEGURANÇA AVANÇADA**

### 4.1 Status 2FA
```http
GET http://localhost:8080/api/v1/security/2fa/status
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
```

### 4.2 Configurar 2FA
```http
POST http://localhost:8080/api/v1/security/2fa/setup
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
```

### 4.3 Logs de Auditoria (Admin/Master apenas)
```http
GET http://localhost:8080/api/v1/security/audit/logs
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
```

---

## 👨‍💼 **5. ADMINISTRAÇÃO**

### 5.1 Listar Usuários (Admin/Master)
```http
GET http://localhost:8080/api/v1/admin/users
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
```

### 5.2 Criar Usuário (Admin/Master)
```http
POST http://localhost:8080/api/v1/admin/users
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
Content-Type: application/json

{
  "name": "Novo Usuário",
  "email": "usuario@teste.com",
  "password": "MinhaSenh@123!",
  "phone": "(11) 99999-9999",
  "cpf": "123.456.789-00",
  "role_id": "ID_DA_ROLE"
}
```

### 5.3 Criar Usuário Master (Master apenas)
```http
POST http://localhost:8080/api/v1/master/users
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
Content-Type: application/json

{
  "name": "General Admin",
  "email": "generaladmin@system.com",
  "password": "admin123",
  "phone": "+5511999888777",
  "cpf": "987.654.321-00",
  "role_id": "6a258ff7-884e-4083-989e-49952b1a2095"
}
```

**⚠️ Campos Obrigatórios (Migração 8):**
- `phone`: Telefone no formato brasileiro (mín: 10, máx: 20 caracteres)
- `cpf`: CPF no formato XXX.XXX.XXX-XX (exatamente 14 caracteres)

---

## 🏢 **6. MULTI-TENANT (Empresas)**

### 6.1 Informações da Empresa
```http
GET http://localhost:8080/api/v1/company/info
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
```

### 6.2 Dashboard da Empresa
```http
GET http://localhost:8080/api/v1/company/dashboard
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
```

### 6.3 Times da Empresa
```http
GET http://localhost:8080/api/v1/company/teams
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
```

### 6.4 Veículos da Empresa
```http
GET http://localhost:8080/api/v1/company/vehicles
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
```

### 6.5 Dispositivos ESP32
```http
GET http://localhost:8080/api/v1/company/devices
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
```

---

## 🌐 **7. IOT E SENSORES**

### 7.1 Registrar Sensor
```http
POST http://localhost:8080/api/v1/sensors/register
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
Content-Type: application/json

{
  "device_id": "ESP32_001",
  "sensor_type": "GPS",
  "location": "Veículo Principal"
}
```

### 7.2 Meus Sensores
```http
GET http://localhost:8080/api/v1/sensors/my
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
```

### 7.3 Dados do Sensor
```http
GET http://localhost:8080/api/v1/sensors/ESP32_001/data
Authorization: Bearer SEU_ACCESS_TOKEN_AQUI
```

### 7.4 Health Check IoT
```http
GET http://localhost:8080/api/v1/iot/health
```

### 7.5 Registrar Dispositivo ESP32
```http
POST http://localhost:8080/api/v1/esp32/register
Content-Type: application/json

{
  "device_id": "ESP32_001",
  "device_name": "Sensor Principal",
  "firmware_version": "1.0.0"
}
```

---

## 🛠️ **8. MONITORAMENTO**

### 8.1 Health Check
```http
GET http://localhost:8080/health
```

### 8.2 Métricas Prometheus
```http
GET http://localhost:8080/metrics
```

---

## 🚨 **Recursos de Segurança Implementados**

### ✅ Controle de Sessões
- **Limite de 3 sessões simultâneas por usuário**
- **Revogação automática de sessões antigas**
- **Tracking de IP, User-Agent e duração**

### ✅ Monitoramento de Segurança
- **Detecção de múltiplos IPs simultâneos**
- **Alertas para muitos dispositivos ativos**
- **Histórico completo de sessões**

### ✅ JWT Avançado
- **Access Token: 15 minutos**
- **Refresh Token: 24 horas**
- **Rotação automática de tokens**

### ✅ Rate Limiting
- **Proteção contra ataques de força bruta**
- **Configurável por endpoint**

### ✅ Auditoria Completa
- **Log de todas as ações importantes**
- **Rastreamento de mudanças**
- **Alertas de segurança**

---

## 📋 **Notas Importantes**

1. **Todas as rotas agora usam o prefixo `/api/v1/`**
2. **O sistema de sessões tem limite de 3 dispositivos simultâneos**
3. **Logout revoga sessões no banco, mas JWT ainda é válido até expirar**
4. **Para JWT realmente invalidado, use refresh token ou espere expiração**
5. **Dashboard de sessões fornece visibilidade completa da atividade do usuário**
6. **⚠️ MIGRAÇÃO 8 - Campos obrigatórios para criação de usuários:**
   - `phone`: Telefone brasileiro (10-20 caracteres)
   - `cpf`: CPF no formato XXX.XXX.XXX-XX (exatamente 14 caracteres)

---

## 🚨 **Correções Necessárias nos Testes do Postman**

### **⚠️ Teste 07B - Create a global Admin**
**Problema**: Após migração 8, campos `phone` e `cpf` são obrigatórios.

**Body corrigido**:
```json
{
    "name": "General Admin",
    "email": "generaladmin@system.com",
    "password": "admin123",
    "phone": "+5511999888777",
    "cpf": "987.654.321-00",
    "role_id": "6a258ff7-884e-4083-989e-49952b1a2095"
}
```

### **⚠️ Teste 07C - Create Company Admin User**
**Verificação**: Este teste já tem os campos `phone` e `cpf` corretos! ✅

**Body atual (já correto)**:
```json
{
    "name": "Company Admin",
    "email": "companyadmin@testcompany.com", 
    "password": "admin123",
    "phone": "+5511999999999",
    "cpf": "123.456.789-01",
    "role_id": "0d87651d-eade-48e8-9b72-9917711b9ec4",
    "company_id": "{{company_id}}"
}
```

### **📋 Resumo das Mudanças Necessárias**:
1. **Apenas o teste 07B** precisa ser atualizado
2. **Adicionar os campos**:
   - `"phone": "+5511999888777"`
   - `"cpf": "987.654.321-00"`
3. **Todos os outros testes** continuam funcionando normalmente

---

## 🚨 **Mudanças da Migração 8**

A **migração 008_make_phone_cpf_required** tornou obrigatórios os campos:

### Para criação de novos usuários:
```json
{
  "name": "Nome do Usuário",
  "email": "email@example.com", 
  "password": "SenhaSegur@123!",
  "phone": "(11) 99999-9999",    // ⚠️ OBRIGATÓRIO
  "cpf": "123.456.789-00",       // ⚠️ OBRIGATÓRIO  
  "role_id": "uuid-da-role"
}
```

### Formatos aceitos:
- **Phone**: "(11) 99999-9999", "11999999999", "+5511999999999"
- **CPF**: "123.456.789-00" (exatamente 14 caracteres com pontos e traço)

---

## 🔄 **Fluxo Recomendado de Teste**

1. **Login** → Obter access_token
2. **Dashboard de Sessões** → Ver atividade atual
3. **Teste de múltiplas sessões** → Login de diferentes "dispositivos"
4. **Verificar limites** → Sistema deve revogar sessões antigas
5. **Logout** → Limpar sessões ativas
6. **Testar outros endpoints** → Profile, admin, company, etc.

---

Este guia garante testes completos de toda a funcionalidade implementada! 🚀