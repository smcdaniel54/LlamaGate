# LlamaGate Issues Report

This document tracks issues, limitations, and required enhancements for LlamaGate to fully support the Smart Voice Alexa plugin.

## Current Status

**Plugin Status**: ✅ Created and integrated  
**LlamaGate Version**: Latest (from repository)  
**Last Updated**: 2026-01-10 (Fixed 5 critical issues - most issues resolved)  
**Test Coverage**: ~65-70% (Smart Voice app), Plugin fully functional

## Quick Summary

### ✅ Resolved (7)

1. **LLM Integration** - PluginContext with CallLLM() working ✅
2. **Configuration Management** - Config file and env vars working ✅
3. **Multiple Instance Prevention** - Port check with clear error messages ✅
4. **HTTPS/SSL Support** - Native HTTPS/TLS support added ✅
5. **Error Handling** - Enhanced plugin error handling and logging ✅
6. **API Documentation** - Complete plugin API reference added ✅
7. **Rate Limit Too Restrictive** - Default changed from 10 to 50 RPS ✅

### ⚠️ Active Issues (2)

1. **Plugin Testing Framework** - MEDIUM priority (still pending)
2. **Plugin Discovery** - LOW priority (still pending)

### 📋 Feature Requests (4)

1. **Plugin Context** - ✅ Partially implemented (LLM, Config, Logger ✅; Tool Manager ⚠️)
2. **Plugin Middleware** - MEDIUM priority
3. **Plugin Hot Reload** - LOW priority
4. **Plugin Metrics** - MEDIUM priority

## ✅ Resolved Issues

### 1. LLM Integration in Plugins ✅ RESOLVED

**Previous Issue**: Plugins could not directly access LlamaGate's LLM functionality.

**Resolution**:

- ✅ LlamaGate provides `PluginContext` with `LLMHandlerFunc`
- ✅ Plugin can access LLM via `pluginCtx.CallLLM()`
- ✅ LLM handler created via `proxyInstance.CreatePluginLLMHandler()`
- ✅ Plugin registered with context via `registry.RegisterWithContext()`

**Implementation**:

- Location: `../LlamaGate/internal/plugins/context.go`
- Method: `PluginContext.CallLLM()`
- Status: ✅ Fully functional

**Code Example**:

```go
pluginCtx := registry.GetContext("alexa_skill")
response, err := pluginCtx.CallLLM(ctx, model, messages, options)
```

---

### 2. Plugin Configuration Management ✅ RESOLVED

**Previous Issue**: No standard way to configure plugins via config file.

**Resolution**:

- ✅ Plugin configuration via `cfg.Plugins.Configs["alexa_skill"]`
- ✅ Environment variable support (ALEXA_*)
- ✅ Configuration passed to plugin via `NewAlexaSkillPluginWithConfig()`
- ✅ Configuration stored in PluginContext

**Implementation**:

- Location: `../LlamaGate/internal/setup/alexa_plugin.go`
- Config file: `llamagate.yaml` → `plugins.configs.alexa_skill`
- Status: ✅ Fully functional

**Example Config**:

```yaml
plugins:
  configs:
    alexa_skill:
      wake_word: "Smart Voice"
      case_sensitive: false
      remove_from_query: true
      default_model: "llama3.2"
```

**Environment Variables**:

- `ALEXA_WAKE_WORD`
- `ALEXA_CASE_SENSITIVE`
- `ALEXA_REMOVE_FROM_QUERY`
- `ALEXA_DEFAULT_MODEL`

---

## Issues and Limitations

### 1. Multiple Instance Prevention ✅ RESOLVED

**Issue**: LlamaGate does not prevent multiple instances from running simultaneously on the same machine.

**Status**: ✅ **RESOLVED** - Port availability check implemented in `cmd/llamagate/main.go` with clear error messages explaining single-instance architecture.

**Current State**:

