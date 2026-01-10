@echo off
REM LlamaGate Runner with Debug Mode Enabled
REM This script runs LlamaGate with debug logging enabled

REM Set environment variables
set OLLAMA_HOST=http://localhost:11434
set API_KEY=
set RATE_LIMIT_RPS=50
set DEBUG=true
set PORT=8080

REM Display configuration
echo ========================================
echo LlamaGate - OpenAI-Compatible Proxy
echo ========================================
echo.
echo Configuration:
echo   OLLAMA_HOST: %OLLAMA_HOST%
echo   API_KEY: (Authentication disabled)
echo   RATE_LIMIT_RPS: %RATE_LIMIT_RPS%
echo   DEBUG: %DEBUG% (Debug logging enabled)
echo   PORT: %PORT%
echo.
echo Starting server...
echo Press Ctrl+C to stop
echo ========================================
echo.

REM Run the application
go run ./cmd/llamagate

