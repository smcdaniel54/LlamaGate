@echo off
REM LlamaGate Installer Launcher
REM This script launches the PowerShell installer

echo ========================================
echo LlamaGate Installer
echo ========================================
echo.

REM Check if PowerShell is available
powershell -Command "exit 0" >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo Error: PowerShell is required but not found.
    echo Please install PowerShell and try again.
    pause
    exit /b 1
)

REM Launch the PowerShell installer
powershell -ExecutionPolicy Bypass -File "%~dp0install.ps1"

if %ERRORLEVEL% EQU 0 (
    echo.
    echo Installation completed successfully!
) else (
    echo.
    echo Installation failed. Please check the errors above.
)

pause

