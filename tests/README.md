# DashTrack Testing Suite

Este documento descreve a suíte de testes abrangente do DashTrack, organizada para fornecer cobertura completa do sistema de gerenciamento de usuários multi-tenant.

## 📁 Estrutura de Testes

```
tests/
├── unit/                          # Testes unitários isolados
│   ├── services/                  # Testes da camada de serviços
│   │   └── user_service_test.go   # Testes do UserService
│   ├── middleware/                # Testes de middleware
│   │   └── auth_middleware_test.go # Testes de autenticação
│   └── repositories/              # Testes de repositórios
│       └── user_repository_test.go # Testes do UserRepository
├── integration/                   # Testes de integração
│   └── user_management_test.go    # Fluxos completos de usuário
├── e2e/                          # Testes end-to-end
│   └── user_workflows_test.go     # Cenários reais completos
├── benchmarks/                    # Testes de performance
│   └── performance_test.go        # Benchmarks de endpoints
├── testutils/                     # Utilitários compartilhados
│   ├── helpers.go                 # Helpers para testes
│   └── mocks/                     # Mocks gerados
└── postman/                       # Coleções do Postman
```

## 🧪 Tipos de Testes

### 1. Testes Unitários (`tests/unit/`)

Testam componentes individuais de forma isolada usando mocks.

**Cobertura:**
- ✅ UserService: CRUD completo, validação de permissões, hierarquia de roles
- ✅ AuthMiddleware: Validação JWT, controle de acesso baseado em roles
- ✅ UserRepository: Operações de banco de dados com SQL mocks

**Características:**
- Isolamento completo com mocks
- Testes rápidos e determinísticos
- Cobertura de casos de edge e cenários de erro

### 2. Testes de Integração (`tests/integration/`)

Testam a interação entre componentes reais com banco de dados.

**Cobertura:**
- ✅ Fluxo completo de autenticação
- ✅ CRUD de usuários com permissões reais
- ✅ Isolamento de dados por empresa
- ✅ Validação de hierarquia de roles

**Características:**
- Usa banco de dados real (PostgreSQL)
- Testa integrações entre camadas
- Valida configuração e migrations

### 3. Testes End-to-End (`tests/e2e/`)

Simulam cenários reais completos de usuário.

**Cenários Testados:**
- ✅ Workflow completo de gerenciamento de usuários
- ✅ Enforcement da hierarquia de permissões
- ✅ Isolamento de dados entre empresas
- ✅ Tratamento de tokens inválidos
- ✅ Operações concorrentes

**Características:**
- Servidor HTTP real em teste
- Cenários de múltiplos usuários
- Validação de segurança e isolamento

### 4. Testes de Performance (`tests/benchmarks/`)

Avaliam performance e identificam gargalos.

**Métricas:**
- ✅ Geração e validação de tokens JWT
- ✅ Performance de endpoints CRUD
- ✅ Overhead do middleware de autenticação
- ✅ Operações concorrentes
- ✅ Alocação de memória

**Características:**
- Benchmarks detalhados com métricas
- Testes de carga e concorrência
- Análise de alocação de memória

## 🚀 Como Executar os Testes

### Pré-requisitos

1. **PostgreSQL em execução:**
```bash
# Docker (recomendado)
docker run --name dashtrack-test-db -e POSTGRES_PASSWORD=password -e POSTGRES_DB=dashtrack_test -p 5432:5432 -d postgres:13

# Ou instale PostgreSQL localmente
```

2. **Variáveis de ambiente:**
```bash
export TEST_DATABASE_URL="postgres://postgres:password@localhost:5432/dashtrack_test?sslmode=disable"
export E2E_DATABASE_URL="postgres://postgres:password@localhost:5432/dashtrack_e2e?sslmode=disable"
```

### Execução dos Testes

#### 1. Todos os Testes
```bash
# Executa toda a suíte de testes
go test ./tests/... -v

# Com coverage
go test ./tests/... -v -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

#### 2. Testes Unitários
```bash
# Testes unitários (rápidos, sem banco)
go test ./tests/unit/... -v

