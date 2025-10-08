# 🧪 Guia de Testes Manuais - DashTrack

## 🚀 Configuração Inicial

### 1. Subir a Aplicação
```bash
cd dashtrack
docker-compose up -d
```

### 2. Verificar Status
```bash
# Verificar se todos os containers estão rodando
docker-compose ps

# Verificar logs da API
docker-compose logs api --tail=10
```

### 3. Criar Usuário Master
Execute o script SQL no banco de dados:
```bash
docker-compose exec db psql -U user -d dashtrack -f /scripts/create_master_user.sql
```

**Ou execute manualmente:**
```bash
docker-compose exec db psql -U user -d dashtrack -c "
-- Inserir empresa master se não existir
INSERT INTO companies (name, slug, email) 
SELECT 'Master Company', 'master', 'master@dashtrack.com'
WHERE NOT EXISTS (SELECT 1 FROM companies WHERE slug = 'master');

-- Inserir usuário master se não existir
INSERT INTO users (name, email, password, role_id, company_id) 
SELECT 
    'Master Admin',
    'master@dashtrack.com',
    '\$2a\$10\$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    r.id,
    c.id
FROM roles r, companies c 
WHERE r.name = 'master' AND c.slug = 'master'
AND NOT EXISTS (SELECT 1 FROM users WHERE email = 'master@dashtrack.com');
"
```

## 📋 Sequência de Testes Recomendada

### Fase 1: Verificação Básica
1. **Health Check** ✅
2. **Login Master** ✅
3. **Perfil Master** ✅

### Fase 2: Criação de Empresa
4. **Criar Nova Empresa** ✅
5. **Listar Empresas** ✅

### Fase 3: Gestão de Usuários
6. **Criar Company Admin** ✅
7. **Login Company Admin** ✅
8. **Criar Motorista** ✅
9. **Listar Usuários da Empresa** ✅

### Fase 4: Testes de Segurança
10. **Teste de Autorização** ❌ (Esperado)
11. **Token Inválido** ❌ (Esperado)
12. **Rate Limiting** ❌ (Esperado)

### Fase 5: Funcionalidades Avançadas
13. **Refresh Token** ✅
14. **Logout** ✅
15. **Change Password** ✅

## 🔧 Configuração do Postman

### Variáveis de Ambiente
```json
{
  "base_url": "http://localhost:8080",
  "master_email": "master@dashtrack.com",
  "master_password": "password",
  "company_admin_email": "",
  "company_admin_password": "",
  "driver_email": "",
  "driver_password": "",
  "access_token": "",
  "refresh_token": "",
  "company_id": ""
}
```

### Headers Globais
```
Content-Type: application/json
Accept: application/json
```

## 📝 Scripts de Teste Automatizado

### Teste 1: Health Check
```javascript
// Test script
pm.test("Status code is 200", function () {
    pm.response.to.have.status(200);
});

pm.test("Response has correct structure", function () {
    var jsonData = pm.response.json();
    pm.expect(jsonData).to.have.property('status', 'ok');
    pm.expect(jsonData).to.have.property('database', 'connected');
});
```

### Teste 2: Login Master
```javascript
// Pre-request script
pm.globals.set("master_email", "master@dashtrack.com");
pm.globals.set("master_password", "password");

// Test script
pm.test("Login successful", function () {
    pm.response.to.have.status(200);
    var jsonData = pm.response.json();
    pm.expect(jsonData).to.have.property('access_token');
    
    // Salvar tokens
    pm.globals.set("access_token", jsonData.access_token);
    pm.globals.set("refresh_token", jsonData.refresh_token);
});
```

### Teste 3: Criar Empresa
```javascript
// Test script
pm.test("Company created", function () {
    pm.response.to.have.status(201);
    var jsonData = pm.response.json();
    pm.expect(jsonData).to.have.property('id');
    
    // Salvar company_id
    pm.globals.set("company_id", jsonData.id);
});
```

## 🐛 Problemas Conhecidos e Soluções

### Problema: "Token inválido"
**Solução**: Verificar se o token está sendo enviado corretamente no header:
```
Authorization: Bearer {{access_token}}
```

### Problema: "Rate limit exceeded"
**Solução**: Aguardar alguns minutos ou resetar o container:
```bash
docker-compose restart api
```

### Problema: "Database connection failed"
**Solução**: Verificar se o PostgreSQL está rodando:
```bash
docker-compose logs db
```

### Problema: Usuário master não encontrado
**Solução**: Executar novamente o script de criação do usuário master.

## 📊 Dados de Teste Sugeridos

### Empresa Exemplo
```json
{
  "name": "Transportadora São Paulo",
  "slug": "transportadora-sp",
  "email": "contato@transportadorasp.com",
  "phone": "+5511999887766",
  "address": "Av. Paulista, 1000, São Paulo, SP, 01310-100"
}
```

### Company Admin Exemplo
```json
{
  "name": "Maria Silva",
  "email": "maria@transportadorasp.com",
  "password": "senha123456",
  "phone": "+5511888777666",
  "cpf": "12345678901",
  "role_name": "company_admin"
}
```

### Motorista Exemplo
```json
{
  "name": "João Motorista",
  "email": "joao@transportadorasp.com",
  "password": "motorista123",
  "phone": "+5511777666555",
  "cpf": "98765432109",
  "role_name": "driver"
}
```

### Ajudante Exemplo
```json
{
  "name": "Pedro Ajudante",
  "email": "pedro@transportadorasp.com",
  "password": "ajudante123",
  "phone": "+5511666555444",
  "role_name": "helper"
}
```

## 🎯 Cenários de Teste Avançados

### Cenário 1: Fluxo Completo Multi-tenant
1. Master cria empresa A
2. Master cria empresa B
3. Master cria admin para empresa A
4. Master cria admin para empresa B
5. Admin A tenta acessar dados da empresa B (deve falhar)
6. Admin A cria usuários apenas na empresa A

### Cenário 2: Teste de Hierarquia
1. Company Admin cria motoristas
2. Motorista tenta criar usuários (deve falhar)
3. Motorista acessa apenas seus próprios dados
4. Helper tem acesso ainda mais limitado

### Cenário 3: Teste de Segurança
1. Múltiplos logins simultâneos
2. Logout e tentativa de usar token expirado
3. Alteração de senha
4. Rate limiting com múltiplas requisições

## 📈 Métricas de Sucesso

- ✅ **100% dos endpoints funcionais respondem corretamente**
- ✅ **Autenticação JWT funciona em todos os cenários**
- ✅ **Multi-tenancy isolando dados corretamente**
- ✅ **RBAC negando acesso adequadamente**
- ✅ **Rate limiting protegendo a API**
- ✅ **Logs de auditoria sendo gerados**

---

**🚀 Boa sorte com os testes! Se encontrar algum problema, verifique os logs do Docker e a documentação completa no arquivo SYSTEM_DOCUMENTATION.md**