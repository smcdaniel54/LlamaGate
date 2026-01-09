# Plugin System Testing Guide

Complete guide for testing the plugin system, including all 8 use cases.

## Overview

The plugin system includes comprehensive top-to-bottom tests that verify:
- Plugin registration (adding plugins)
- Input validation (validating plugin inputs)
- Plugin execution (running plugins)

## Test Structure

### Test Plugins

Test plugins are defined in `tests/plugins/test_plugins.go`:
- 8 plugins, one for each use case
- Simplified implementations for testing
- Focus on demonstrating use case patterns

### Test Scripts

**Windows:**
- `scripts/windows/test-plugins.cmd` - Comprehensive plugin tests

**Unix/Linux/macOS:**
- `scripts/unix/test-plugins.sh` - Comprehensive plugin tests

## Running Tests

### Prerequisites

1. **LlamaGate Running**
   ```bash
   # Start LlamaGate
   llamagate.exe  # Windows
   ./llamagate    # Unix
   ```

2. **Test Plugins Registered** (Optional)
   - Set `ENABLE_TEST_PLUGINS=true` in `.env`
   - Or register plugins manually in code

### Quick Test

**Windows:**
```cmd
scripts\windows\test-plugins.cmd
```

**Unix/Linux/macOS:**
```bash
chmod +x scripts/unix/test-plugins.sh
./scripts/unix/test-plugins.sh
```

## Test Coverage

### 1. Plugin Discovery

Tests that plugins can be discovered:

```bash
GET /v1/plugins
```

**Expected:**
- Returns list of registered plugins
- Includes metadata for each plugin
- HTTP 200 OK

### 2. Plugin Registration

Tests that plugins can be registered:

```bash
GET /v1/plugins/{plugin_name}
```

**Expected:**
- Returns plugin metadata
- Includes input/output schemas
- HTTP 200 OK

### 3. Input Validation

Tests input validation for each use case:

#### Use Case 1: Environment-Aware
```bash
POST /v1/plugins/usecase1_environment_aware/execute
{
  "input": "test",
  "environment": "production"
}
```

**Validates:**
- Required input "input" present
- Environment value valid
- Returns environment-specific configuration

#### Use Case 2: User-Configurable Workflow
```bash
POST /v1/plugins/usecase2_user_configurable/execute
{
  "query": "test query",
  "max_depth": 5,
  "use_cache": true
}
```

**Validates:**
- Required input "query" present
- max_depth is integer
- use_cache is boolean
- Builds workflow dynamically

#### Use Case 3: Configuration-Driven Tool Selection
```bash
POST /v1/plugins/usecase3_tool_selection/execute
{
  "action": "process",
  "enabled_tools": ["tool1", "tool2"]
}
```

**Validates:**
- Required input "action" present
- enabled_tools is array
- Selects tools based on configuration

#### Use Case 4: Adaptive Timeout
```bash
POST /v1/plugins/usecase4_adaptive_timeout/execute
{
  "text": "Long text...",
  "complexity": "high"
}
```

**Validates:**
- Required input "text" present
- Calculates timeout based on text length
- Adjusts for complexity

#### Use Case 5: Configuration File-Based
```bash
POST /v1/plugins/usecase5_config_file/execute
{
  "operation": "process",
  "config_file": "custom.json"
}
```

**Validates:**
- Required input "operation" present
- Config file parameter accepted
- Returns configuration loaded status

#### Use Case 6: Runtime Configuration Updates
```bash
# First, update config
POST /v1/plugins/usecase6_runtime_config/execute
{
  "action": "update_config",
  "config": {"timeout": "60s", "retries": 5}
}

# Then, use updated config
POST /v1/plugins/usecase6_runtime_config/execute
{
  "action": "execute"
}
```

**Validates:**
- Config update succeeds
- Updated config persists
- Subsequent calls use updated config

#### Use Case 7: Context-Aware Configuration
```bash
POST /v1/plugins/usecase7_context_aware/execute
{
  "query": "Process with context"
}
```

**Validates:**
- Required input "query" present
- Uses context from previous steps
- Adapts workflow based on context

#### Use Case 8: Multi-Tenant Configuration
```bash
POST /v1/plugins/usecase8_multi_tenant/execute
{
  "tenant_id": "tenant1",
  "operation": "process"
}
```

**Validates:**
- Required inputs "tenant_id" and "operation" present
- Loads tenant-specific configuration
- Returns tenant-specific results

### 4. Validation Error Testing

Tests that validation errors are handled correctly:

#### Missing Required Input
```bash
POST /v1/plugins/usecase1_environment_aware/execute
{}
```

**Expected:**
- HTTP 400 Bad Request
- Error message: "required input 'input' is missing"

#### Invalid Input Type
```bash
POST /v1/plugins/usecase4_adaptive_timeout/execute
{
  "text": 123  # Should be string
}
```

**Expected:**
- HTTP 400 Bad Request
- Error message indicating type mismatch

#### Valid Input
```bash
POST /v1/plugins/usecase1_environment_aware/execute
{
  "input": "valid input"
}
```

**Expected:**
- HTTP 200 OK
- Success response with data

## Test Results

### Success Criteria

Each test should:
- ✅ Return appropriate HTTP status code
- ✅ Include structured JSON response
- ✅ Contain execution metadata
- ✅ Handle errors gracefully

### Expected Responses

**Successful Execution:**
```json
{
  "success": true,
  "data": {
    "result": "..."
  },
  "metadata": {
    "execution_time": "10ms",
    "steps_executed": 1,
    "timestamp": "2026-01-07T12:00:00Z"
  }
}
```

**Validation Error:**
```json
{
  "error": {
    "type": "invalid_request_error",
    "message": "required input 'input' is missing"
  }
}
```

## Manual Testing

### Test Individual Use Case

```bash
# Use Case 1: Environment-Aware
curl -X POST http://localhost:8080/v1/plugins/usecase1_environment_aware/execute \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{"input":"test","environment":"production"}'
```

### Test Validation

```bash
# Missing required input (should fail)
curl -X POST http://localhost:8080/v1/plugins/usecase1_environment_aware/execute \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{}'
```

### Test Discovery

```bash
# List all plugins
curl http://localhost:8080/v1/plugins \
  -H "X-API-Key: sk-llamagate"

# Get specific plugin
curl http://localhost:8080/v1/plugins/usecase1_environment_aware \
  -H "X-API-Key: sk-llamagate"
```

## Integration with Main Test Suite

The main test suite (`test.cmd` / `test.sh`) now includes:
- [8/9] Plugin System Discovery
- [9/9] Plugin Use Cases Overview

For comprehensive plugin testing, use:
- `test-plugins.cmd` / `test-plugins.sh`

## Troubleshooting

### "Plugin not found"

**Issue:** Plugin not registered
**Solution:** 
- Set `ENABLE_TEST_PLUGINS=true` in `.env`
- Or register plugins manually in code
- Check plugin registry is initialized

### "Plugin system not available"

**Issue:** Plugin system not enabled
**Solution:**
- Ensure plugin API endpoints are registered
- Check `cmd/llamagate/main.go` includes plugin routes

### "Validation failed"

**Issue:** Input validation error
**Solution:**
- Check required inputs are provided
- Verify input types match schema
- Review error message for details

## Summary

The test suite provides:
- ✅ Complete coverage of all 8 use cases
- ✅ Registration testing
- ✅ Validation testing
- ✅ Execution testing
- ✅ Error handling verification

All tests are automated and can be run via the test scripts.
