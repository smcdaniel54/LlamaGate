# LlamaGate Extensions Documentation Index

**Version:** 0.9.1  
**Last Updated:** 2026-01-10

---

## Documentation Overview

Complete documentation for the LlamaGate Extensions system v0.9.1.

---

## User Documentation

### 🚀 [Extension Quick Start](./EXTENSIONS_QUICKSTART.md)
**For:** End users and developers who want to create and use extensions  
**Content:**
- Write, validate, load, and run extensions
- Extension types (workflow, middleware, observer)
- Available step types
- CLI commands
- Common failure modes and troubleshooting

**Start here if:** You want to create your own extensions or understand how they work.

---

### 📖 [Extensions README](../extensions/README.md)
**For:** Extension users and developers  
**Content:**
- Overview of the three example extensions
- Extension API endpoints
- Extension structure
- Creating your own extensions
- Configuration options

**Start here if:** You want to understand how extensions work or create your own.

---

## Developer Documentation

### 📋 [Extension Specification](./EXTENSIONS_SPEC_V0.9.1.md)
**For:** Developers implementing the extension system  
**Content:**
- Complete extension specification
- Manifest schema
- Lifecycle management
- Safety boundaries
- Versioning guarantees

**Start here if:** You're implementing or extending the extension system.

---

### 🛠️ [Implementation Plan](./EXTENSIONS_IMPLEMENTATION_PLAN.md)
**For:** Developers migrating from plugins to extensions  
**Content:**
- 12-phase migration plan
- Step-by-step tasks
- Risk mitigation
- Success criteria

**Start here if:** You're planning the migration from plugins to extensions.

---

### ✅ Migration Checklist
**Status:** Migration complete - checklist no longer needed  
**Note:** The plugins-to-extensions migration was completed in v0.9.1. This checklist has been removed as it's no longer relevant.

---

### 🧪 [Testing Documentation](./EXTENSIONS_TESTING.md)
**For:** Developers and QA  
**Content:**
- Complete test coverage
- Test execution instructions
- Integration test examples
- Acceptance criteria validation

**Start here if:** You want to understand test coverage or run tests.

---

### 📊 [Summary](./EXTENSIONS_SUMMARY.md)
**For:** Project managers and stakeholders  
**Content:**
- Design decisions summary
- Migration scope
- Status and next steps

**Start here if:** You want a high-level overview.

---

## API Documentation

### 🌐 [API Reference](./API.md)
**For:** API consumers  
**Content:**
- Extension API endpoints
- Request/response formats
- Error handling
- Authentication

**Sections:**
- `GET /v1/extensions` - List all extensions
- `GET /v1/extensions/:name` - Get extension details
- `PUT /v1/extensions/:name` - Upsert extension (optional; `EXTENSIONS_UPSERT_ENABLED=true`)
- `POST /v1/extensions/:name/execute` - Execute workflow extension
- `POST /v1/extensions/refresh` - Re-discover extensions

**Start here if:** You're integrating with the extension API.

---

## Example Extensions

### 1. Prompt Template Executor
**Location:** `extensions/prompt-template-executor/`  
**Type:** Workflow  
**Documentation:** [Extensions README](../extensions/README.md#1-prompt-template-executor)

**What it does:**
- Loads prompt templates
- Renders with variables
- Calls LLM
- Writes output files

---

### 2. Request Inspector
**Location:** `extensions/request-inspector/`  
**Type:** Middleware  
**Documentation:** [Extensions README](../extensions/README.md#2-request-inspector)

**What it does:**
- Intercepts HTTP requests
- Creates audit logs
- Applies redaction rules

---

### 3. Cost Usage Reporter
**Location:** `extensions/cost-usage-reporter/`  
**Type:** Observer  
**Documentation:** [Extensions README](../extensions/README.md#3-cost-usage-reporter)

**What it does:**
- Tracks token usage
- Generates usage reports
- Accumulates cost data

---

## Documentation by Use Case

### I want to...

**...use the example extensions:**
1. Read [Extension Quick Start](./EXTENSION_QUICKSTART.md)
2. Check [Extensions README](../extensions/README.md)

**...create my own extension:**
1. Read [Extension Specification](./EXTENSIONS_SPEC_V0.9.1.md) - Manifest schema
2. Review [Extensions README](../extensions/README.md) - Structure and examples
3. Check example extensions in `extensions/` directory

**...understand the architecture:**
1. Read [Extension Specification](./EXTENSIONS_SPEC_V0.9.1.md) - Complete spec
2. Review [Summary](./EXTENSIONS_SUMMARY.md) - High-level overview

**...migrate from plugins:**
1. Review [Extension Specification](./EXTENSIONS_SPEC_V0.9.1.md) - Migration section
2. See [PLANS_REVIEW.md](./PLANS_REVIEW.md) for migration completion details
3. Note: Migration was completed in v0.9.1 - plugins system has been removed

**...test extensions:**
1. Read [Testing Documentation](./EXTENSIONS_TESTING.md)
2. Run tests: `go test ./internal/extensions/... -v`

**...integrate via API:**
1. Read [API Reference](./API.md) - Extension endpoints section
2. Review [Extension Quick Start](./EXTENSIONS_QUICKSTART.md) - API examples

---

## Quick Links

- **Specification:** [EXTENSIONS_SPEC_V0.9.1.md](./EXTENSIONS_SPEC_V0.9.1.md)
- **Quick Start:** [EXTENSIONS_QUICKSTART.md](./EXTENSIONS_QUICKSTART.md)
- **API Reference:** [API.md](./API.md) (Extension endpoints section)
- **Example Extensions:** [extensions/README.md](../extensions/README.md)
- **Testing:** [EXTENSIONS_TESTING.md](./EXTENSIONS_TESTING.md)

---

## Documentation Status

| Document | Status | Audience |
|----------|--------|----------|
| Extension Specification | ✅ Complete | Developers |
| Extension Quick Start | ✅ Complete | Users |
| Extensions README | ✅ Complete | Users/Developers |
| API Reference | ✅ Complete | API Consumers |
| Implementation Plan | ✅ Complete (Historical) | Developers |
| Migration Checklist | ✅ Complete (Removed) | Developers |
| Testing Documentation | ✅ Complete | Developers/QA |
| Summary | ✅ Complete | Stakeholders |

---

**All extension documentation is complete and ready for use.** ✅

*Last Updated: 2026-01-10*
