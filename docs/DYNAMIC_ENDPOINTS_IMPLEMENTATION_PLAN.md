# LlamaGate Dynamic Endpoints Implementation Plan

## Overview

This plan outlines the implementation of dynamic endpoint support in LlamaGate, allowing extensions to define custom HTTP endpoints that are automatically registered with the router.

## Goal

Enable extensions to define custom HTTP endpoints in their manifests, which LlamaGate will automatically register and route to the extension's workflow execution.

## Current State

### What Exists
- ✅ Extension manifest system (`internal/extensions/manifest.go`)
- ✅ Extension registry (`internal/extensions/registry.go`)
- ✅ Extension handler (`internal/extensions/handler.go`)
- ✅ Workflow executor (`internal/extensions/workflow.go`)
- ✅ Router setup in `main.go` (Gin framework)

### What's Missing
- ❌ `endpoints:` field in manifest struct
- ❌ Endpoint definition types
- ❌ Route registration infrastructure
- ❌ Dynamic route handlers
- ❌ Endpoint validation

## Specification Reference

Based on `docs/EXTENSIONS_SPEC_V0.9.1.md`:

```yaml
endpoints:
  - path: string                # Relative path (e.g., "/custom/action")
    method: string              # HTTP method (GET, POST, PUT, DELETE, etc.)
    description: string
    request_schema: {...}       # Optional: JSON Schema
    response_schema: {...}      # Optional: JSON Schema
    requires_auth: boolean      # Default: true
    requires_rate_limit: boolean # Default: true
```

## Implementation Phases

### Phase 1: Manifest Structure Updates

**Goal**: Add endpoint definitions to manifest struct

**Files to Modify**:
- `internal/extensions/manifest.go`

**Changes**:

1. **Add EndpointDefinition struct**:
   ```go
   // EndpointDefinition defines a custom HTTP endpoint for an extension
   type EndpointDefinition struct {
       Path            string                 `yaml:"path"`
       Method          string                 `yaml:"method"`
       Description     string                 `yaml:"description"`
       RequestSchema   map[string]interface{} `yaml:"request_schema,omitempty"`
       ResponseSchema  map[string]interface{} `yaml:"response_schema,omitempty"`
       RequiresAuth    *bool                  `yaml:"requires_auth,omitempty"`
       RequiresRateLimit *bool                `yaml:"requires_rate_limit,omitempty"`
   }
   ```

2. **Add Endpoints field to Manifest**:
   ```go
   type Manifest struct {
       // ... existing fields ...
       Endpoints []EndpointDefinition `yaml:"endpoints,omitempty"`
   }
   ```

3. **Add validation**:
   ```go
   func (m *Manifest) ValidateEndpoints() error {
       for i, endpoint := range m.Endpoints {
           // Validate path
           if endpoint.Path == "" {
               return fmt.Errorf("endpoint %d: path is required", i)
           }
           if !strings.HasPrefix(endpoint.Path, "/") {
               return fmt.Errorf("endpoint %d: path must start with '/'", i)
           }
           
           // Validate method
           validMethods := map[string]bool{
               "GET": true, "POST": true, "PUT": true,
               "DELETE": true, "PATCH": true, "HEAD": true, "OPTIONS": true,
           }
           if !validMethods[strings.ToUpper(endpoint.Method)] {
               return fmt.Errorf("endpoint %d: invalid method '%s'", i, endpoint.Method)
           }
       }
       return nil
   }
   ```

**Testing**:
- Unit tests for endpoint validation
- Test manifest loading with endpoints
- Test invalid endpoint definitions

---

### Phase 2: Route Manager

**Goal**: Create infrastructure for dynamic route registration

**New File**: `internal/extensions/route_manager.go`

**Implementation**:

