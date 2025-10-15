# Security Guidelines - DashTrack

## 🔐 Credenciais e Secrets

### ❌ NUNCA commitar:
- Arquivos `.env` com credenciais reais
- Senhas em texto plano
- Tokens de API
- Chaves privadas
- Certificados SSL
- Credenciais de banco de dados

### ✅ SEMPRE usar:
- Variáveis de ambiente (`.env`)
- `.env.example` com placeholders
- Secrets managers em produção (AWS Secrets, Azure Key Vault)
- `.gitignore` para proteger arquivos sensíveis

## 📧 SMTP Configuration

### Arquivo `.env` (LOCAL - NÃO COMMITADO)
```bash
SMTP_HOST=smtp.umbler.com
SMTP_PORT=587
SMTP_USERNAME=seu-email@seudominio.com
SMTP_PASSWORD=sua-senha-real-aqui
SMTP_FROM=seu-email@seudominio.com
SMTP_FROM_NAME=DashTrack
SMTP_USE_TLS=true
```

### Em caso de vazamento:

1. **IMEDIATO**: Trocar senha do email
2. **URGENTE**: Revogar tokens/keys expostos
3. **IMPORTANTE**: Verificar logs de acesso não autorizado
4. **RECOMENDADO**: Usar senhas únicas por serviço

## 🛡️ Proteções Implementadas

### Git
- `.env` no `.gitignore`
- `.env.example` com placeholders
- Nenhuma credencial em commits

### Aplicação
- Credenciais carregadas de variáveis de ambiente
- Nunca logadas em plain text
- Conexões TLS/SSL obrigatórias

## 🔍 Como verificar vazamentos

### Verificar histórico Git:
```bash
# Verificar se .env foi commitado
git log --all --full-history -- .env

# Buscar por padrões de senha
git log -p | grep -i "password"

# Ver arquivos em commit específico
git show <commit-hash> --name-only
```

### Ferramentas recomendadas:
- [GitGuardian](https://www.gitguardian.com/) - Scan automático
- [git-secrets](https://github.com/awslabs/git-secrets) - Pre-commit hooks
- [truffleHog](https://github.com/trufflesecurity/trufflehog) - Scan de histórico

## 📝 Checklist antes de commit

- [ ] `.env` não está staged (`git status`)
- [ ] Nenhuma senha em plain text no código
- [ ] Apenas `.env.example` com placeholders
- [ ] Logs não contêm informações sensíveis
- [ ] Tokens removidos de comentários

## 🚨 Resposta a Incidentes

### Se credenciais foram expostas:

1. **Trocar credenciais imediatamente**
2. **Remover do histórico Git** (se possível)
   ```bash
   # BFG Repo-Cleaner (recomendado)
   bfg --delete-files .env
   
   # Ou git filter-branch (mais complexo)
   git filter-branch --force --index-filter \
     "git rm --cached --ignore-unmatch .env" \
     --prune-empty --tag-name-filter cat -- --all
   ```
3. **Force push** (atenção: reescreve histórico)
   ```bash
   git push --force --all
   ```
4. **Notificar time** sobre comprometimento

## 📚 Referências

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [GitHub Security Best Practices](https://docs.github.com/en/code-security)
- [12 Factor App - Config](https://12factor.net/config)

---

**Última atualização:** Outubro 2025  
**Responsável:** Time DashTrack
