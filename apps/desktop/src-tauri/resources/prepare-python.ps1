# Prepare Python runtime for FinanceApp during MSI installation
# Runs with elevated privileges

param(
    [string]$InstallDir = ''
)

$ErrorActionPreference = 'Stop'

$mutexName = 'Global\FinanceAppPythonPrepare'
$mutex = New-Object System.Threading.Mutex($false, $mutexName)
$acquired = $false
try {
    try {
        $acquired = $mutex.WaitOne(30000)
    } catch [System.Threading.AbandonedMutexException] {
        Write-Host "WARNING: Mutex abandoned, proceeding with validation"
        $acquired = $true
    }
    if (-not $acquired) {
        Write-Host "ERROR: Could not acquire preparation lock, another instance is running"
        exit 1
    }

    # Resolve install dir
    $installDir = $InstallDir
    if (-not $installDir) { $installDir = $env:FINAPP_INSTALLDIR }
    if (-not $installDir) { $installDir = Join-Path $env:ProgramFiles 'FinanceApp' }

    $resourceDir = Join-Path $installDir 'resources'
    $pythonZip = Join-Path $resourceDir 'python-runtime\python-3.13.17-amd64.zip'
    $wheelsDir = Join-Path $resourceDir 'python-wheels'
    $processorReq = Join-Path $resourceDir 'processor\requirements-lock.txt'

    $targetBase = Join-Path $env:ProgramData 'FinanceApp'
    $runtimeDir = Join-Path $targetBase 'python-runtime\3.13.17'
    $venvsRoot = Join-Path $targetBase 'venvs'
    $logDir = Join-Path $targetBase 'logs'
    New-Item -ItemType Directory -Force -Path $logDir | Out-Null
    $logFile = Join-Path $logDir 'install.log'

    function Write-Log($msg) {
        $ts = Get-Date -Format 'yyyy-MM-dd HH:mm:ss'
        $line = "[$ts] $msg"
        Write-Host $line
        Add-Content -Path $logFile -Value $line
    }

    Write-Log "FinanceApp preparation started. installDir=$installDir targetBase=$targetBase"

    if (-not (Test-Path $pythonZip)) { Write-Log "ERROR: Python zip missing $pythonZip"; exit 1 }
    if (-not (Test-Path $wheelsDir)) { Write-Log "ERROR: Wheels dir missing $wheelsDir"; exit 1 }
    if (-not (Test-Path $processorReq)) { Write-Log "ERROR: requirements-lock missing $processorReq"; exit 1 }

    $expectedSha = 'b26ca4b4ee51e157836049934d2b8eabd55939be296f3130254251b6df96ec2f'
    $actualSha = (Get-FileHash -Path $pythonZip -Algorithm SHA256).Hash.ToLower()
    if ($actualSha -ne $expectedSha) {
        Write-Log "ERROR: SHA256 mismatch for python zip"; exit 1
    }
    Write-Log "Python zip SHA256 verified"

    $reqHash = (Get-FileHash -Path $processorReq -Algorithm SHA256).Hash.ToLower()
    Write-Log "Requirements lock hash $reqHash"

    $identifier = "3.13.17-$reqHash"
    $venvDir = Join-Path $venvsRoot $identifier
    $marker = Join-Path $targetBase '.python-ready'

    if (Test-Path $marker) {
        $markerContent = Get-Content $marker -Raw
        $existingVersion = ($markerContent | Select-String -Pattern 'version=(.*)' -AllMatches).Matches[0].Groups[1].Value
        $existingHash = ($markerContent | Select-String -Pattern 'requirements_hash=(.*)' -AllMatches).Matches[0].Groups[1].Value
        $existingId = ($markerContent | Select-String -Pattern 'identifier=(.*)' -AllMatches).Matches[0].Groups[1].Value
        if ($existingVersion -eq '3.13.17' -and $existingHash -eq $reqHash -and $existingId -eq $identifier) {
            $venvPython = Join-Path $venvDir 'Scripts\python.exe'
            if (Test-Path $venvPython) {
                Write-Log "Existing environment matches, validating integrity..."
                & $venvPython -c "import pdfplumber, pydantic, fitz"
                if ($LASTEXITCODE -eq 0) {
                    Write-Log "Existing environment integrity OK. Skipping preparation."
                    exit 0
                }
            }
        } else {
            Write-Log "Existing environment outdated or mismatched. Will prepare new environment."
        }
    }

    New-Item -ItemType Directory -Force -Path $runtimeDir | Out-Null
    $pythonExe = Join-Path $runtimeDir 'python.exe'
    if (-not (Test-Path $pythonExe)) {
        Write-Log "Extracting Python runtime..."
        Expand-Archive -Path $pythonZip -DestinationPath $runtimeDir -Force
    }

    $venvPython = Join-Path $venvDir 'Scripts\python.exe'
    if (-not (Test-Path $venvPython)) {
        Write-Log "Creating virtual environment..."
        & $pythonExe -m venv $venvDir
        if ($LASTEXITCODE -ne 0) { Write-Log "ERROR: venv creation failed"; exit 1 }
    }

    Write-Log "Upgrading pip..."
    & $venvPython -m pip install --upgrade pip --no-index --find-links $wheelsDir
    if ($LASTEXITCODE -ne 0) { Write-Log "ERROR: pip upgrade failed"; exit 1 }

    Write-Log "Installing requirements offline..."
    & $venvPython -m pip install --no-index --find-links $wheelsDir -r $processorReq
    if ($LASTEXITCODE -ne 0) { Write-Log "ERROR: pip install failed"; exit 1 }

    Write-Log "Verifying imports..."
    & $venvPython -c "import pdfplumber, pydantic, fitz"
    if ($LASTEXITCODE -ne 0) { Write-Log "ERROR: import verification failed"; exit 1 }

    $installed = & $venvPython -m pip freeze | Select-String -Pattern 'pdfplumber==0.11.10|pydantic==2.11.7|pymupdf==1.26.3'
    if (-not $installed) { Write-Log "ERROR: version verification failed"; exit 1 }
    Write-Log "Version verification passed"

    $markerTmp = Join-Path $targetBase '.python-ready.tmp'
    $markerContent = @"
version=3.13.17
requirements_hash=$reqHash
identifier=$identifier
venv_path=$venvDir
marker_format=1
"@
    Set-Content -Path $markerTmp -Value $markerContent
    Move-Item -Path $markerTmp -Destination $marker -Force
    Write-Log "Marker written"

    Write-Log "FinanceApp Python preparation completed successfully"
    exit 0
} finally {
    if ($acquired) {
        try { $mutex.ReleaseMutex() } catch {}
    }
    if ($mutex) { $mutex.Dispose() }
}