```go
package extensions

import (
    "fmt"
    "net/http"
    "strings"
    
    "github.com/gin-gonic/gin"
)

// RouteManager manages dynamic routes for extensions
type RouteManager struct {
    router   *gin.Engine
    registry *Registry
    executor *WorkflowExecutor
    routes   map[string]*RouteInfo // Track registered routes
}

// RouteInfo tracks registered route information
type RouteInfo struct {
    ExtensionName string
    Endpoint      EndpointDefinition
    Handler       gin.HandlerFunc
}

// NewRouteManager creates a new route manager
func NewRouteManager(router *gin.Engine, registry *Registry, executor *WorkflowExecutor) *RouteManager {
    return &RouteManager{
        router:   router,
        registry: registry,
        executor: executor,
        routes:   make(map[string]*RouteInfo),
    }
}

// RegisterExtensionRoutes registers all endpoints for an extension
func (rm *RouteManager) RegisterExtensionRoutes(manifest *Manifest) error {
    if len(manifest.Endpoints) == 0 {
        return nil // No endpoints to register
    }
    
    for _, endpoint := range manifest.Endpoints {
        if err := rm.registerRoute(manifest, endpoint); err != nil {
            return fmt.Errorf("failed to register route for extension %s: %w", manifest.Name, err)
        }
    }
    
    return nil
}

// registerRoute registers a single endpoint route
func (rm *RouteManager) registerRoute(manifest *Manifest, endpoint EndpointDefinition) error {
    // Build full path: /v1/extensions/{name}{endpoint.path}
    fullPath := fmt.Sprintf("/v1/extensions/%s%s", manifest.Name, endpoint.Path)
    
    // Normalize path (remove trailing slashes, etc.)
    fullPath = normalizePath(fullPath)
    
    // Create route key for tracking
    routeKey := fmt.Sprintf("%s:%s", strings.ToUpper(endpoint.Method), fullPath)
    
    // Check for conflicts
    if existing, exists := rm.routes[routeKey]; exists {
        return fmt.Errorf("route conflict: %s already registered by extension %s", routeKey, existing.ExtensionName)
    }
    
    // Create handler
    handler := rm.createEndpointHandler(manifest, endpoint)
    
    // Register with router based on method
    method := strings.ToUpper(endpoint.Method)
    switch method {
    case "GET":
        rm.router.GET(fullPath, handler)
    case "POST":
        rm.router.POST(fullPath, handler)
    case "PUT":
        rm.router.PUT(fullPath, handler)
    case "DELETE":
        rm.router.DELETE(fullPath, handler)
    case "PATCH":
        rm.router.PATCH(fullPath, handler)
    case "HEAD":
        rm.router.HEAD(fullPath, handler)
    case "OPTIONS":
        rm.router.OPTIONS(fullPath, handler)
    default:
        return fmt.Errorf("unsupported HTTP method: %s", method)
    }
    
    // Track route
    rm.routes[routeKey] = &RouteInfo{
        ExtensionName: manifest.Name,
        Endpoint:      endpoint,
        Handler:      handler,
    }
    
    return nil
}

// createEndpointHandler creates a Gin handler for an extension endpoint
func (rm *RouteManager) createEndpointHandler(manifest *Manifest, endpoint EndpointDefinition) gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := middleware.GetRequestID(c)
        
        // Check if extension is enabled
        if !rm.registry.IsEnabled(manifest.Name) {
            response.ServiceUnavailable(c, "Extension is disabled", requestID)
            return
        }
        
        // Only workflow extensions can have custom endpoints
        if manifest.Type != "workflow" {
            response.BadRequest(c, "Only workflow extensions can define custom endpoints", requestID)
            return
        }
        
        // Parse request body (for POST, PUT, PATCH)
        var input map[string]interface{}
        if endpoint.Method == "POST" || endpoint.Method == "PUT" || endpoint.Method == "PATCH" {
            if err := c.ShouldBindJSON(&input); err != nil {
                // For GET/DELETE, input is empty
                if endpoint.Method != "GET" && endpoint.Method != "DELETE" {
                    response.BadRequest(c, "Invalid request body", requestID)
                    return
                }
            }
        }
        
        // Add query parameters to input (for GET requests)
        if endpoint.Method == "GET" {
            input = make(map[string]interface{})
            for key, values := range c.Request.URL.Query() {
                if len(values) > 0 {
                    input[key] = values[0] // Take first value
                }
            }
        }
        
        // Add path parameters to input
        for _, param := range c.Params {
            input[param.Key] = param.Value
        }
        
        // Create execution context
        execCtx := NewExecutionContext(c.Request.Context(), requestID, GetExtensionDir("extensions", manifest.Name))
        
        // Execute workflow
        result, err := rm.executor.Execute(execCtx, manifest, input)
        if err != nil {
            log.Error().
                Str("request_id", requestID).
                Str("extension", manifest.Name).
                Str("endpoint", endpoint.Path).
                Err(err).
                Msg("Extension endpoint execution failed")
            response.InternalError(c, fmt.Sprintf("Extension execution failed: %v", err), requestID)
            return
        }
        
        // Return response
        c.JSON(http.StatusOK, gin.H{
            "success": true,
            "data":    result,
        })
    }
}

// normalizePath normalizes a path (removes trailing slashes, etc.)
func normalizePath(path string) string {
    // Remove trailing slash (except root)
    if len(path) > 1 && strings.HasSuffix(path, "/") {
        path = path[:len(path)-1]
    }
    return path
}

// UnregisterExtensionRoutes removes routes for an extension
func (rm *RouteManager) UnregisterExtensionRoutes(extensionName string) {
    // Note: Gin doesn't support route removal at runtime
    // This would require router recreation or route tracking
    // For now, routes remain until server restart
    // Future enhancement: implement route removal
}
```

