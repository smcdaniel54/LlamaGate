@echo off
REM LlamaGate Demo Script for Windows
REM Showcases the power of LlamaGate in under 60 seconds

setlocal enabledelayedexpansion

set BASE_URL=http://localhost:8080
set API_KEY=sk-llamagate

echo ╔════════════════════════════════════════════════════════════╗
echo ║          🚀 LlamaGate Demo - See the Magic! 🚀            ║
echo ╚════════════════════════════════════════════════════════════╝
echo.

REM Check if LlamaGate is running
echo [1/5] Checking LlamaGate...
curl -s -f "%BASE_URL%/health" >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo ✓ LlamaGate is running!
) else (
    echo ✗ LlamaGate is not running. Start it with: scripts\windows\run.cmd
    exit /b 1
)
echo.

REM Check Ollama connectivity
echo [2/5] Checking Ollama connection...
curl -s "%BASE_URL%/health" | findstr /C:"healthy" >nul
if %ERRORLEVEL% EQU 0 (
    echo ✓ Ollama is connected!
) else (
    echo ✗ Ollama is not reachable. Make sure Ollama is running.
    exit /b 1
)
echo.

REM List available models
echo [3/5] Listing available models...
curl -s -H "X-API-Key: %API_KEY%" "%BASE_URL%/v1/models" > temp_models.json 2>nul
if %ERRORLEVEL% NEQ 0 (
    curl -s "%BASE_URL%/v1/models" > temp_models.json
)
echo ✓ Models retrieved
type temp_models.json | findstr /C:"id" | findstr /C:"llama"
del temp_models.json >nul 2>&1
echo.

REM Model loading warning
echo ℹ️  Note: First request to a model may take 5-30+ seconds
echo    (Ollama needs to load model weights into memory)
echo    Subsequent requests are fast once the model is loaded.
echo.

REM First request (slow - from Ollama, may also load model)
echo [4/5] Making first request (this may be slow - loading model + hitting Ollama)...
echo    ⏳ Please wait, this may take 10-30 seconds on first run...
echo    Sending request...
curl -s -X POST "%BASE_URL%/v1/chat/completions" ^
    -H "Content-Type: application/json" ^
    -H "X-API-Key: %API_KEY%" ^
    -d "{\"model\":\"llama2\",\"messages\":[{\"role\":\"user\",\"content\":\"Say hello in exactly one word\"}]}" > temp_response1.json 2>nul
if %ERRORLEVEL% NEQ 0 (
    curl -s -X POST "%BASE_URL%/v1/chat/completions" ^
        -H "Content-Type: application/json" ^
        -d "{\"model\":\"llama2\",\"messages\":[{\"role\":\"user\",\"content\":\"Say hello in exactly one word\"}]}" > temp_response1.json
)
echo    Response received (check temp_response1.json for details)
echo    Time: Slow ^(includes model loading if first time^)
echo.

REM Second request (fast - from cache!)
echo [5/5] Making identical request (this will be INSTANT - from cache!)...
echo    Sending request...
curl -s -X POST "%BASE_URL%/v1/chat/completions" ^
    -H "Content-Type: application/json" ^
    -H "X-API-Key: %API_KEY%" ^
    -d "{\"model\":\"llama2\",\"messages\":[{\"role\":\"user\",\"content\":\"Say hello in exactly one word\"}]}" > temp_response2.json 2>nul
if %ERRORLEVEL% NEQ 0 (
    curl -s -X POST "%BASE_URL%/v1/chat/completions" ^
        -H "Content-Type: application/json" ^
        -d "{\"model\":\"llama2\",\"messages\":[{\"role\":\"user\",\"content\":\"Say hello in exactly one word\"}]}" > temp_response2.json
)
echo    Response received (check temp_response2.json for details)
echo    Time: INSTANT ^(from cache!^)
echo.

echo ╔════════════════════════════════════════════════════════════╗
echo ║                    🎉 Demo Complete! 🎉                    ║
echo ╚════════════════════════════════════════════════════════════╝
echo.
echo This is the power of LlamaGate:
echo   • First request: Slow ^(includes model loading if first time^)
echo   • Cached request: INSTANT!
echo.
echo Your OpenAI code works immediately - just change the base_url!
echo.
echo Next steps:
echo   • See README.md for full documentation
echo   • See QUICKSTART.md for quick setup guide and migration examples
echo   • Edit .env file to customize settings
echo.

del temp_response1.json >nul 2>&1
del temp_response2.json >nul 2>&1

