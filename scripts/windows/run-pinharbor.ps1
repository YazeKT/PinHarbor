param(
    [switch]$NoBrowser
)

$ErrorActionPreference = "Stop"

$appDir = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$binaryDir = Join-Path $appDir "bin"
$dataDir = Join-Path $appDir "webdata"
$binaryPath = Join-Path $binaryDir "pinharbor.exe"
$nextBinaryPath = Join-Path $binaryDir "pinharbor-next.exe"
$readyMarker = Join-Path $dataDir ".playwright-ready"
$healthUrl = "http://127.0.0.1:8080/health"
$appUrl = "http://localhost:8080"

$host.UI.RawUI.WindowTitle = "PinHarbor"
Set-Location -LiteralPath $appDir

Write-Host ""
Write-Host "  PINHARBOR" -ForegroundColor Green
Write-Host "  Local lead discovery" -ForegroundColor Cyan
Write-Host ""

try {
    $health = Invoke-RestMethod -Uri $healthUrl -TimeoutSec 2
    if ($health.status -eq "ok" -and $health.app -eq "PinHarbor") {
        Write-Host "PinHarbor is already running at $appUrl" -ForegroundColor Green
        if (-not $NoBrowser) {
            Start-Process $appUrl
        }
        exit 0
    }
} catch {
    # No healthy PinHarbor instance is listening; prepare a fresh local build.
}

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "Go is required. Install it from https://go.dev/dl/ and run this launcher again." -ForegroundColor Red
    exit 1
}

New-Item -ItemType Directory -Force -Path $binaryDir, $dataDir | Out-Null

Write-Host "Building the current PinHarbor source..." -ForegroundColor Cyan
& go build -o $nextBinaryPath .
if ($LASTEXITCODE -ne 0) {
    throw "PinHarbor build failed."
}
Move-Item -LiteralPath $nextBinaryPath -Destination $binaryPath -Force

if (-not (Test-Path -LiteralPath $readyMarker)) {
    Write-Host "Installing the browser engine (first run only)..." -ForegroundColor Cyan
    $env:PLAYWRIGHT_INSTALL_ONLY = "1"
    & $binaryPath
    $installExit = $LASTEXITCODE
    Remove-Item Env:PLAYWRIGHT_INSTALL_ONLY -ErrorAction SilentlyContinue
    if ($installExit -ne 0) {
        throw "The browser engine installation failed."
    }
    New-Item -ItemType File -Force -Path $readyMarker | Out-Null
}

$openedBrowser = $false
$restartCount = 0

while ($true) {
    Write-Host ""
    Write-Host "Starting PinHarbor at $appUrl" -ForegroundColor Green
    Write-Host "Keep this window open. PinHarbor will restart automatically if it exits." -ForegroundColor DarkGray

    $process = Start-Process -FilePath $binaryPath -ArgumentList @("-web", "-data-folder", $dataDir) -WorkingDirectory $appDir -NoNewWindow -PassThru

    if (-not $openedBrowser) {
        $deadline = (Get-Date).AddSeconds(45)
        while ((Get-Date) -lt $deadline -and -not $process.HasExited) {
            try {
                $health = Invoke-RestMethod -Uri $healthUrl -TimeoutSec 2
                if ($health.status -eq "ok") {
                    if (-not $NoBrowser) {
                        Start-Process $appUrl
                    }
                    $openedBrowser = $true
                    break
                }
            } catch {
                Start-Sleep -Milliseconds 600
            }
        }
    }

    $process.WaitForExit()
    $restartCount++
    Write-Host "PinHarbor exited with code $($process.ExitCode). Restarting in 3 seconds (restart $restartCount)..." -ForegroundColor Yellow
    Start-Sleep -Seconds 3
}