**Testing**:
- Unit tests for route registration
- Test route conflict detection
- Test handler creation
- Test path normalization

---

### Phase 3: Integration with Main Server

**Goal**: Integrate route manager into server startup

**Files to Modify**:
- `cmd/llamagate/main.go`

**Changes**:

1. **Create RouteManager after extension discovery**:
   ```go
   // After extension discovery and registration
   extensionRegistry := extensions.NewRegistry()
   manifests, err := extensions.DiscoverExtensions(extensionBaseDir)
   // ... register extensions ...
   
   // Create route manager
   routeManager := extensions.NewRouteManager(router, extensionRegistry, workflowExecutor)
   
   // Register extension routes
   for _, manifest := range manifests {
       if len(manifest.Endpoints) > 0 {
           if err := routeManager.RegisterExtensionRoutes(manifest); err != nil {
               log.Error().
                   Str("extension", manifest.Name).
                   Err(err).
                   Msg("Failed to register extension routes")
               // Continue with other extensions
           } else {
               log.Info().
                   Str("extension", manifest.Name).
                   Int("endpoints", len(manifest.Endpoints)).
                   Msg("Registered extension routes")
           }
       }
   }
   ```

2. **Update extension refresh endpoint**:
   ```go
   // In RefreshExtensions handler, also refresh routes
   // After registering/updating extensions, re-register routes
   for _, manifest := range manifests {
       if len(manifest.Endpoints) > 0 {
           routeManager.RegisterExtensionRoutes(manifest)
       }
   }
   ```

**Location in main.go**:
- After extension discovery (around line 200)
- Before server start
- After router setup

**Testing**:
- Integration test: server starts with extension routes
- Test route registration during refresh
- Test route conflicts

---

### Phase 4: Authentication and Rate Limiting

**Goal**: Apply auth and rate limiting to extension endpoints

**Files to Modify**:
- `internal/extensions/route_manager.go`
- `cmd/llamagate/main.go`

**Changes**:

1. **Apply middleware conditionally**:
   ```go
   // In registerRoute, apply middleware based on endpoint config
   handler := rm.createEndpointHandler(manifest, endpoint)
   
   // Apply auth middleware if required
   if endpoint.RequiresAuth != nil && *endpoint.RequiresAuth {
       handler = middleware.AuthMiddleware(cfg.APIKey)(handler)
   }
   
   // Apply rate limiting if required
   if endpoint.RequiresRateLimit != nil && *endpoint.RequiresRateLimit {
       handler = rateLimitMiddleware.Handler()(handler)
   }
   ```

2. **Default values**:
   ```go
   // In createEndpointHandler, set defaults
   requiresAuth := true // Default
   if endpoint.RequiresAuth != nil {
       requiresAuth = *endpoint.RequiresAuth
   }
   
   requiresRateLimit := true // Default
   if endpoint.RequiresRateLimit != nil {
       requiresRateLimit = *endpoint.RequiresRateLimit
   }
   ```

**Testing**:
- Test auth enforcement
- Test rate limiting
- Test disabled auth/rate limiting

---

### Phase 5: Schema Validation (Optional)

**Goal**: Validate request/response against schemas

**Files to Modify**:
- `internal/extensions/route_manager.go`

**Changes**:

1. **Add schema validation**:
   ```go
   // In createEndpointHandler, validate request schema
   if endpoint.RequestSchema != nil {
       if err := validateSchema(input, endpoint.RequestSchema); err != nil {
           response.BadRequest(c, fmt.Sprintf("Request validation failed: %v", err), requestID)
           return
       }
   }
   ```

2. **Add validation function**:
   ```go
   func validateSchema(data interface{}, schema map[string]interface{}) error {
       // Use JSON Schema validation library
       // For v1, basic validation; can enhance later
       return nil // Placeholder
   }
   ```

**Note**: This is optional for v1. Can be added later.

---

### Phase 6: Testing

**Goal**: Comprehensive testing

**Test Files to Create**:
- `internal/extensions/route_manager_test.go`
- `internal/extensions/manifest_endpoint_test.go`
- `cmd/llamagate/main_route_test.go`

