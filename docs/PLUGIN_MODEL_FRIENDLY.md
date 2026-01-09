# Plugin System - Model-Friendly Design

This document explains how the plugin system is designed to be easily understood, defined, and used by local LLM models.

## Is It Model-Friendly? Yes! ✅

The plugin system is designed with AI models in mind:

### 1. **Self-Documenting Structure**
- Clear metadata with descriptions
- JSON Schema for inputs/outputs
- Human-readable names and descriptions

### 2. **Simple JSON Interface**
- All data structures are JSON-serializable
- Easy to parse and generate
- No complex Go-specific concepts

### 3. **Declarative Definitions**
- Plugins can be described in JSON/YAML
- Schemas are standard JSON Schema
- Workflows are declarative

## How Models Can Define Plugins

### Option 1: JSON/YAML Definition (Recommended for Models)

A model can define a plugin in pure JSON:

```json
{
  "name": "text_processor",
  "version": "1.0.0",
  "description": "Processes and analyzes text",
  "input_schema": {
    "type": "object",
    "properties": {
      "text": {
        "type": "string",
        "description": "Text to process"
      },
      "operation": {
        "type": "string",
        "enum": ["summarize", "analyze", "extract"],
        "description": "Operation to perform"
      }
    },
    "required": ["text", "operation"]
  },
  "output_schema": {
    "type": "object",
    "properties": {
      "result": {
        "type": "string",
        "description": "Processed result"
      }
    }
  },
  "workflow": {
    "steps": [
      {
        "id": "process",
        "type": "llm_call",
        "config": {
          "model": "llama3.2",
          "prompt": "Process this text: {{text}} with operation: {{operation}}"
        }
      }
    ]
  }
}
```

### Option 2: Natural Language Description

A model can describe a plugin in natural language, and the system can generate the definition:

**Model Prompt:**
```
Create a plugin that:
- Takes text and a style parameter
- Summarizes the text in the specified style
- Returns the summary and metrics
```

**Generated Definition:**
```json
{
  "name": "text_summarizer",
  "description": "Summarizes text in specified style",
  "input_schema": {
    "properties": {
      "text": {"type": "string"},
      "style": {"type": "string", "enum": ["brief", "detailed"]}
    },
    "required": ["text", "style"]
  }
}
```

## Model-Friendly Features

### 1. Clear Schema Structure

All schemas use standard JSON Schema, which models understand well:

```json
{
  "type": "object",
  "properties": {
    "field_name": {
      "type": "string",
      "description": "What this field does",
      "default": "default_value"
    }
  },
  "required": ["field_name"]
}
```

### 2. Self-Explanatory Metadata

Every plugin describes itself:

```go
Metadata: plugins.PluginMetadata{
    Name:        "plugin_name",           // Clear identifier
    Description: "What it does",          // Human-readable
    InputSchema: {...},                    // Structured definition
    OutputSchema: {...},                   // Clear outputs
}
```

### 3. Declarative Workflows

Workflows are defined declaratively, easy for models to generate:

```json
{
  "workflow": {
    "steps": [
      {
        "id": "step1",
        "type": "llm_call",
        "config": {
          "model": "llama3.2",
          "prompt": "Do something"
        }
      },
      {
        "id": "step2",
        "type": "tool_call",
        "config": {
          "tool_name": "mcp.filesystem.read_file"
        },
        "dependencies": ["step1"]
      }
    ]
  }
}
```

## How Models Can Use Plugins

### 1. Discovery

Models can query available plugins:

```bash
GET /v1/plugins
```

Returns:
```json
{
  "plugins": [
    {
      "name": "text_summarizer",
      "description": "Summarizes text",
      "input_schema": {...},
      "output_schema": {...}
    }
  ]
}
```

### 2. Understanding

Models can get detailed information:

```bash
GET /v1/plugins/text_summarizer
```

Returns full metadata with schemas.

### 3. Execution

Models can execute plugins with structured input:

```bash
POST /v1/plugins/text_summarizer/execute
{
  "text": "Long text...",
  "style": "brief"
}
```

## Model-Generated Plugin Example

Here's how a model might generate a plugin definition:

**Model Output:**
```json
{
  "plugin_definition": {
    "name": "sentiment_analyzer",
    "description": "Analyzes sentiment of text input",
    "input_schema": {
      "type": "object",
      "properties": {
        "text": {
          "type": "string",
          "description": "Text to analyze"
        },
        "detail_level": {
          "type": "string",
          "enum": ["basic", "detailed"],
          "default": "basic"
        }
      },
      "required": ["text"]
    },
    "output_schema": {
      "type": "object",
      "properties": {
        "sentiment": {
          "type": "string",
          "enum": ["positive", "negative", "neutral"]
        },
        "confidence": {
          "type": "number",
          "minimum": 0,
          "maximum": 1
        }
      }
    },
    "workflow": {
      "steps": [
        {
          "id": "analyze",
          "type": "llm_call",
          "config": {
            "model": "llama3.2",
            "prompt": "Analyze sentiment of: {{text}}. Detail level: {{detail_level}}"
          }
        }
      ]
    }
  }
}
```

## Simplifications for Models

### 1. Minimal Required Fields

For a model to define a plugin, only these are required:

```json
{
  "name": "plugin_name",
  "description": "What it does",
  "required_inputs": ["input1", "input2"]
}
```

Everything else is optional!

### 2. Flexible Input/Output

- Inputs/outputs are `map[string]interface{}` - very flexible
- No strict typing required
- Models can work with any structure

### 3. Optional Validation

Validation is optional - models can skip it initially:

```go
func (p *Plugin) ValidateInput(input map[string]interface{}) error {
    return nil // Models can start with no validation
}
```

## Model-Friendly API

### Simple Execution

```bash
POST /v1/plugins/{name}/execute
Content-Type: application/json

{
  "input1": "value1",
  "input2": "value2"
}
```

### Clear Responses

```json
{
  "success": true,
  "data": {
    "result": "..."
  },
  "metadata": {
    "execution_time": "1.2s",
    "steps_executed": 3
  }
}
```

## Example: Model Creating a Plugin

**Step 1: Model generates definition**
```json
{
  "name": "data_formatter",
  "description": "Formats data in requested format",
  "input_schema": {
    "properties": {
      "data": {"type": "object"},
      "format": {"type": "string", "enum": ["json", "yaml", "csv"]}
    },
    "required": ["data", "format"]
  }
}
```

**Step 2: System creates plugin from definition**
```go
// Auto-generated from JSON definition
plugin := CreatePluginFromDefinition(definition)
```

**Step 3: Model uses the plugin**
```bash
POST /v1/plugins/data_formatter/execute
{
  "data": {...},
  "format": "json"
}
```

## Benefits for Models

### ✅ Easy to Understand
- Clear structure
- Self-documenting
- Standard schemas

### ✅ Easy to Generate
- JSON/YAML format
- Declarative structure
- Minimal boilerplate

### ✅ Easy to Use
- Simple API
- Clear inputs/outputs
- Structured responses

### ✅ Flexible
- Optional features
- Progressive complexity
- No strict requirements

## Summary

**Yes, the plugin system is very model-friendly:**

1. **Easy to Define**: JSON/YAML definitions, declarative structure
2. **Easy to Understand**: Self-documenting, clear schemas
3. **Easy to Use**: Simple API, structured I/O
4. **Flexible**: Optional features, minimal requirements

Models can:
- ✅ Generate plugin definitions in JSON
- ✅ Understand plugin capabilities from metadata
- ✅ Execute plugins via simple API
- ✅ Create workflows declaratively

The system is designed to be both human-friendly and AI-friendly!