# Testes específicos
go test ./tests/unit/services/ -v
go test ./tests/unit/middleware/ -v
go test ./tests/unit/repositories/ -v
```

#### 3. Testes de Integração
```bash
# Testes de integração (requer banco)
go test ./tests/integration/... -v

# Com timeout maior para operações de banco
go test ./tests/integration/... -v -timeout=30s
```

#### 4. Testes End-to-End
```bash
# Testes E2E (requer banco e setup completo)
go test ./tests/e2e/... -v -timeout=60s

# Com logs detalhados
go test ./tests/e2e/... -v -timeout=60s -args -test.v
```

#### 5. Benchmarks
```bash
# Executar benchmarks
go test ./tests/benchmarks/ -bench=. -benchmem

# Benchmarks específicos
go test ./tests/benchmarks/ -bench=BenchmarkJWT -benchmem

# Com profiling
go test ./tests/benchmarks/ -bench=. -cpuprofile=cpu.prof
go test ./tests/benchmarks/ -bench=. -memprofile=mem.prof
```

## 📊 Cobertura de Testes

### Funcionalidades Testadas

#### 🔐 Autenticação e Autorização
- ✅ Login com email/password
- ✅ Geração e validação de tokens JWT
- ✅ Middleware de autenticação
- ✅ Controle de acesso baseado em roles
- ✅ Validação de tokens expirados/inválidos

#### 👥 Gerenciamento de Usuários
- ✅ Criação de usuários com validação de permissões
- ✅ Listagem com filtros por empresa e role
- ✅ Atualização de dados de usuários
- ✅ Soft delete de usuários
- ✅ Paginação e ordenação

#### 🏢 Multi-tenancy
- ✅ Isolamento de dados por empresa
- ✅ Hierarquia de permissões (master → company_admin → admin → driver)
- ✅ Validação de acesso entre empresas
- ✅ Contexto de usuário em requests

#### 🗄️ Persistência
- ✅ Operações CRUD no banco de dados
- ✅ Transações e rollbacks
- ✅ Migrations e schema
- ✅ Soft deletes e timestamps

## 🛠️ Utilitários de Teste

### TestDatabase
Helper para operações de banco de dados em testes:
```go
testDB := testutils.NewTestDatabase(db, t)
companyID := testDB.CreateTestCompany("Test Company")
user := testDB.CreateTestUser("John", "john@test.com", "admin", &companyID)
testDB.AssertUserExists(user.ID)
```

### TestDataBuilder
Builder pattern para criar dados de teste complexos:
```go
data := testutils.NewTestDataBuilder(testDB).
    WithCompany("Company A").
    WithRole("admin").
    WithUser("Admin", "admin@test.com", "admin", 0).
    Build()
```

### AssertionHelpers
Helpers para validações comuns:
```go
helpers := testutils.NewAssertionHelpers(t)
helpers.AssertValidUser(user)
helpers.AssertUserBelongsToCompany(user, companyID)
helpers.AssertPaginatedResponse(response, expectedTotal)
```

## 🎯 Melhores Práticas

### 1. Estrutura de Testes
- ✅ Testes organizados por tipo e responsabilidade
- ✅ Uso de test suites para setup/teardown
- ✅ Helpers reutilizáveis para operações comuns
- ✅ Mocks bem estruturados e realistas

### 2. Isolamento
- ✅ Cada teste é independente
- ✅ Cleanup automático de dados de teste
- ✅ Uso de transações para isolamento
- ✅ Mocks para dependências externas

### 3. Performance
- ✅ Testes unitários rápidos (< 10ms cada)
- ✅ Paralelização quando possível
- ✅ Benchmarks para código crítico
- ✅ Profiling para identificar gargalos

### 4. Manutenibilidade
- ✅ Testes descritivos e bem documentados
- ✅ Assertion messages claras
- ✅ Dados de teste realistas
- ✅ Refatoração regular dos testes
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