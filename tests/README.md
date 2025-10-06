# Sistema de Testes - DashTrack

Estrutura organizada de testes para manter controle e qualidade do sistema.

## 📁 Estrutura de Pastas

```
tests/
├── unit/           # Testes unitários (handlers, services, repositories)
│   └── handlers/   # Testes de handlers específicos
├── integration/    # Testes de integração (componentes funcionando juntos)
├── e2e/           # Testes end-to-end (fluxos completos da aplicação)
├── benchmarks/    # Testes de performance e benchmarks
├── testutils/     # Utilitários compartilhados para testes
└── postman/       # Coleções Postman para testes manuais
```

## 🧪 Tipos de Teste

### 1. Testes Unitários (`unit/`)
- **Objetivo**: Testar componentes isoladamente
- **Escopo**: Handlers, Services, Repositories individuais
- **Exemplo**: `tests/unit/handlers/auth_test.go`

### 2. Testes de Integração (`integration/`)
- **Objetivo**: Testar integração entre componentes
- **Escopo**: Combinação de handlers + services + repositories
- **Exemplo**: Teste de fluxo de autenticação completo

### 3. Testes End-to-End (`e2e/`)
- **Objetivo**: Testar fluxos completos da aplicação
- **Escopo**: Requisições HTTP reais, respostas completas
- **Exemplo**: Teste de login → dashboard → CRUD de usuários

### 4. Benchmarks (`benchmarks/`)
- **Objetivo**: Medir performance e identificar gargalos
- **Escopo**: Endpoints críticos, operações de banco
- **Exemplo**: Performance de login, listagem de usuários

### 5. Utilitários de Teste (`testutils/`)
- **Objetivo**: Código compartilhado entre testes
- **Escopo**: Configuração de DB, mocks, helpers
- **Exemplo**: Setup de database para testes

## 🚀 Como Executar

```bash
# Todos os testes
go test ./tests/... -v

# Apenas testes unitários
go test ./tests/unit/... -v

# Apenas testes de integração
go test ./tests/integration/... -v

# Apenas testes e2e
go test ./tests/e2e/... -v

# Apenas benchmarks
go test ./tests/benchmarks/... -v -bench=.

# Com coverage
go test ./tests/... -v -coverprofile=coverage.out
```

## 📋 Boas Práticas

1. **Nomenclatura**: Usar `_test.go` como sufixo
2. **Packages**: Usar `*_test` como package name para evitar dependências circulares
3. **Setup/Teardown**: Implementar nos test suites para cleanup adequado
4. **Isolamento**: Cada teste deve ser independente
5. **Mocks**: Usar mocks para dependências externas
6. **Database**: Usar transações para rollback automático

## 🔧 Ferramentas Necessárias

- PostgreSQL (para testes de integração)
- Docker (opcional, para ambiente isolado)
- Testify (framework de testes)
- Mocks/Stubs para dependências externas

## 📊 Cobertura de Testes

Objetivo: Manter cobertura > 80% para:
- ✅ Handlers (lógica de requisição/resposta)
- ✅ Services (regras de negócio)
- ✅ Repositories (acesso a dados)
- ✅ Middleware (autenticação, autorização)