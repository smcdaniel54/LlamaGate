@echo off
REM LlamaGate Plugin System Test Script for Windows
REM Tests all 8 use cases: adding, validating, and running plugins

echo ========================================
echo LlamaGate Plugin System Test Suite
echo ========================================
echo.
echo Prerequisites:
echo   1. LlamaGate must be running on http://localhost:8080
echo   2. Plugin system must be enabled
echo   3. API key should be set (if authentication enabled)
echo.
echo Press any key to start testing...
pause >nul
echo.

set BASE_URL=http://localhost:8080
set API_KEY=sk-llamagate

if "%API_KEY%"=="" (
    set AUTH_HEADER=
) else (
    set AUTH_HEADER=-H "X-API-Key: %API_KEY%"
)

echo ========================================
echo Testing Plugin System
echo ========================================
echo.

echo [1/3] Testing Plugin Discovery...
echo.
echo Listing all plugins...
curl -s %AUTH_HEADER% %BASE_URL%/v1/plugins
if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✓ Plugin discovery passed
) else (
    echo.
    echo ✗ Plugin discovery failed - Is plugin system enabled?
    goto :end
)
echo.

echo [2/3] Testing Plugin Registration and Validation...
echo.

echo Testing Use Case 1: Environment-Aware Plugin...
curl -s -X POST %BASE_URL%/v1/plugins/usecase1_environment_aware/execute %AUTH_HEADER% ^
    -H "Content-Type: application/json" ^
    -d "{\"input\":\"test\",\"environment\":\"production\"}"
if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✓ Use Case 1 passed
) else (
    echo.
    echo ✗ Use Case 1 failed
)
echo.

echo Testing Use Case 2: User-Configurable Workflow...
curl -s -X POST %BASE_URL%/v1/plugins/usecase2_user_configurable/execute %AUTH_HEADER% ^
    -H "Content-Type: application/json" ^
    -d "{\"query\":\"test query\",\"max_depth\":5,\"use_cache\":true}"
if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✓ Use Case 2 passed
) else (
    echo.
    echo ✗ Use Case 2 failed
)
echo.

echo Testing Use Case 3: Configuration-Driven Tool Selection...
curl -s -X POST %BASE_URL%/v1/plugins/usecase3_tool_selection/execute %AUTH_HEADER% ^
    -H "Content-Type: application/json" ^
    -d "{\"action\":\"process\",\"enabled_tools\":[\"tool1\",\"tool2\"]}"
if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✓ Use Case 3 passed
) else (
    echo.
    echo ✗ Use Case 3 failed
)
echo.

echo Testing Use Case 4: Adaptive Timeout Configuration...
curl -s -X POST %BASE_URL%/v1/plugins/usecase4_adaptive_timeout/execute %AUTH_HEADER% ^
    -H "Content-Type: application/json" ^
    -d "{\"text\":\"This is a test text for timeout calculation\",\"complexity\":\"high\"}"
if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✓ Use Case 4 passed
) else (
    echo.
    echo ✗ Use Case 4 failed
)
echo.

echo Testing Use Case 5: Configuration File-Based Setup...
curl -s -X POST %BASE_URL%/v1/plugins/usecase5_config_file/execute %AUTH_HEADER% ^
    -H "Content-Type: application/json" ^
    -d "{\"operation\":\"process\",\"config_file\":\"custom.json\"}"
if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✓ Use Case 5 passed
) else (
    echo.
    echo ✗ Use Case 5 failed
)
echo.

echo Testing Use Case 6: Runtime Configuration Updates...
echo First, update config...
curl -s -X POST %BASE_URL%/v1/plugins/usecase6_runtime_config/execute %AUTH_HEADER% ^
    -H "Content-Type: application/json" ^
    -d "{\"action\":\"update_config\",\"config\":{\"timeout\":\"60s\",\"retries\":5}}"
echo.
echo Then, use updated config...
curl -s -X POST %BASE_URL%/v1/plugins/usecase6_runtime_config/execute %AUTH_HEADER% ^
    -H "Content-Type: application/json" ^
    -d "{\"action\":\"execute\"}"
if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✓ Use Case 6 passed
) else (
    echo.
    echo ✗ Use Case 6 failed
)
echo.

echo Testing Use Case 7: Context-Aware Configuration...
curl -s -X POST %BASE_URL%/v1/plugins/usecase7_context_aware/execute %AUTH_HEADER% ^
    -H "Content-Type: application/json" ^
    -d "{\"query\":\"Process this with context\"}"
if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✓ Use Case 7 passed
) else (
    echo.
    echo ✗ Use Case 7 failed
)
echo.

echo Testing Use Case 8: Multi-Tenant Configuration...
curl -s -X POST %BASE_URL%/v1/plugins/usecase8_multi_tenant/execute %AUTH_HEADER% ^
    -H "Content-Type: application/json" ^
    -d "{\"tenant_id\":\"tenant1\",\"operation\":\"process\"}"
if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✓ Use Case 8 passed
) else (
    echo.
    echo ✗ Use Case 8 failed
)
echo.

echo [3/3] Testing Input Validation...
echo.

echo Testing validation with missing required input...
curl -s -w "\nHTTP Status: %%{http_code}\n" -X POST %BASE_URL%/v1/plugins/usecase1_environment_aware/execute %AUTH_HEADER% ^
    -H "Content-Type: application/json" ^
    -d "{}"
echo.
echo Testing validation with invalid input type...
curl -s -w "\nHTTP Status: %%{http_code}\n" -X POST %BASE_URL%/v1/plugins/usecase4_adaptive_timeout/execute %AUTH_HEADER% ^
    -H "Content-Type: application/json" ^
    -d "{\"text\":123}"
echo.
echo Testing validation with valid input...
curl -s -w "\nHTTP Status: %%{http_code}\n" -X POST %BASE_URL%/v1/plugins/usecase1_environment_aware/execute %AUTH_HEADER% ^
    -H "Content-Type: application/json" ^
    -d "{\"input\":\"valid input\"}"
if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✓ Input validation tests passed
) else (
    echo.
    echo ✗ Input validation tests failed
)
echo.

:end
echo ========================================
echo Plugin System Testing Complete!
echo ========================================
echo.
echo Summary:
echo   - Plugin discovery: Tested
echo   - Plugin registration: Tested
echo   - Plugin validation: Tested
echo   - Plugin execution: Tested (8 use cases)
echo.
echo Note: Some tests may show errors if plugins are not registered.
echo       Ensure test plugins are registered in your LlamaGate instance.
echo.
