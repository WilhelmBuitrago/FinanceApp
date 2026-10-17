# Prepare Python runtime for FinanceApp during MSI installation
# Runs with elevated privileges

$ErrorActionPreference = 'Stop'

$installDir = $env:FIXAPP_INSTALLDIR
if (-not $installDir) {
    $installDir = $env:ProgramFiles\FinanceApp
}

$resourceDir = Join-Path $installDir 'resources'
$pythonZip = Join-Path $resourceDir 'python-runtime\python-3.12.10-amd64.zip'
$wheelsDir = Join-Path $resourceDir 'python-wheels'
$processorReq = Join-Path $resourceDir 'processor\requirements-lock.txt'

$targetBase = Join-Path $env:ProgramData 'FinanceApp'
$runtimeDir = Join-Path $targetBase 'python-runtime\3.12.10'
$venvDir = Join-Path $targetBase 'venv'

Write-Host "FinanceApp preparation: installDir=$installDir targetBase=$targetBase"

New-Item -ItemType Directory -Force -Path $targetBase | Out-Null
New-Item -ItemType Directory -Force -Path $runtimeDir | Out-Null

# Extract Python zip if not already extracted
$pythonExe = Join-Path $runtimeDir 'python.exe'
if (-not (Test-Path $pythonExe)) {
    Write-Host "Extracting Python runtime..."
    Expand-Archive -Path $pythonZip -DestinationPath $runtimeDir -Force
} else {
    Write-Host "Python runtime already present"
}

# Create venv if missing
$venvPython = Join-Path $venvDir 'Scripts\python.exe'
if (-not (Test-Path $venvPython)) {
    Write-Host "Creating virtual environment..."
    & $pythonExe -m venv $venvDir
}

# Upgrade pip
Write-Host "Upgrading pip..."
& $venvPython -m pip install --upgrade pip --no-index --find-links $wheelsDir

# Install requirements offline
Write-Host "Installing processor requirements offline..."
& $venvPython -m pip install --no-index --find-links $wheelsDir -r $processorReq

# Verify imports
Write-Host "Verifying imports..."
& $venvPython -c "import pdfplumber, pydantic, fitz"

# Write marker
$marker = Join-Path $targetBase '.python-ready'
Set-Content -Path $marker -Value '3.12.10'

# Write log
$logDir = Join-Path $targetBase 'logs'
New-Item -ItemType Directory -Force -Path $logDir | Out-Null
$logFile = Join-Path $logDir 'install.log'
Add-Content -Path $logFile -Value "Installation completed at $(Get-Date)"

Write-Host "FinanceApp Python preparation completed successfully"
exit 0
