param(
    [string]$Version = "2.2",
    [string]$AppName = "NeuralPath Tactical Guard",
    [string]$ExeName = "NeuralPathTacticalGuard.exe"
)

$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $Root

$Artifacts = Join-Path $ProjectRoot "artifacts"
$StageRoot = Join-Path $Artifacts "stage"
$StageDir = Join-Path $StageRoot $AppName
$OutDir = Join-Path $Artifacts "out"

$ZipPath = Join-Path $OutDir "NeuralPath-Tactical-Guard-$Version-win-x64-portable.zip"

Remove-Item $StageRoot -Recurse -Force -ErrorAction SilentlyContinue
Remove-Item $OutDir -Recurse -Force -ErrorAction SilentlyContinue

New-Item -ItemType Directory -Path $StageDir | Out-Null
New-Item -ItemType Directory -Path $OutDir | Out-Null

# ── Step 1: Compilazione Go ──────────────────────────────────────────
Write-Host "[1/3] Compilazione Go (windows/amd64)..." -ForegroundColor Yellow
Set-Location $ProjectRoot
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "1"
$OutputExe = Join-Path $ProjectRoot $ExeName
go build -ldflags="-s -w -H windowsgui" -o $OutputExe .
if (-not (Test-Path $OutputExe)) {
    Write-Host "ERRORE: build Go fallita, .exe non trovato." -ForegroundColor Red
    exit 1
}
Write-Host "  OK -> $ExeName compilato" -ForegroundColor Green

# ── Step 2: Staging file ───────────────────────────────────────────────
Write-Host "[2/3] Staging file..." -ForegroundColor Yellow
Copy-Item $OutputExe                               $StageDir -Force
Copy-Item (Join-Path $ProjectRoot "wintun.dll")    $StageDir -Force -ErrorAction SilentlyContinue
Copy-Item (Join-Path $ProjectRoot "Satellite.ico") $StageDir -Force -ErrorAction SilentlyContinue

$excludePatterns = @(
    "*.pdb", "*.ipdb", "*.iobj",
    "appsettings.Development.json",
    ".env", ".env.*",
    "*keygen*",
    "*test*",
    "*debug*",
    "*sandbox*",
    "*fake-license*"
)

foreach ($pattern in $excludePatterns) {
    Get-ChildItem -Path $StageDir -Recurse -Force -ErrorAction SilentlyContinue |
    Where-Object { $_.Name -like $pattern } |
    Remove-Item -Recurse -Force -ErrorAction SilentlyContinue
}

$releaseFiles = @(
    (Join-Path $ProjectRoot "relase\EULA.txt"),
    (Join-Path $ProjectRoot "relase\LICENSE.txt"),
    (Join-Path $ProjectRoot "relase\README.txt")
)

foreach ($file in $releaseFiles) {
    if (Test-Path $file) {
        Copy-Item $file $StageDir -Force
    }
}

$manifest = [ordered]@{
    app       = $AppName
    version   = $Version
    exe       = $ExeName
    buildTime = (Get-Date).ToString("s")
}

$manifest | ConvertTo-Json -Depth 3 | Set-Content (Join-Path $StageDir "release-manifest.json") -Encoding UTF8

Write-Host ""
Write-Host "[3/3] Creazione ZIP portabile (attesa rilascio lock Antivirus)..." -ForegroundColor Yellow
Start-Sleep -Seconds 2
Compress-Archive -Path (Join-Path $StageDir "*") -DestinationPath $ZipPath -Force

Get-FileHash $ZipPath -Algorithm SHA256 |
ForEach-Object { "$($_.Algorithm)  $($_.Hash)  $(Split-Path $_.Path -Leaf)" } |
Set-Content (Join-Path $OutDir "SHA256.txt") -Encoding UTF8

# ── Step 4: Inno Setup Installer ─────────────────────────────────────────
Write-Host ""
Write-Host "[4/4] Creazione installer con Inno Setup..." -ForegroundColor Yellow

$InnoSetup = "C:\Program Files (x86)\Inno Setup 6\ISCC.exe"
if (-not (Test-Path $InnoSetup)) {
    $InnoSetup = "$env:LOCALAPPDATA\Programs\Inno Setup 6\ISCC.exe"
}
$IssFile = Join-Path $PSScriptRoot "NeuralPathTacticalGuard.iss"

if (-not (Test-Path $InnoSetup)) {
    Write-Host "  ATTENZIONE: Inno Setup non trovato, skip installer." -ForegroundColor DarkYellow
    Write-Host "  Installa Inno Setup 6 da https://jrsoftware.org/isinfo.php" -ForegroundColor DarkYellow
}
else {
    & "$InnoSetup" "$IssFile"
    Write-Host "  OK -> Installer generato in artifacts\installer\" -ForegroundColor Green
}

Write-Host ""
Write-Host "===================================" -ForegroundColor Green
Write-Host "  BUILD COMPLETATA CON SUCCESSO!  " -ForegroundColor Green
Write-Host "===================================" -ForegroundColor Green
Write-Host "Stage    : $StageDir"                              -ForegroundColor Cyan
Write-Host "ZIP      : $ZipPath"                               -ForegroundColor Cyan
Write-Host "SHA256   : $(Join-Path $OutDir 'SHA256.txt')"      -ForegroundColor Cyan
Write-Host "Installer: $(Join-Path $ProjectRoot 'artifacts\installer\')" -ForegroundColor Cyan