**Test Cases**:

1. **Manifest Tests**:
   - Load manifest with endpoints
   - Validate endpoint definitions
   - Test invalid endpoints

2. **Route Manager Tests**:
   - Register single endpoint
   - Register multiple endpoints
   - Test route conflicts
   - Test path normalization

3. **Integration Tests**:
   - Server starts with extension routes
   - Routes are accessible
   - Workflow execution works
   - Auth/rate limiting works

4. **End-to-End Tests**:
   - Full webhook receiver test
   - Test with real HTTP requests
   - Test error handling

---

## Implementation Checklist

### Phase 1: Manifest Structure
- [ ] Add `EndpointDefinition` struct
- [ ] Add `Endpoints` field to `Manifest`
- [ ] Add endpoint validation
- [ ] Update manifest loading
- [ ] Write unit tests

### Phase 2: Route Manager
- [ ] Create `route_manager.go`
- [ ] Implement `RouteManager` struct
- [ ] Implement route registration
- [ ] Implement handler creation
- [ ] Add route conflict detection
- [ ] Write unit tests

### Phase 3: Integration
- [ ] Update `main.go` to create RouteManager
- [ ] Register routes during startup
- [ ] Update refresh endpoint
- [ ] Add logging
- [ ] Write integration tests

### Phase 4: Auth & Rate Limiting
- [ ] Apply auth middleware conditionally
- [ ] Apply rate limiting conditionally
- [ ] Handle defaults
- [ ] Write tests

### Phase 5: Schema Validation (Optional)
- [ ] Add schema validation
- [ ] Integrate validation library
- [ ] Write tests

### Phase 6: Testing
- [ ] Unit tests for all components
- [ ] Integration tests
- [ ] End-to-end tests
- [ ] Test with webhook receiver extension

---

## Dependencies

### Required
- Gin framework (already used)
- Extension system (already exists)

### Optional
- JSON Schema validation library (for Phase 5)

## Considerations

### Route Conflicts
- **Solution**: Track routes in map, detect conflicts
- **Behavior**: Fail fast on conflict, log error

### Route Removal
- **Current**: Routes remain until server restart
- **Future**: Implement route removal for hot-reload

### Path Prefixing
- **Pattern**: `/v1/extensions/{extension-name}{endpoint-path}`
- **Example**: `/v1/extensions/github-webhook-receiver/webhooks/github`

### Default Values
- `requires_auth`: `true` (default)
- `requires_rate_limit`: `true` (default)

### Error Handling
- Invalid endpoints: Log error, skip extension
- Route conflicts: Log error, skip conflicting route
- Execution failures: Return 500 with error message

## Testing Strategy

### Unit Tests
- Manifest loading with endpoints
- Route registration
- Handler creation
- Path normalization

### Integration Tests
- Server startup with extension routes
- Route accessibility
- Workflow execution via custom endpoints

### End-to-End Tests
- Full webhook receiver workflow
- Multiple endpoints per extension
- Auth/rate limiting enforcement

## Documentation Updates

### Required Updates
1. **`docs/EXTENSIONS_SPEC_V0.9.1.md`**
   - Mark endpoints as implemented
   - Add examples

2. **`docs/API.md`**
   - Document extension endpoint pattern
   - Add examples

3. **`README.md`**
   - Mention dynamic endpoints feature
   - Link to examples

### Example Documentation
- Webhook receiver extension (already created)
- Add to examples repo after implementation

## Timeline Estimate

- **Phase 1**: 1-2 days (manifest updates)
- **Phase 2**: 2-3 days (route manager)
- **Phase 3**: 1-2 days (integration)
- **Phase 4**: 1 day (auth/rate limiting)
- **Phase 5**: 1-2 days (optional, schema validation)
- **Phase 6**: 2-3 days (testing)

**Total**: ~8-13 days (1.5-2.5 weeks)

## Success Criteria

1. ✅ Extensions can define `endpoints:` in manifest
2. ✅ Routes are automatically registered at startup
3. ✅ Custom endpoints execute extension workflows
4. ✅ Auth and rate limiting work correctly
5. ✅ Webhook receiver extension works end-to-end
6. ✅ Documentation updated
7. ✅ Tests pass

## Next Steps

1. **Review plan** with team
2. **Start Phase 1** (manifest updates)
3. **Iterate** through phases
4. **Test** with webhook receiver extension
5. **Update documentation**
6. **Release** feature

---

**Status**: 📋 Ready for implementation  
**Priority**: High (enables webhook receiver and other integrations)  
**Dependencies**: None (uses existing infrastructure)
