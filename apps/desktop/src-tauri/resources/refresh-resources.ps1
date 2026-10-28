# Refresh the bundled Python resources for a LOCAL MSI build.
#
# Why this exists: tauri.conf.json maps resources/python-runtime/ and
# resources/python-wheels/ as they are on disk, so a local build bundles
# whatever is there. CI generates both fresh on every run; a local build
# does not, and stale content (an old interpreter, wheels for the wrong
# Python) ships silently — it either fails at install ("Python zip
# missing") or installs a Python the backend rejects.
#
# This script makes the local path equal to the CI path. It reads the
# expected runtime and its SHA from prepare-python.ps1 (the single
# source of truth), fails loudly on any mismatch, and is safe to run
# repeatedly. Run it before every local MSI build:
#
#   pwsh -File apps/desktop/src-tauri/resources/refresh-resources.ps1
#
# Requirements: network access (python.org + PyPI) and a Python with pip
# to resolve the wheels (it may be any version: the wheels are resolved
# for the bundled interpreter).

param(
    [string]$PythonExe = 'C:\Users\wilhe\miniconda3\envs\agent\python.exe'
)

$ErrorActionPreference = 'Stop'

$resources = $PSScriptRoot
$repoRoot = (Resolve-Path (Join-Path $resources '..\..\..\..')).Path
$lockFile = Join-Path $repoRoot 'apps\processor\requirements-lock.txt'
$prepareScript = Join-Path $resources 'prepare-python.ps1'

function Fail($msg) {
    Write-Host "ERROR: $msg" -ForegroundColor Red
    exit 1
}

# ---- 1. Read the expectations from prepare-python.ps1 ----
if (-not (Test-Path $prepareScript)) { Fail "prepare-python.ps1 not found at $prepareScript" }
$prepare = Get-Content $prepareScript -Raw

$zipName = [regex]::Match($prepare, "python-runtime\\(python-[0-9][0-9.]*-amd64\.zip)").Groups[1].Value
if (-not $zipName) { Fail "could not read the expected python zip name from prepare-python.ps1" }
$expectedSha = [regex]::Match($prepare, "\`$expectedSha\s*=\s*'([0-9a-fA-F]{64})'").Groups[1].Value.ToLower()
if (-not $expectedSha) { Fail "could not read the pinned SHA256 from prepare-python.ps1" }

$version = [regex]::Match($zipName, "python-([0-9.]+)-amd64").Groups[1].Value
if (-not $version) { Fail "could not derive the python version from $zipName" }
if (-not (Test-Path $lockFile)) { Fail "requirements-lock.txt not found at $lockFile" }

Write-Host "Expected runtime: $zipName"
Write-Host "Pinned SHA256:    $expectedSha"
Write-Host "Requirements:     $lockFile"

# ---- 2. Download the runtime zip to staging and verify ----
$runtimeDir = Join-Path $resources 'python-runtime'
$wheelsDir = Join-Path $resources 'python-wheels'
$work = Join-Path $env:TEMP "finapp-refresh-$version"
if (Test-Path $work) { Remove-Item $work -Recurse -Force }
New-Item -ItemType Directory -Force -Path $work | Out-Null

try {
    $zipStaging = Join-Path $work $zipName
    $url = "https://www.python.org/ftp/python/$version/$zipName"
    Write-Host "Downloading $url ..."
    Invoke-WebRequest -Uri $url -OutFile $zipStaging -UseBasicParsing

    $actualSha = (Get-FileHash -Path $zipStaging -Algorithm SHA256).Hash.ToLower()
    if ($actualSha -ne $expectedSha) {
        Fail "SHA256 mismatch for $zipName (expected $expectedSha, got $actualSha). Refusing to install it."
    }
    Write-Host "SHA256 verified."

    # ---- 3. Resolve the wheels for the bundled interpreter ----
    if (-not (Test-Path $PythonExe)) { Fail "python not found at $PythonExe (pass -PythonExe to override)" }
    Write-Host "Resolving wheels for Python $version with $PythonExe ..."
    & $PythonExe -m pip download -r $lockFile --dest $work --platform win_amd64 --python-version $version --only-binary=:all: --no-deps
    if ($LASTEXITCODE -ne 0) { Fail "pip download (no-deps) failed" }
    & $PythonExe -m pip download -r $lockFile --dest $work --platform win_amd64 --python-version $version --only-binary=:all:
    if ($LASTEXITCODE -ne 0) { Fail "pip download (with deps) failed" }

    $stagedWheels = Get-ChildItem $work -Filter *.whl
    if (-not $stagedWheels) { Fail "no wheels were downloaded" }
    Write-Host "Resolved $($stagedWheels.Count) wheels ($([math]::Round((($stagedWheels | Measure-Object -Property Length -Sum).Sum)/1MB,1)) MB)."

    # ---- 4. Swap everything into place ----
    # Order matters: install the new set FIRST (same names are overwritten,
    # different names land alongside), then remove what the new set does
    # not include. At every instant the directory holds a complete set of
    # wheels — an interrupted run never leaves a half-populated payload.
    New-Item -ItemType Directory -Force -Path $runtimeDir, $wheelsDir | Out-Null

    Move-Item -Path $zipStaging -Destination (Join-Path $runtimeDir $zipName) -Force

    $newNames = $stagedWheels | ForEach-Object { $_.Name }
    $stagedWheels | ForEach-Object { Move-Item -Path $_.FullName -Destination $wheelsDir -Force }
    Get-ChildItem $wheelsDir -Filter *.whl | Where-Object { $newNames -notcontains $_.Name } | ForEach-Object {
        Remove-Item $_.FullName -Force
    }

    Get-ChildItem $runtimeDir -Filter *.zip | Where-Object { $_.Name -ne $zipName } | ForEach-Object {
        Remove-Item $_.FullName -Force
    }

    Write-Host ""
    Write-Host "Resources refreshed:"
    Write-Host "  $runtimeDir\$zipName ($([math]::Round((Get-Item (Join-Path $runtimeDir $zipName)).Length/1MB,1)) MB)"
    Write-Host "  $wheelsDir\ ($($stagedWheels.Count) wheels)"
    Write-Host "The MSI build can now bundle the correct payload."
} finally {
    if (Test-Path $work) { Remove-Item $work -Recurse -Force -ErrorAction SilentlyContinue }
}