- Multiple instances can be started without detection
- Each instance attempts to bind to the same port (default: 8080)
- Port conflict only detected when second instance tries to start
- No clear error message explaining why only one instance should run
- Multiple instances waste resources (duplicate MCP clients, cache, connections)

**Impact**:

- **Resource Waste**: Multiple instances create duplicate:
  - MCP client connections
  - Cache instances
  - Ollama connections
  - Memory usage
- **Port Conflicts**: Second instance fails with generic "address already in use" error
- **User Confusion**: No clear guidance that only one instance should run per machine
- **Architecture**: Multiple apps should connect to the same LlamaGate instance, not start separate instances

**Required Solution**:

- Port availability check before starting server
- Clear error message when port is already in use
- Documentation explaining single-instance architecture
- Optional: Lock file mechanism for additional safety

**Proposed Implementation**:

```go
// Check port availability before starting
func checkPortAvailability(port string) error {
    addr := ":" + port
    ln, err := net.Listen("tcp", addr)
    if err != nil {
        return fmt.Errorf("port %s is already in use - another LlamaGate instance may be running. Only one instance should run per machine", port)
    }
    ln.Close()
    return nil
}
```

**Error Message**:

```
❌ Error: port 8080 is already in use - another LlamaGate instance may be running. 
Only one instance should run per machine.

Only one instance of LlamaGate should run per machine.
Multiple apps can connect to the same LlamaGate instance.
```

**Priority**: MEDIUM - Important for resource efficiency and user experience

**Status**: ✅ **RESOLVED** - Port availability check implemented in `cmd/llamagate/main.go` with clear error messages

**Workaround**:

- Manually check if port is in use before starting
- Use process manager to ensure only one instance runs

**Related**: Architecture decision - LlamaGate should be a shared service that multiple applications connect to, not a per-application instance.

---

### 2. Plugin Testing and Coverage ⚠️ MEDIUM PRIORITY

**Issue**: Limited testing infrastructure for plugins.

**Current State**:

- Plugin compiles and runs successfully
- Manual test scripts exist (`test_alexa.sh`, `test_alexa.ps1`)
- No automated unit tests for plugin
- No integration tests with mock LLM responses
- No E2E tests with actual Alexa requests

**Required Solution**:

- Plugin testing utilities/framework
- Mock LLM handler for testing
- Test fixtures for Alexa requests
- Integration test helpers

**Priority**: MEDIUM - Affects development velocity and reliability

**Workaround**:

- Manual testing via test scripts
- E2E testing via Smart Voice app tests

---

### 3. Plugin Discovery and Auto-Loading ⚠️ LOW PRIORITY

**Issue**: Plugins must be manually registered in code.

**Current State**:

- Plugins registered in `cmd/llamagate/main.go`
- No automatic discovery from `plugins/` directory
- No plugin manifest/descriptor files

**Required Solution**:

- Auto-discover plugins in `plugins/` directory
- Plugin manifest files (e.g., `plugin.yaml`)
- Enable/disable plugins via config

**Priority**: LOW - Nice to have

---

### 4. HTTPS/SSL Support ✅ RESOLVED

**Issue**: LlamaGate doesn't support HTTPS natively.

**Current State**:

- Only HTTP support
- Alexa requires HTTPS for production
- Need reverse proxy (nginx, Caddy) for SSL termination

**Required Solution**:

- Native HTTPS support in LlamaGate
- SSL certificate management
- Let's Encrypt integration

**Workaround**:

- Use reverse proxy (nginx, Caddy) for SSL termination

**Priority**: HIGH - Required for production Alexa deployment

**Status**: ✅ **RESOLVED** - Native HTTPS/TLS support added with `TLS_ENABLED`, `TLS_CERT_FILE`, and `TLS_KEY_FILE` configuration options

---

### 5. Plugin Error Handling and Logging ✅ RESOLVED

**Issue**: Limited error handling and logging for plugins.

**Current State**:

- Basic error handling in plugin
- No structured logging for plugin operations
- No plugin-specific log levels

