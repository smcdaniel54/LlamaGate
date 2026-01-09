# Plugin System UX Improvements

Actionable roadmap to transform the plugin system into a UX designer's dream.

## Quick Reference

| Priority | Item | Impact | Effort | Status |
|----------|------|--------|--------|--------|
| P1 | Enhanced Error Messages | ⭐⭐⭐⭐⭐ | Medium | 📋 Planned |
| P1 | OpenAPI/Swagger Docs | ⭐⭐⭐⭐⭐ | Medium | 📋 Planned |
| P1 | Validation Feedback | ⭐⭐⭐⭐⭐ | Low | 📋 Planned |
| P2 | Discovery Enhancement | ⭐⭐⭐⭐ | Medium | 📋 Planned |
| P2 | Rate Limit Headers | ⭐⭐⭐⭐ | Low | 📋 Planned |
| P2 | Async Execution | ⭐⭐⭐⭐ | High | 📋 Planned |
| P3 | Playground UI | ⭐⭐⭐ | High | 📋 Planned |
| P3 | Analytics | ⭐⭐⭐ | Medium | 📋 Planned |

---

## Priority 1: Critical (Do First)

### 1. Enhanced Error Messages

**Problem**: Generic errors don't help users fix issues.

**Current:**
```json
{
  "error": {
    "type": "invalid_request_error",
    "message": "Plugin execution failed"
  }
}
```

**Target:**
```json
{
  "error": {
    "type": "validation_error",
    "message": "Input validation failed",
    "details": {
      "field": "text",
      "issue": "required field is missing",
      "expected": "string",
      "received": null,
      "example": "Hello, world!"
    },
    "request_id": "req_123",
    "documentation": "/docs/plugins/text_summarizer#inputs"
  }
}
```

**Implementation Steps:**

1. **Create validation error structure** (`internal/response/validation.go`):
   ```go
   type ValidationError struct {
       Field   string      `json:"field"`
       Issue   string      `json:"issue"`
       Expected interface{} `json:"expected,omitempty"`
       Received interface{} `json:"received,omitempty"`
       Example interface{} `json:"example,omitempty"`
   }
   ```

2. **Update error response helpers** (`internal/response/errors.go`):
   ```go
   func ValidationError(c *gin.Context, errors []ValidationError, requestID string) {
       // Return structured validation errors
   }
   ```

3. **Enhance plugin validation** (`internal/api/plugins.go`):
   - Parse validation errors from plugins
   - Extract field-level information
   - Include examples from plugin metadata

**Files to Modify:**
- `internal/response/errors.go` - Add validation error helpers
- `internal/api/plugins.go` - Enhanced error handling
- `internal/plugins/types.go` - Add validation error types (optional)

**Estimated Effort**: 4-6 hours

---

### 2. OpenAPI/Swagger Specification

**Problem**: No interactive API documentation.

**Target**: Interactive Swagger UI at `/swagger` with try-it-out functionality.

**Implementation Steps:**

1. **Add Swagger dependencies**:
   ```bash
   go get -u github.com/swaggo/swag/cmd/swag
   go get -u github.com/swaggo/gin-swagger
   go get -u github.com/swaggo/files
   ```

2. **Add Swagger annotations** to handlers:
   ```go
   // @Summary List all plugins
   // @Description Returns a list of all registered plugins
   // @Tags plugins
   // @Produce json
   // @Success 200 {object} PluginListResponse
   // @Router /v1/plugins [get]
   func (h *PluginHandler) ListPlugins(c *gin.Context) {
       // ...
   }
   ```

3. **Generate and serve spec** (`cmd/llamagate/main.go`):
   ```go
   import "github.com/swaggo/gin-swagger"
   import "github.com/swaggo/files"
   
   router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
   ```

4. **Generate spec**:
   ```bash
   swag init -g cmd/llamagate/main.go
   ```

