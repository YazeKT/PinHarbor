@echo off
setlocal
set "APP_DIR=%~dp0"
set "APP_DIR=%APP_DIR:~0,-1%"
title PinHarbor
powershell.exe -NoLogo -NoProfile -ExecutionPolicy Bypass -File "%APP_DIR%\scripts\windows\run-pinharbor.ps1"
if errorlevel 1 (
    echo.
    echo PinHarbor could not start. Review the message above.
    pause
)
endlocal
