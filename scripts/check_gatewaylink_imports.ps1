# 检查 gatewaylink 导入的脚本 (PowerShell)
# 禁止除白名单外的包引用 gatewaylink

$ErrorActionPreference = "Stop"

$gameserverDir = "server/service/gameserver"
$rootDir = Split-Path -Parent $PSScriptRoot

# 允许的文件列表（白名单）
# PlayerActor 相关文件允许使用 gatewaylink
$allowedFiles = @(
    "$gameserverDir/internel/app/playeractor/entity/player_network.go",
    "$gameserverDir/internel/app/playeractor/entity/player_role.go",
    "$gameserverDir/internel/adapter/gateway/network_gateway.go",
    "$gameserverDir/internel/adapter/gateway/session_gateway.go",
    "$gameserverDir/internel/infrastructure/gatewaylink/handler.go",
    "$gameserverDir/internel/infrastructure/gatewaylink/sender.go",
    "$gameserverDir/internel/infrastructure/gatewaylink/export.go",
    "$gameserverDir/internel/app/engine/server.go",
    "$gameserverDir/internel/infrastructure/dungeonserverlink/dungeon_cli.go"
)

Write-Host "Checking gatewaylink imports..."

# 查找所有引用 gatewaylink 的文件
$allGoFiles = Get-ChildItem -Path (Join-Path $rootDir $gameserverDir) -Recurse -Include "*.go" -ErrorAction SilentlyContinue
$violations = @()

foreach ($file in $allGoFiles) {
    $content = Get-Content $file.FullName -Raw -ErrorAction SilentlyContinue
    if ($content -match "postapocgame/server/service/gameserver/internel/infrastructure/gatewaylink") {
        $relativePath = $file.FullName.Replace($rootDir + "\", "").Replace("\", "/")
        $isAllowed = $false
        
        foreach ($allowedFile in $allowedFiles) {
            $fullAllowedPath = Join-Path $rootDir $allowedFile
            if ($file.FullName -eq $fullAllowedPath -or $relativePath -like "*$allowedFile*") {
                $isAllowed = $true
                break
            }
        }
        
        if (-not $isAllowed) {
            $violations += $relativePath
        }
    }
}

if ($violations.Count -eq 0) {
    Write-Host "✅ All gatewaylink imports are in allowed files" -ForegroundColor Green
    exit 0
} else {
    Write-Host "❌ Unauthorized gatewaylink imports found:" -ForegroundColor Red
    foreach ($violation in $violations) {
        Write-Host "  - $violation" -ForegroundColor Red
    }
    Write-Host ""
    Write-Host "Only the following files are allowed to import gatewaylink:"
    foreach ($allowedFile in $allowedFiles) {
        Write-Host "  - $allowedFile"
    }
    exit 1
}

