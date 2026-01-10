@echo off
REM LlamaGate Runner with Authentication Enabled
REM This script runs LlamaGate with API key authentication

REM Set environment variables
set OLLAMA_HOST=http://localhost:11434
set API_KEY=sk-llamagate
set RATE_LIMIT_RPS=50
set DEBUG=false
set PORT=8080

REM Display configuration
echo ========================================
echo LlamaGate - OpenAI-Compatible Proxy
echo ========================================
echo.
echo Configuration:
echo   OLLAMA_HOST: %OLLAMA_HOST%
echo   API_KEY: %API_KEY% (Authentication enabled)
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

