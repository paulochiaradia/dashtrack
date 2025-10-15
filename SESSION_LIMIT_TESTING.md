# Script de Teste - Limite de Sessões

## Setup
```bash
API_URL="http://localhost:8080/api/v1"
EMAIL="test@example.com"
PASSWORD="senha123"
```

## Teste 1: Revogação Automática de Sessões

### Passo 1: Fazer 3 logins (limite de sessões)
```bash
# Login 1
TOKEN1=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" | jq -r '.access_token')

echo "Token 1: $TOKEN1"

# Login 2
TOKEN2=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" | jq -r '.access_token')

echo "Token 2: $TOKEN2"

# Login 3
TOKEN3=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" | jq -r '.access_token')

echo "Token 3: $TOKEN3"
```

### Passo 2: Verificar sessões ativas (deve ter 3)
```bash
curl -s -X GET "$API_URL/sessions/active" \
  -H "Authorization: Bearer $TOKEN3" | jq
```

**Resultado Esperado**:
```json
{
  "sessions": [
    {...},
    {...},
    {...}
  ],
  "total": 3
}
```

### Passo 3: Fazer 4º login (deve revogar Token1 automaticamente)
```bash
TOKEN4=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" | jq -r '.access_token')

echo "Token 4: $TOKEN4"
```

### Passo 4: Verificar email recebido
Verifique o email configurado no sistema com:
- ✉️ **Assunto**: 🔒 Nova sessão ativada - Sessões antigas revogadas
- 📋 **Conteúdo**:
  - Detalhes da nova sessão (IP, user-agent, data/hora)
  - Informação sobre 1 sessão revogada
  - Instruções de segurança
  - Botão "Ver Sessões Ativas"

### Passo 5: Verificar sessões ativas (deve ter 3, Token1 revogado)
```bash
curl -s -X GET "$API_URL/sessions/active" \
  -H "Authorization: Bearer $TOKEN4" | jq
```

**Resultado Esperado**: 3 sessões (Token2, Token3, Token4)

### Passo 6: Tentar usar Token1 (deve falhar - revogado)
```bash
curl -s -X GET "$API_URL/sessions/active" \
  -H "Authorization: Bearer $TOKEN1" | jq
```

**Resultado Esperado**:
```json
{
  "error": "Token inválido ou revogado"
}
```

---

## Teste 2: Revogação Manual de Todas Sessões Exceto Atual

### Passo 1: Setup - Criar 3 sessões
```bash
# Login e obter 3 tokens diferentes
TOKEN1=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" | jq -r '.access_token')

TOKEN2=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" | jq -r '.access_token')

TOKEN3=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" | jq -r '.access_token')

echo "Criadas 3 sessões"
```

### Passo 2: Verificar sessões ativas com TOKEN3
```bash
curl -s -X GET "$API_URL/sessions/active" \
  -H "Authorization: Bearer $TOKEN3" | jq
```

**Resultado Esperado**: 3 sessões ativas

### Passo 3: Revogar todas exceto TOKEN3 (atual)
```bash
curl -s -X DELETE "$API_URL/sessions/revoke-all-except-current" \
  -H "Authorization: Bearer $TOKEN3" | jq
```

**Resultado Esperado**:
```json
{
  "message": "Todas as outras sessões foram revogadas com sucesso",
  "revoked_count": 2
}
```

### Passo 4: Verificar que apenas TOKEN3 está ativo
```bash
# TOKEN3 deve funcionar
curl -s -X GET "$API_URL/sessions/active" \
  -H "Authorization: Bearer $TOKEN3" | jq

# TOKEN1 deve falhar
curl -s -X GET "$API_URL/sessions/active" \
  -H "Authorization: Bearer $TOKEN1" | jq

# TOKEN2 deve falhar
curl -s -X GET "$API_URL/sessions/active" \
  -H "Authorization: Bearer $TOKEN2" | jq
```

**Resultado Esperado**:
- TOKEN3: Retorna 1 sessão ativa
- TOKEN1: Erro "Token inválido ou revogado"
- TOKEN2: Erro "Token inválido ou revogado"

---

## Teste 3: Fluxo Completo de Segurança

### Cenário: Usuário suspeita de acesso não autorizado

```bash
# 1. Usuário faz login em dispositivo confiável
TRUSTED_TOKEN=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" | jq -r '.access_token')

# 2. Atacante faz login em dispositivo suspeito (simular)
ATTACKER_TOKEN=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" | jq -r '.access_token')

# 3. Usuário recebe email de nova sessão (revisar)

# 4. Usuário verifica sessões ativas
curl -s -X GET "$API_URL/sessions/active" \
  -H "Authorization: Bearer $TRUSTED_TOKEN" | jq

# 5. Usuário identifica sessão suspeita e revoga todas exceto atual
curl -s -X DELETE "$API_URL/sessions/revoke-all-except-current" \
  -H "Authorization: Bearer $TRUSTED_TOKEN" | jq

# 6. Usuário muda a senha (endpoint existente)
curl -s -X PUT "$API_URL/auth/change-password" \
  -H "Authorization: Bearer $TRUSTED_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"old_password":"senha123","new_password":"novaSenha456"}' | jq

# 7. Atacante não consegue mais acessar
curl -s -X GET "$API_URL/sessions/active" \
  -H "Authorization: Bearer $ATTACKER_TOKEN" | jq
```

**Resultado Esperado**: Acesso do atacante bloqueado, apenas usuário legítimo tem acesso

---

## Validações de Logs

### Logs esperados no servidor:

```log
# Revogação automática
INFO Revoked old sessions due to limit
  user_id: <UUID>
  revoked_count: 1

# Email enviado
INFO Email de limite de sessões enviado com sucesso
  user_id: <UUID>
  email: test@example.com

# Revogação manual
INFO Usuário revogou todas as outras sessões
  user_id: <UUID>
  session_id: <UUID>
  revoked_count: 2
```

### Verificar logs do container:
```bash
docker logs dashtrack-api-1 --tail 50 | grep -i "session"
```

---

## Checklist de Validação

- [ ] Limite de 3 sessões ativas é respeitado
- [ ] Sessão mais antiga é revogada automaticamente no 4º login
- [ ] Email de notificação é enviado com informações corretas
- [ ] Endpoint de revogação manual funciona corretamente
- [ ] Sessão atual nunca é revogada na operação manual
- [ ] Tokens revogados não conseguem mais acessar endpoints protegidos
- [ ] Logs são registrados corretamente
- [ ] Contadores de sessões revogadas estão corretos
- [ ] Mensagens de erro são apropriadas para tokens inválidos

---

## Troubleshooting

### Email não chegou?
1. Verificar configuração SMTP no `.env`:
   ```
   SMTP_HOST=smtp.umbler.com
   SMTP_PORT=587
   SMTP_FROM=seu-email@seudominio.com
   SMTP_PASSWORD=<senha>
   ```
2. Verificar logs: `docker logs dashtrack-api-1 | grep -i email`
3. Verificar caixa de spam

### Sessões não sendo revogadas?
1. Verificar banco de dados:
   ```sql
   SELECT * FROM session_tokens WHERE user_id = '<UUID>' ORDER BY created_at DESC;
   ```
2. Verificar campo `revoked` (deve ser `true` para revogadas)
3. Verificar logs de erro no container

### Endpoint retorna 401?
1. Verificar formato do token: `Bearer <token>`
2. Verificar se token não expirou
3. Verificar se token não foi revogado
