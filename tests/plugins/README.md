# Test Plugins for Use Case Testing

This directory contains test plugins that implement all 8 dynamic configuration use cases.

## Test Plugins

Each use case has a corresponding test plugin:

1. **UseCase1Plugin** (`usecase1_environment_aware`)
   - Environment-aware plugin configuration
   - Adapts behavior based on environment variable

2. **UseCase2Plugin** (`usecase2_user_configurable`)
   - User-configurable workflow parameters
   - Dynamic workflow building based on input

3. **UseCase3Plugin** (`usecase3_tool_selection`)
   - Configuration-driven tool selection
   - Selects tools based on configuration

4. **UseCase4Plugin** (`usecase4_adaptive_timeout`)
   - Adaptive timeout configuration
   - Calculates timeout based on input complexity

5. **UseCase5Plugin** (`usecase5_config_file`)
   - Configuration file-based plugin setup
   - Loads configuration from files

6. **UseCase6Plugin** (`usecase6_runtime_config`)
   - Runtime configuration updates
   - Supports updating configuration at runtime

7. **UseCase7Plugin** (`usecase7_context_aware`)
   - Context-aware configuration
   - Uses context from previous steps

8. **UseCase8Plugin** (`usecase8_multi_tenant`)
   - Multi-tenant configuration
   - Supports tenant-specific configurations

## Usage

### Registering Test Plugins

To use these test plugins, you need to register them in your application:

```go
import (
    "github.com/llamagate/llamagate/internal/plugins"
    testplugins "github.com/llamagate/llamagate/tests/plugins"
)

func main() {
    registry := plugins.NewRegistry()
    
    // Register all test plugins
    for _, plugin := range testplugins.CreateTestPlugins() {
        if err := registry.Register(plugin); err != nil {
            log.Error().Err(err).Msg("Failed to register test plugin")
        }
    }
}
```

### Running Tests

Use the test scripts to test all use cases:

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

The test scripts verify:

1. **Plugin Discovery**
   - List all plugins
   - Get plugin metadata

2. **Input Validation**
   - Required inputs
   - Type validation
   - Error handling

3. **Plugin Execution**
   - All 8 use cases
   - Proper response format
   - Execution metadata

## Notes

- Test plugins are simplified implementations for testing
- They may not include full workflow execution (LLM/tool calls)
- They focus on demonstrating the use case pattern
- For production plugins, see `plugins/examples/`