**Files to Modify:**
- `cmd/llamagate/main.go` - Add Swagger route
- `internal/api/plugins.go` - Add annotations
- `go.mod` - Add dependencies

**Estimated Effort**: 6-8 hours

---

### 3. Validation Feedback with Examples

**Problem**: Users must guess valid input format.

**Target**: Return validation errors with examples and valid ranges.

**Implementation Steps:**

1. **Enhance plugin metadata** to include examples:
   ```go
   type PluginMetadata struct {
       // ... existing fields
       ExampleRequest  map[string]interface{} `json:"example_request,omitempty"`
       ExampleResponse map[string]interface{} `json:"example_response,omitempty"`
   }
   ```

2. **Update validation error response** to include:
   - Field-level errors
   - Valid ranges (min/max)
   - Example values
   - Complete example request

3. **Modify plugin handler** to extract examples from metadata:
   ```go
   if err := plugin.ValidateInput(input); err != nil {
       metadata := plugin.Metadata()
       return ValidationErrorResponse(c, err, metadata.ExampleRequest, requestID)
   }
   ```

**Files to Modify:**
- `internal/plugins/types.go` - Add example fields
- `internal/api/plugins.go` - Include examples in errors
- `internal/response/errors.go` - Validation error response

**Estimated Effort**: 3-4 hours

---

## Priority 2: High Value (Do Soon)

### 4. Enhanced Discovery Endpoint

**Problem**: Hard to find and filter plugins.

**Target**: Rich discovery with categories, tags, search, and examples.

**Implementation Steps:**

1. **Extend plugin metadata**:
   ```go
   type PluginMetadata struct {
       // ... existing fields
       Category string   `json:"category,omitempty"`
       Tags     []string `json:"tags,omitempty"`
       ExampleRequest  map[string]interface{} `json:"example_request,omitempty"`
       ExampleResponse map[string]interface{} `json:"example_response,omitempty"`
   }
   ```

2. **Add query parameters** to `ListPlugins`:
   - `?category=text-processing`
   - `?tag=summarization`
   - `?search=text`

3. **Enhance response**:
   ```go
   type PluginListResponse struct {
       Plugins   []PluginMetadata `json:"plugins"`
       Count     int              `json:"count"`
       Categories []string         `json:"categories,omitempty"`
       Search    SearchOptions    `json:"search,omitempty"`
   }
   ```

**Files to Modify:**
- `internal/plugins/types.go` - Add category/tags
- `internal/api/plugins.go` - Filtering logic
- Plugin examples - Add metadata

**Estimated Effort**: 4-6 hours

---

### 5. Rate Limiting Headers

**Problem**: Users can't plan retries when rate limited.

**Target**: Include rate limit info in all responses.

**Implementation Steps:**

1. **Update rate limit middleware** (`internal/middleware/rate_limit.go`):
   ```go
   func (rl *RateLimitMiddleware) Handler() gin.HandlerFunc {
       return func(c *gin.Context) {
           // ... existing logic
           
           // Always set headers
           c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
           c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
           c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))
           
           if !allowed {
               c.Header("Retry-After", strconv.Itoa(retryAfter))
           }
           
           c.Next()
       }
   }
   ```

2. **Calculate reset time** based on limiter state.

**Files to Modify:**
- `internal/middleware/rate_limit.go` - Add headers

**Estimated Effort**: 2-3 hours

---

### 6. Async Execution (Optional)

**Problem**: No feedback for long-running plugins.

**Target**: Job-based execution with status polling.

**Implementation Steps:**

1. **Create job system** (`internal/plugins/jobs.go`):
   ```go
   type JobManager struct {
       jobs map[string]*Job
       mu   sync.RWMutex
   }
   
   type Job struct {
       ID        string
       Status    string
       Progress  float64
       Result    *PluginResult
       Error     error
       CreatedAt time.Time
   }
   ```

2. **Add endpoints**:
   - `POST /v1/plugins/:name/execute` → Returns job ID
   - `GET /v1/plugins/jobs/:job_id` → Returns job status
   - `DELETE /v1/plugins/jobs/:job_id` → Cancel job

