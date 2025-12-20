@echo off
REM Build LlamaGate for Windows
REM This script builds the llamagate binary

echo ========================================
echo Building LlamaGate...
echo ========================================
echo.

go build -o llamagate.exe ./cmd/llamagate

if %ERRORLEVEL% EQU 0 (
    echo.
    echo Build successful!
    echo Binary created: llamagate.exe
    echo.
    echo To run the binary:
    echo   llamagate.exe
    echo.
    echo Or set environment variables and run:
    echo   set OLLAMA_HOST=http://localhost:11434
    echo   set API_KEY=sk-llamagate
    echo   llamagate.exe
) else (
    echo.
    echo Build failed!
    exit /b %ERRORLEVEL%
)