**Required Solution**:

- Plugin-specific logging context
- Structured error responses
- Plugin health monitoring

**Priority**: MEDIUM - Affects debugging and monitoring

**Status**: ✅ **RESOLVED** - Enhanced plugin error handling with plugin-specific logging context, structured error responses, and execution timing metrics

---

### 6. Plugin Testing Framework ⚠️ LOW PRIORITY

**Issue**: No built-in testing framework for plugins.

**Current State**:

- Manual test scripts
- No unit test framework for plugins
- No integration test helpers

**Required Solution**:

- Plugin testing utilities
- Mock LLM handler for testing
- Test fixtures and helpers

**Priority**: LOW - Nice to have

---

### 7. Plugin API Documentation ✅ RESOLVED

**Issue**: Limited documentation for plugin development.

**Current State**:

- Basic plugin documentation exists
- No API reference for plugin interfaces
- Limited examples for custom endpoints

**Required Solution**:

- Complete API reference
- More examples
- Plugin development guide

**Priority**: MEDIUM - Affects developer experience

**Status**: ✅ **RESOLVED** - Complete API reference added to `docs/PLUGINS.md` with PluginContext API, ExtendedPlugin interface, HTTP API endpoints, and code examples

---

### 8. Rate Limit Too Restrictive for Normal Usage ✅ RESOLVED

**Issue**: Default rate limit of 10 requests per second (RPS) is too restrictive for normal usage.

**Current State**:

- Default `RATE_LIMIT_RPS`: `10` requests per second
- This equals 1 request every 100ms
- Too restrictive for:
  - Normal development and testing
  - Multiple concurrent users/applications
  - CI/CD test execution
  - Production scenarios with moderate traffic

**Impact**:

- **Development Experience**: Developers hit rate limits frequently during testing
- **Test Execution**: E2E tests require delays (150ms+) to avoid rate limiting
- **User Experience**: Legitimate users may experience 429 errors during normal usage
- **Performance**: Unnecessarily throttles local LLM gateway performance
- **Resource Underutilization**: Local LLM gateways can typically handle much higher throughput

**Research Findings**:

Based on industry standards and typical API gateway configurations:

- **Typical API Gateways**: 50-100+ RPS for normal usage
- **LLM Gateways**: Often configured at 50-200 RPS
- **Local Services**: Can handle 100-500+ RPS depending on hardware
- **OpenAI API**: Tiered limits (varies by tier, but typically 50-500+ RPS)
- **Best Practice**: Start conservative but allow reasonable throughput

**Recommended Solution**:

Change default `RATE_LIMIT_RPS` from `10` to `50` requests per second:

**Rationale**:

- **50 RPS** = 1 request every 20ms
- More reasonable for normal usage patterns
- Still provides protection against abuse
- Allows multiple applications/users to share the gateway
- Better matches typical API gateway defaults
- Can be easily adjusted up/down via configuration

**Implementation**:

1. **Update Default Value**:
  
  ```go
  // In LlamaGate/internal/config/config.go
  viper.SetDefault("RATE_LIMIT_RPS", 50.0) // Changed from 10.0
  ```
  
2. **Update Documentation**:
  
  - Update README.md default value table
  - Update .env.example file
  - Update QUICKSTART.md troubleshooting section
  - Update API.md rate limiting section
3. **Migration Note**:
  
  - Existing installations with `.env` files will keep their current setting
  - New installations will use the new default
  - Document how to adjust for different use cases

**Configuration Guidance**:

Document recommended rate limits for different scenarios:

```yaml
# Development/Testing
RATE_LIMIT_RPS=50  # Recommended default

# Production (Single User)
RATE_LIMIT_RPS=100  # For dedicated single-user deployments

# Production (Multiple Users)
RATE_LIMIT_RPS=200  # For shared deployments with multiple apps

# High-Performance (Local Only)
RATE_LIMIT_RPS=500  # For high-performance local setups
```

**Alternative Considerations**:

