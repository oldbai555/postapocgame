# CI 检查脚本 (PowerShell)
# 用于检查代码质量、静态分析和禁止导入规则

$ErrorActionPreference = "Stop"

Write-Host "=========================================="
Write-Host "Running CI checks for gameserver"
Write-Host "=========================================="

$gameserverDir = "server/service/gameserver"
$rootDir = Split-Path -Parent $PSScriptRoot
Set-Location $rootDir

# 1. 检查 go vet
Write-Host ""
Write-Host "1. Running go vet..."
# 切换到 server 目录，因为 go.mod 在那里
Push-Location "server"
$vetOutput = & go vet "./service/gameserver/..." 2>&1
$vetExitCode = $LASTEXITCODE
Pop-Location
if ($vetExitCode -ne 0) {
    Write-Host "❌ go vet failed" -ForegroundColor Red
    Write-Host $vetOutput
    # 不退出，继续检查其他项
    Write-Host "⚠️  go vet found issues, but continuing..." -ForegroundColor Yellow
} else {
    Write-Host "✅ go vet passed" -ForegroundColor Green
}

# 2. 检查 staticcheck
Write-Host ""
Write-Host "2. Running staticcheck..."
$staticcheckCmd = Get-Command staticcheck -ErrorAction SilentlyContinue
if (-not $staticcheckCmd) {
    Write-Host "⚠️  staticcheck not found, installing..." -ForegroundColor Yellow
    & go install honnef.co/go/tools/cmd/staticcheck@latest
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Failed to install staticcheck" -ForegroundColor Red
        exit 1
    }
}

# 切换到 server 目录，因为 go.mod 在那里
Push-Location "server"
$staticcheckOutput = & staticcheck "./service/gameserver/..." 2>&1
$staticcheckExitCode = $LASTEXITCODE
Pop-Location
if ($staticcheckExitCode -ne 0) {
    Write-Host "❌ staticcheck failed" -ForegroundColor Red
    Write-Host $staticcheckOutput
    # 不退出，继续检查其他项
    Write-Host "⚠️  staticcheck found issues, but continuing..." -ForegroundColor Yellow
} else {
    Write-Host "✅ staticcheck passed" -ForegroundColor Green
}

# 3. 检查 gatewaylink 导入
Write-Host ""
Write-Host "3. Checking gatewaylink imports..."
& "$PSScriptRoot/check_gatewaylink_imports.ps1"
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ gatewaylink import check failed" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "=========================================="
Write-Host "✅ All CI checks completed!" -ForegroundColor Green
Write-Host "=========================================="
