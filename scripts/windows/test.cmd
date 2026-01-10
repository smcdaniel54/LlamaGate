@echo off
REM LlamaGate Test Script for Windows
REM This script tests all LlamaGate endpoints

echo ========================================
echo LlamaGate Test Suite
echo ========================================
echo.
echo Prerequisites:
echo   1. Ollama must be running on http://localhost:11434
echo   2. LlamaGate must be running on http://localhost:11435
echo   3. At least one model should be available in Ollama (e.g., llama2)
echo.
echo Press any key to start testing...
pause >nul
echo.

set BASE_URL=http://localhost:11435
set API_KEY=sk-llamagate

echo [1/9] Testing Health Check...
curl -s %BASE_URL%/health
if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✓ Health check passed
) else (
    echo.
    echo ✗ Health check failed - Is LlamaGate running?
    goto :end
)
echo.

echo [2/9] Testing Models Endpoint...
if "%API_KEY%"=="" (
    curl -s %BASE_URL%/v1/models
) else (
    curl -s -H "X-API-Key: %API_KEY%" %BASE_URL%/v1/models
)
if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✓ Models endpoint passed
) else (
    echo.
    echo ✗ Models endpoint failed
)
echo.

echo [3/9] Testing Chat Completions (Non-Streaming)...
if "%API_KEY%"=="" (
    curl -s -X POST %BASE_URL%/v1/chat/completions ^
        -H "Content-Type: application/json" ^
        -d "{\"model\":\"llama2\",\"messages\":[{\"role\":\"user\",\"content\":\"Say hello in one word\"}]}"
) else (
    curl -s -X POST %BASE_URL%/v1/chat/completions ^
        -H "Content-Type: application/json" ^
        -H "X-API-Key: %API_KEY%" ^
        -d "{\"model\":\"llama2\",\"messages\":[{\"role\":\"user\",\"content\":\"Say hello in one word\"}]}"
)
if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✓ Chat completions (non-streaming) passed
) else (
    echo.
    echo ✗ Chat completions failed
)
echo.

echo [4/9] Testing Caching (Same Request Twice)...
echo First request (should be slow):
if "%API_KEY%"=="" (
    curl -s -w "\nTime: %%{time_total}s\n" -X POST %BASE_URL%/v1/chat/completions ^
        -H "Content-Type: application/json" ^
        -d "{\"model\":\"llama2\",\"messages\":[{\"role\":\"user\",\"content\":\"What is 2+2?\"}]}"
) else (
    curl -s -w "\nTime: %%{time_total}s\n" -X POST %BASE_URL%/v1/chat/completions ^
        -H "Content-Type: application/json" ^
        -H "X-API-Key: %API_KEY%" ^
        -d "{\"model\":\"llama2\",\"messages\":[{\"role\":\"user\",\"content\":\"What is 2+2?\"}]}"
)
echo.
echo Second request (should be fast - cached):
if "%API_KEY%"=="" (
    curl -s -w "\nTime: %%{time_total}s\n" -X POST %BASE_URL%/v1/chat/completions ^
        -H "Content-Type: application/json" ^
        -d "{\"model\":\"llama2\",\"messages\":[{\"role\":\"user\",\"content\":\"What is 2+2?\"}]}"
) else (
    curl -s -w "\nTime: %%{time_total}s\n" -X POST %BASE_URL%/v1/chat/completions ^
        -H "Content-Type: application/json" ^
        -H "X-API-Key: %API_KEY%" ^
        -d "{\"model\":\"llama2\",\"messages\":[{\"role\":\"user\",\"content\":\"What is 2+2?\"}]}"
)
echo.
echo ✓ Cache test completed (check times above - second should be much faster)
echo.

echo [5/9] Testing Authentication (if enabled)...
if "%API_KEY%"=="" (
    echo Authentication is disabled, skipping auth test
) else (
    echo Testing with invalid API key (should fail)...
    curl -s -w "\nHTTP Status: %%{http_code}\n" -X GET %BASE_URL%/v1/models ^
        -H "X-API-Key: invalid-key"
    echo.
    echo Testing with valid API key (should succeed)...
    curl -s -w "\nHTTP Status: %%{http_code}\n" -X GET %BASE_URL%/v1/models ^
        -H "X-API-Key: %API_KEY%"
    echo.
    echo ✓ Authentication test completed
)
echo.

echo [6/9] Testing MCP API Endpoints (if MCP enabled)...
if "%API_KEY%"=="" (
    set AUTH_HEADER=
) else (
    set AUTH_HEADER=-H "X-API-Key: %API_KEY%"
)
echo Testing MCP servers list...
curl -s %AUTH_HEADER% %BASE_URL%/v1/mcp/servers
if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✓ MCP API endpoints are accessible
    echo   Note: If you see "MCP is not enabled", configure MCP in your config file
) else (
    echo.
    echo ℹ MCP API test skipped (MCP may not be enabled)
)
echo.

echo [7/9] Testing MCP URI Scheme (if MCP enabled)...
if "%API_KEY%"=="" (
    set AUTH_HEADER=
) else (
    set AUTH_HEADER=-H "X-API-Key: %API_KEY%"
)
echo Testing chat completion with MCP URI...
echo Note: This requires an MCP server with resources configured
curl -s -X POST %BASE_URL%/v1/chat/completions %AUTH_HEADER% ^
    -H "Content-Type: application/json" ^
    -d "{\"model\":\"llama2\",\"messages\":[{\"role\":\"user\",\"content\":\"Test mcp://test-server/resource\"}]}" >nul
if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✓ MCP URI scheme test completed
    echo   Note: If MCP is not enabled or server not found, request will continue without resource context
) else (
    echo.
    echo ℹ MCP URI test skipped (MCP may not be enabled or server not configured)
)
echo.

echo [8/9] Testing Plugin System (if enabled)...
if "%API_KEY%"=="" (
    set AUTH_HEADER=
) else (
    set AUTH_HEADER=-H "X-API-Key: %API_KEY%"
)
echo Testing plugin discovery...
curl -s %AUTH_HEADER% %BASE_URL%/v1/plugins >nul
if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✓ Plugin system is accessible
    echo   Run scripts\windows\test-plugins.cmd for comprehensive plugin tests
) else (
    echo.
    echo ℹ Plugin system test skipped (Plugin system may not be enabled)
)
echo.

echo [9/9] Testing Plugin Use Cases (if plugins registered)...
echo Note: This requires test plugins to be registered
echo       See scripts\windows\test-plugins.cmd for full plugin testing
echo       Or set ENABLE_TEST_PLUGINS=true to enable test plugins
echo.

:end
echo ========================================
echo Testing Complete!
echo ========================================
echo.
echo Check the log file if LOG_FILE is set in your .env
echo Check console output for request logs
echo.

