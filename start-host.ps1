# Compatibility entry point. The maintained Windows supervisor lives under scripts/windows.
& (Join-Path $PSScriptRoot "scripts\windows\run-pinharbor.ps1")
exit $LASTEXITCODE
