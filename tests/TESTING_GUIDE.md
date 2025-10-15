# Testing Strategy - DashTrack API

## 📋 Estrutura de Testes

```
tests/
├── unit/              # Testes unitários (isolados, com mocks)
├── integration/       # Testes de integração (múltiplos componentes)
├── e2e/              # Testes end-to-end (sistema completo)
├── benchmarks/       # Testes de performance
└── testutils/        # Utilitários compartilhados
```

## 🚀 Executando Testes

### Todos os testes
```bash
go test ./... -v
```

### Testes por categoria
```bash
# Unit tests apenas
go test ./tests/unit/... -v

# Integration tests apenas
go test ./tests/integration/... -v

# E2E tests apenas
go test ./tests/e2e/... -v
```

### Testes específicos
```bash
# Rodar uma suite específica
go test ./tests/integration -run TestTeamMembersIntegrationSuite -v

# Rodar um teste específico
go test ./tests/integration -run TestTeamMembersIntegrationSuite/TestAddTeamMember -v
```

### Com coverage
```bash
go test ./... -cover
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Benchmarks
```bash
go test ./tests/benchmarks/... -bench=. -benchmem
```

## 📝 Convenções

### Nomenclatura
- **Unit tests**: `package_test.go` (ex: `user_test.go`)
- **Integration tests**: `feature_integration_test.go`
- **E2E tests**: `workflow_e2e_test.go`

### Estrutura de Suite
```go
type FeatureTestSuite struct {
    suite.Suite
    // dependencies
}

func TestFeatureSuite(t *testing.T) {
    suite.Run(t, new(FeatureTestSuite))
}

func (s *FeatureTestSuite) SetupSuite() {
    // Setup once before all tests
}

func (s *FeatureTestSuite) SetupTest() {
    // Setup before each test
}

func (s *FeatureTestSuite) TearDownTest() {
    // Cleanup after each test
}

func (s *FeatureTestSuite) TearDownSuite() {
    // Cleanup once after all tests
}

func (s *FeatureTestSuite) TestSomething() {
    // Test implementation
}
```

## 🎯 Testes Implementados

### ✅ Task 1: Team Members Management API
**Arquivo**: `tests/integration/team_members_test.go`

Testes:
- `TestAddTeamMember` - Adicionar membro ao time
- `TestGetTeamMembers` - Listar membros do time
- `TestUpdateMemberRole` - Atualizar role do membro
- `TestTransferMember` - Transferir membro para outro time
- `TestRemoveMember` - Remover membro do time
- `TestCompleteWorkflow` - Fluxo completo

### ✅ Task 2: Vehicle Assignment History
**Arquivo**: `tests/integration/vehicle_assignment_history_test.go`

Testes:
- `TestUpdateDriverAssignment` - Atualizar motorista
- `TestUpdateHelperAssignment` - Atualizar ajudante
- `TestGetAssignmentHistory` - Recuperar histórico
- `TestAutomaticHistoryCreation` - Verificar criação automática
- `TestCompleteWorkflow` - Fluxo completo

### ✅ Task 3: Team Member History
**Arquivo**: `tests/integration/team_member_history_test.go`

Testes:
- `TestAddMemberCreatesHistory` - Histórico ao adicionar
- `TestUpdateRoleCreatesHistory` - Histórico ao atualizar role
- `TestRemoveMemberCreatesHistory` - Histórico ao remover
- `TestGetTeamMemberHistory` - Recuperar histórico do time
- `TestGetUserTeamHistory` - Recuperar histórico do usuário
- `TestCompleteWorkflow` - Fluxo completo

## 🔧 Setup de Ambiente de Teste

### Banco de Dados de Teste
```bash
# Docker compose para testes
docker-compose -f docker-compose.test.yml up -d

# Rodar migrations
make migrate-test
```

### Variáveis de Ambiente
```env
TEST_DB_HOST=localhost
TEST_DB_PORT=5433
TEST_DB_NAME=dashtrack_test
TEST_DB_USER=postgres
TEST_DB_PASSWORD=postgres
```

## 📊 Métricas de Qualidade

### Cobertura Mínima
- **Unit tests**: 80%
- **Integration tests**: 70%
- **Total**: 75%

### Performance
- **Unit tests**: < 100ms por teste
- **Integration tests**: < 1s por teste
- **E2E tests**: < 5s por teste

## 🐛 Debugging

### Logs Verbosos
```bash
go test ./... -v -count=1
```

### Rodar teste específico com debug
```bash
go test ./tests/integration -run TestTeamMembersIntegrationSuite/TestAddTeamMember -v -count=1
```

### Com race detector
```bash
go test ./... -race
```

## 📚 Recursos

- [Testing package](https://pkg.go.dev/testing)
- [Testify suite](https://pkg.go.dev/github.com/stretchr/testify/suite)
- [Testify assert](https://pkg.go.dev/github.com/stretchr/testify/assert)
- [Testify mock](https://pkg.go.dev/github.com/stretchr/testify/mock)

## ✅ Checklist para Novos Testes

- [ ] Teste tem nome descritivo
- [ ] Teste é independente (não depende de ordem)
- [ ] Teste tem assertions claras
- [ ] Teste limpa recursos (TearDown)
- [ ] Teste documenta o cenário (comentários)
- [ ] Teste cobre casos de erro
- [ ] Teste cobre happy path
- [ ] Teste é rápido (< limites acima)