3. **Update plugin handler** to support async mode:
   ```go
   if async := c.Query("async"); async == "true" {
       jobID := jobManager.CreateJob(plugin, input)
       return JobResponse{JobID: jobID, StatusURL: "/v1/plugins/jobs/" + jobID}
   }
   ```

**Files to Create:**
- `internal/plugins/jobs.go` - Job management
- `internal/api/plugin_jobs.go` - Job endpoints

**Estimated Effort**: 8-12 hours

**Note**: Only implement if plugins can be long-running (>5 seconds).

---

## Priority 3: Polish (Nice to Have)

### 7. Plugin Playground UI

**Target**: Web UI for testing plugins interactively.

**Implementation**: Separate frontend project or embedded HTML/JS.

**Estimated Effort**: 16-24 hours

---

### 8. Plugin Analytics

**Target**: Usage statistics and performance metrics.

**Implementation Steps:**

1. **Add analytics tracking** to plugin execution
2. **Create analytics endpoint**: `GET /v1/plugins/:name/stats`
3. **Track**: executions, success rate, avg time, errors

**Estimated Effort**: 6-8 hours

---

### 9. Plugin Versioning

**Target**: Support multiple plugin versions.

**Implementation**: Extend registry to handle versions, add version endpoints.

**Estimated Effort**: 4-6 hours

---

## Implementation Roadmap

### Week 1-2: Critical Items
- [ ] Enhanced error messages (4-6h)
- [ ] OpenAPI/Swagger (6-8h)
- [ ] Validation feedback (3-4h)
- **Total**: ~13-18 hours

### Week 3-4: High Value
- [ ] Discovery enhancement (4-6h)
- [ ] Rate limit headers (2-3h)
- [ ] Async execution (if needed) (8-12h)
- **Total**: ~14-21 hours

### Week 5+: Polish
- [ ] Playground UI (16-24h)
- [ ] Analytics (6-8h)
- [ ] Versioning (4-6h)
- **Total**: ~26-38 hours

---

## Quick Wins (Can Do Today)

### 1. Add Rate Limit Headers (30 minutes)
```go
// In internal/middleware/rate_limit.go
c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
```

### 2. Add Example to Plugin Metadata (15 minutes)
```go
// In plugin Metadata() method
ExampleRequest: map[string]interface{}{
    "text": "Example text...",
    "max_length": 200,
}
```

### 3. Enhance Error Message (30 minutes)
```go
// In internal/api/plugins.go
response.BadRequest(c, fmt.Sprintf("%s. Example: %v", err.Error(), example), requestID)
```

**Total Quick Wins**: ~1.25 hours

---

## Success Metrics

### User Experience
- **Time to first success**: < 5 minutes (current: ~15 minutes)
- **Self-service error resolution**: > 80% (current: ~40%)
- **Plugin discovery**: > 90% find without docs (current: ~60%)

### Developer Experience
- **Documentation completeness**: 100% endpoint coverage
- **Example availability**: 1+ example per endpoint
- **Support tickets**: < 10% for errors (current: ~30%)

---

## Current State Assessment

### ✅ Strengths
- Simple RESTful API
- Self-documenting metadata
- Basic error handling
- Good documentation

### ⚠️ Gaps
- Generic error messages
- No interactive docs
- Limited validation feedback
- Basic discovery
- No rate limit info

### 🎯 Target State
- Detailed, actionable errors
- Interactive API docs
- Rich validation feedback
- Enhanced discovery
- Complete rate limit info

---

## Next Steps

1. **Review and prioritize** based on user feedback
2. **Start with Quick Wins** (1.25 hours)
3. **Implement Priority 1** (13-18 hours)
4. **Measure impact** and iterate
5. **Continue with Priority 2** based on results

**With these improvements, the plugin system will be exceptional!** 🚀