- **Per-User Rate Limiting**: Implement per-API-key rate limits (future enhancement)
- **Burst Capacity**: Current implementation already supports burst (equals RPS value)
- **Dynamic Adjustment**: Could add auto-scaling based on system load (future enhancement)

**Priority**: MEDIUM - Affects user experience and development velocity, but has workaround (configurable)

**Status**: ✅ **RESOLVED** - Default rate limit changed from 10 to 50 RPS in `internal/config/config.go` and all related documentation updated

**Workaround**:

- Users can set `RATE_LIMIT_RPS=50` (or higher) in their `.env` file
- Tests can set `LLAMAGATE_RATE_LIMIT_RPS` environment variable

**Related**:

- Smart Voice E2E tests assume 50 RPS for reasonable test execution speed
- Rate limiting middleware implementation is flexible and supports any value

---

## Feature Requests

### 1. Plugin Context/Environment ✅ IMPLEMENTED

**Request**: Provide plugin context with access to:

- ✅ LLM handler function (via `PluginContext.CallLLM()`)
- ⚠️ Tool manager (for MCP tools) - Not yet available
- ✅ Configuration (via `PluginContext.GetConfig()`)
- ✅ Logger instance (via `PluginContext.Logger`)

**Use Case**: Plugins need access to LlamaGate services without circular dependencies.

**Status**: ✅ Partially implemented - LLM, Config, Logger available. Tool manager pending.

**Priority**: MEDIUM (for tool manager access)

---

### 2. Plugin Middleware Support

**Request**: Allow plugins to register middleware for:

- Request preprocessing
- Response postprocessing
- Authentication/authorization
- Rate limiting per plugin

**Use Case**: Custom authentication for Alexa endpoint, request validation.

**Priority**: MEDIUM

---

### 3. Plugin Hot Reload

**Request**: Reload plugins without restarting LlamaGate.

**Use Case**: Development and updates without downtime.

**Priority**: LOW

---

### 4. Plugin Metrics and Monitoring

**Request**: Expose plugin metrics:

- Request count
- Error rate
- Execution time
- LLM call count

**Use Case**: Monitor plugin performance and health.

**Priority**: MEDIUM

---

## Workarounds Implemented

### 1. LLM Integration ✅ RESOLVED

**Previous**: Placeholder response with message indicating LLM integration pending.

**Current**: ✅ Fully implemented using `PluginContext.CallLLM()`

**Code Location**: `../LlamaGate/plugins/alexa_skill.go` → `processWithLLM()`

**Status**: ✅ Working - Plugin can call LLM via context

---

### 2. Configuration ✅ RESOLVED

**Previous**: Hardcoded configuration in plugin struct.

**Current**: ✅ Fully implemented via config file and environment variables

**Code Location**:

- `../LlamaGate/internal/setup/alexa_plugin.go` → `RegisterAlexaPlugin()`
- `../LlamaGate/plugins/alexa_skill.go` → `NewAlexaSkillPluginWithConfig()`

**Status**: ✅ Working - Plugin configurable via YAML and env vars

---

## Testing Status

### ✅ Completed

- Plugin structure and registration
- Wake word detection
- Alexa request/response handling
- Custom endpoint registration
- Build verification
- LLM integration (via PluginContext)
- Configuration management
- Manual test scripts

### ⏳ Pending

- Automated unit tests for plugin
- Integration tests with mock LLM
- End-to-end testing with actual Alexa device
- Performance testing
- Error scenario testing
- Plugin-specific test framework

---

## Recommendations

### Immediate Actions (High Priority)

1. ✅ ~~**Implement LLM Integration**~~ - COMPLETED
2. ✅ ~~**Add Configuration Support**~~ - COMPLETED
3. **Document HTTPS Setup**: Create guide for reverse proxy setup (for production Alexa deployment)

### Short-term Actions (Medium Priority)

