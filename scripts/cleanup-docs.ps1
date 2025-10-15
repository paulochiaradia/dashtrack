# Script de Limpeza - Remover Arquivos MD e PS1 Desnecessarios
Write-Host "=== DashTrack - Limpeza de Arquivos Desnecessarios ===" -ForegroundColor Cyan
Write-Host ""

$rootPath = "C:\Users\paulo\dashtrack"
$filesRemoved = 0

# Arquivos MD a remover (raiz)
$mdFilesToRemove = @(
    "AUDIT_ARCHITECTURE.md",
    "AUDIT_COMPLETE.md",
    "AUDIT_MIDDLEWARE_COMPLETE.md",
    "AUDIT_PROGRESS.md",
    "AUDIT_PROMETHEUS_COMPLETE.md",
    "AUDIT_QUICK_REFERENCE.md",
    "AUDIT_SESSION_SUMMARY.md",
    "AUDIT_STATUS.md",
    "AUDIT_TESTING_GUIDE.md",
    "BUG_FIXES_REPORT.md",
    "CODE_OPTIMIZATION_SUMMARY.md",
    "ENDPOINT_TESTING_REPORT.md",
    "FINAL_TEST_REPORT.md",
    "GUIA_TESTES_ATUALIZADO.md",
    "IMPLEMENTATION_ROADMAP.md",
    "JAEGER_V2_UPGRADE.md",
    "MONITORING_GUIDE.md",
    "PASSWORD_RECOVERY_IMPLEMENTATION.md",
    "PERMISSION_CHANGES_SUMMARY.md",
    "PHASE_2_COMPLETE.md",
    "PHASE_3_COMPLETE.md",
    "PHASE_4_COMPLETE.md",
    "SESSION_LIMIT_IMPLEMENTATION.md",
    "SESSION_LIMIT_TESTING.md",
    "SYSTEM_DOCUMENTATION.md",
    "TASK_3_FINAL_REPORT.md",
    "TASK_3_TEAM_MEMBER_HISTORY_COMPLETE.md",
    "TEAM_MANAGEMENT_API.md",
    "TEAM_MANAGEMENT_PROGRESS.md",
    "TEAM_MANAGEMENT_TESTING_GUIDE.md",
    "TEAM_MEMBERS_API.md",
    "TEAM_MEMBER_HISTORY_API.md",
    "TEAM_VEHICLE_INTEGRATION.md",
    "TEST_FIXES_SUMMARY.md",
    "TESTING_MANUAL.md",
    "TESTING.md",
    "TOKEN_SYSTEM_REFACTOR.md",
    "VEHICLE_ASSIGNMENT_HISTORY_API.md",
    "VEHICLE_MANAGEMENT_API.md",
    "CLEANUP_PLAN.md"
)

# Scripts PS1 a remover
$ps1FilesToRemove = @(
    "scripts\test-task1-final.ps1",
    "scripts\test-task1-simple.ps1",
    "scripts\test-task2-final.ps1",
    "scripts\test-task3-final.ps1",
    "scripts\test-team-members-api.ps1",
    "scripts\test-vehicle-assignment-history.ps1"
)

# Remover README antigo do tests
$testReadme = "tests\README.md"

Write-Host ""
Write-Host "Removendo arquivos MD desnecessarios..." -ForegroundColor Yellow
foreach ($file in $mdFilesToRemove) {
    $fullPath = Join-Path $rootPath $file
    if (Test-Path $fullPath) {
        Remove-Item $fullPath -Force
        Write-Host "  [OK] Removido: $file" -ForegroundColor Green
        $filesRemoved++
    }
}

Write-Host ""
Write-Host "Removendo scripts PowerShell de teste..." -ForegroundColor Yellow
foreach ($file in $ps1FilesToRemove) {
    $fullPath = Join-Path $rootPath $file
    if (Test-Path $fullPath) {
        Remove-Item $fullPath -Force
        Write-Host "  [OK] Removido: $file" -ForegroundColor Green
        $filesRemoved++
    }
}

Write-Host ""
Write-Host "Removendo README antigo do tests..." -ForegroundColor Yellow
$testReadmePath = Join-Path $rootPath $testReadme
if (Test-Path $testReadmePath) {
    Remove-Item $testReadmePath -Force
    Write-Host "  [OK] Removido: $testReadme" -ForegroundColor Green
    $filesRemoved++
}

Write-Host ""
Write-Host "==================================" -ForegroundColor Cyan
Write-Host "==================================" -ForegroundColor Cyan
Write-Host "LIMPEZA CONCLUIDA" -ForegroundColor Green
Write-Host "==================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Arquivos removidos: $filesRemoved" -ForegroundColor Yellow
Write-Host ""
Write-Host "Arquivos mantidos:" -ForegroundColor Cyan
Write-Host "  [OK] README.md (raiz)" -ForegroundColor Green
Write-Host "  [OK] SECURITY.md" -ForegroundColor Green
Write-Host "  [OK] tests/TESTING_GUIDE.md" -ForegroundColor Green
Write-Host ""
Write-Host "Proximos passos:" -ForegroundColor Yellow
Write-Host "  1. Revisar mudancas: git status" -ForegroundColor White
Write-Host "  2. Commit: git add -A && git commit -m 'Clean up unnecessary MD and PS1 files'" -ForegroundColor White
Write-Host "  3. Executar testes Go: go test ./... -v" -ForegroundColor White
Write-Host ""
