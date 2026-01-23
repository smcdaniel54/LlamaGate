@echo off
REM LlamaGate Runner for Windows
REM This script runs LlamaGate with configurable environment variables
REM Note: If a .env file exists, it will be loaded automatically
REM Environment variables set here will override .env file values

REM Change to project root directory (where this script is located)
cd /d "%~dp0\..\.."

REM Set default environment variables if not already set
if "%OLLAMA_HOST%"=="" set OLLAMA_HOST=http://localhost:11434
if "%API_KEY%"=="" set API_KEY=
if "%RATE_LIMIT_RPS%"=="" set RATE_LIMIT_RPS=50
if "%DEBUG%"=="" set DEBUG=false
if "%PORT%"=="" set PORT=11435

REM Display configuration
echo ========================================
echo LlamaGate - OpenAI-Compatible Proxy
echo ========================================
echo.
if exist .env (
    echo Configuration loaded from .env file
    echo (Environment variables override .env values)
    echo.
) else (
    echo Tip: Create a .env file for easier configuration
    echo.
)
echo Configuration:
echo   OLLAMA_HOST: %OLLAMA_HOST%
echo   API_KEY: %API_KEY%
if "%API_KEY%"=="" (
    echo     (Authentication disabled)
) else (
    echo     (Authentication enabled)
)
echo   RATE_LIMIT_RPS: %RATE_LIMIT_RPS%
echo   DEBUG: %DEBUG%
echo   PORT: %PORT%
echo.
echo Starting server...
echo Press Ctrl+C to stop
echo ========================================
echo.

REM Run the application
go run ./cmd/llamagate