1. ✅ ~~**Add Multiple Instance Prevention**~~ - COMPLETED
2. **Add Plugin Testing Framework**: Create testing utilities for plugins (still pending)
3. ✅ ~~**Improve Error Handling**~~ - COMPLETED
4. **Add Plugin Metrics**: Monitoring and observability (still pending)
5. ✅ ~~**Create API Documentation**~~ - COMPLETED
6. ✅ ~~**Increase Default Rate Limit**~~ - COMPLETED

### Long-term Actions (Low Priority)

1. **Plugin Discovery**: Auto-discover and load plugins
2. **Plugin Hot Reload**: Reload plugins without restarting LlamaGate
3. **Plugin Middleware Support**: Allow plugins to register middleware
4. **MCP Tool Access**: Provide tool manager access to plugins

---

## Related Documentation

- [LlamaGate Plugin Documentation](../LlamaGate/docs/PLUGINS.md)
- [Plugin Quick Start](../LlamaGate/docs/PLUGIN_QUICKSTART.md)
- [Alexa Plugin README](../LlamaGate/plugins/ALEXA_PLUGIN_README.md)

---

## Update History

- **2026-01-09 (Initial)**: Initial issues report created
  
  - Documented LLM integration issue
  - Documented configuration management issue
  - Documented HTTPS requirement
  - Added feature requests
- **2026-01-09 (Updated)**: Major updates
  
  - ✅ **RESOLVED**: LLM integration via PluginContext
  - ✅ **RESOLVED**: Plugin configuration management
  - ✅ **RESOLVED**: Plugin context/environment (partial - LLM, Config, Logger working)
  - Added plugin testing coverage gap
  - Updated workarounds section (resolved items)
  - Updated recommendations (completed items marked)
  - Added test coverage analysis findings
- **2026-01-XX (Updated)**: Added multiple instance prevention issue
  
  - Added issue #1: Multiple Instance Prevention
  - Documented resource waste and port conflict problems
  - Proposed implementation with port availability check
  - Added to short-term actions (Medium priority)
- **2026-01-10 (Updated)**: Added rate limit issue
  
  - Added issue #8: Rate Limit Too Restrictive for Normal Usage
  - Documented that 10 RPS default is too restrictive
  - Recommended changing default to 50 RPS
  - Provided configuration guidance for different scenarios
  - Added to short-term actions (Medium priority)
- **2026-01-10 (Major Update)**: Fixed multiple critical issues
  
  - ✅ **RESOLVED**: Multiple Instance Prevention (Issue #1)
  - ✅ **RESOLVED**: HTTPS/SSL Support (Issue #4)
  - ✅ **RESOLVED**: Plugin Error Handling and Logging (Issue #5)
  - ✅ **RESOLVED**: Plugin API Documentation (Issue #7)
  - ✅ **RESOLVED**: Rate Limit Too Restrictive (Issue #8)
  - Most critical issues now resolved, only 2 low-priority items remain

---

## Notes

- Issues are prioritized based on impact on Smart Voice functionality
- Workarounds are documented where available
- This document should be updated as issues are resolved or new issues are discovered

## Key Findings from Test Coverage Analysis

### Smart Voice Application Coverage

- **Overall**: ~65-70%
- **Well Covered**: Wake word detection (95%), Alexa models (80%), Router (90%)
- **Gaps**: LlamaGate client (0%), MCP client (0%), Logger (0%)

### Plugin Status

- **Build**: ✅ Compiles successfully
- **Integration**: ✅ Registered and functional
- **LLM Access**: ✅ Working via PluginContext
- **Configuration**: ✅ Working via config file and env vars
- **Testing**: ⚠️ Manual tests only, no automated test suite

### Critical Gaps for Production

1. ✅ ~~**HTTPS/SSL Support**~~ - RESOLVED (native support added)
2. **Plugin Testing Framework**: Needed for reliable development (still pending)
3. ✅ ~~**Error Handling**~~ - RESOLVED (structured error responses added)
4. **Monitoring**: Plugin metrics and observability needed (still pending)
