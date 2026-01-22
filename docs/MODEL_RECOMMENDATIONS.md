# Hardware-Based Model Recommendations for LlamaGate

**Based on ArtificialAnalysis.ai Open Source Models & Verified Ollama Availability**  
**Date:** 2026-01-22  
**Source:** [ArtificialAnalysis.ai Open Source Models](https://artificialanalysis.ai/models/open-source)  
**Data Version:** 2.0.0

## 🔍 Automatic Hardware Detection

LlamaGate includes built-in hardware detection that automatically recommends models based on your system's capabilities. **The model recommendations data is embedded directly in the LlamaGate binary** - no external files or configuration required.

- **CPU Detection:** Cores and model information
- **RAM Detection:** Total system memory
- **GPU Detection:** GPU name and VRAM (when available)
- **Data Source:** Model recommendations are compiled into the binary using embedded data

### API Endpoint

```bash
# Get hardware specs and recommended models
curl http://localhost:11434/v1/hardware/recommendations
```

**Response includes:**
- Detected hardware specifications
- Hardware group classification
- **Prioritized list of 3-4 recommended models** (sorted by priority) with:
  - Priority ranking (1 = best match, 2 = second choice, etc.)
  - Intelligence scores (from Artificial Analysis)
  - Parameter counts
  - Hardware requirements
  - Ollama commands
  - Use cases
  - Links to detailed benchmarks

**How to use the recommendations:**
- **Priority 1** = Best overall match - start here for general use
- **Priority 2-3** = Alternative options with different strengths (multilingual, coding, quality)
- **Priority 4+** = Specialized options for specific use cases
- Compare intelligence scores to see quality differences
- Review use cases to match models to your needs

### Example Response

The API returns a **prioritized list of multiple models** (typically 3-4 options) sorted by recommendation priority. Priority 1 is the best match, Priority 2 is the second choice, etc. This gives users multiple options to choose from based on their specific needs.

```json
{
  "success": true,
  "data": {
    "hardware": {
      "cpu_cores": 8,
      "cpu_model": "Intel Core i7-9700K",
      "total_ram_gb": 32,
      "gpu_detected": true,
      "gpu_name": "NVIDIA GeForce RTX 3060",
      "gpu_vram_gb": 12,
      "detection_method": "nvidia-smi"
    },
    "hardware_group": "gpu_8_16gb_vram",
    "recommendations": [
      {
        "name": "Mistral 7B",
        "ollama_name": "mistral",
        "priority": 1,
        "description": "Best balance - quantized for optimal performance",
        "intelligence_score": 7.0,
        "parameters_b": 7.0,
        "min_ram_gb": 8,
        "min_vram_gb": 8,
        "quantized": true,
        "ollama_command": "ollama pull mistral",
        "use_cases": ["general chat", "fast responses", "production workloads"],
        "artificial_analysis_url": "https://artificialanalysis.ai/models/mistral-7b-instruct"
      },
      {
        "name": "Llama 3.2 11B",
        "ollama_name": "llama3.2:11b",
        "priority": 2,
        "description": "Better quality - quantized (requires 12GB+ VRAM)",
        "intelligence_score": 11.0,
        "parameters_b": 11.0,
        "min_ram_gb": 12,
        "min_vram_gb": 12,
        "quantized": true,
        "ollama_command": "ollama pull llama3.2:11b",
        "use_cases": ["general chat", "balanced performance", "quality tasks"],
        "artificial_analysis_url": "https://artificialanalysis.ai/models/llama-3-2-instruct-11b"
      },
      {
        "name": "Qwen 2.5 7B",
        "ollama_name": "qwen2.5:7b",
        "priority": 3,
        "description": "Multilingual option - quantized",
        "intelligence_score": 10.0,
        "parameters_b": 7.0,
        "min_ram_gb": 8,
        "min_vram_gb": 8,
        "quantized": true,
        "ollama_command": "ollama pull qwen2.5:7b",
        "use_cases": ["multilingual", "structured output", "code generation"],
        "artificial_analysis_url": "https://artificialanalysis.ai/models/qwen2-5-7b-instruct"
      },
      {
        "name": "Gemma 3 12B",
        "ollama_name": "gemma3:12b",
        "priority": 4,
        "description": "Quality option - quantized (requires 12GB+ VRAM)",
        "intelligence_score": 9.0,
        "parameters_b": 12.2,
        "min_ram_gb": 12,
        "min_vram_gb": 12,
        "quantized": true,
        "ollama_command": "ollama pull gemma3:12b",
        "use_cases": ["general tasks", "translation", "summarization"],
        "artificial_analysis_url": "https://artificialanalysis.ai/models/gemma-3-12b"
      }
    ]
  }
}
```

**Understanding the Recommendations:**
- **Priority 1** = Best overall match for your hardware (recommended starting point)
- **Priority 2** = Alternative option, often with different strengths (e.g., better quality, multilingual)
- **Priority 3+** = Additional options for specific use cases (multilingual, coding, etc.)

**How to Choose:**
1. **Start with Priority 1** - This is the best general-purpose match
2. **Review Priority 2-3** - Consider if you need specific features (multilingual, structured output, etc.)
3. **Compare intelligence scores** - Higher scores indicate better quality
4. **Check use cases** - Match models to your specific needs

---

---

## 🏢 Typical Business Hardware (On-Premises)

**Most businesses have limited resources:**

| Business Size | Typical RAM | GPU VRAM | Reality |
|--------------|------------|----------|---------|
| **Small-Medium Business (SMB)** | 32-64GB | **None or 16-32GB** | Most common |
| **Larger SMB** | 64-128GB | 16-32GB | Less common |
| **Enterprise** | 128GB+ | 32-48GB | Rare |

**Key Insights:**
- ⚠️ **Most businesses have NO GPU** - Models run on CPU
- ⚠️ **If GPU exists:** Usually 16-32GB VRAM (not 48GB+)
- ✅ **Most common:** 32-64GB total RAM, no dedicated GPU
- 💡 **Recommendation:** Prioritize models that work on CPU or 8-16GB VRAM (quantized)

---

## 🎯 Default Recommendations (Limited Resources)

**For 90% of businesses** (32-64GB RAM, **CPU-only** or 16GB VRAM):

### ⭐ **1. Mistral 7B** - Default Choice for Most Businesses

**Why This is #1:**
- 🚀 **Very fast** - Excellent speed-to-quality ratio
- 💾 **Low memory** - Works on 8GB VRAM (quantized) or CPU
- ✅ **Good accuracy** - Strong instruction following
- 💰 **Cost-effective** - Efficient inference
- 🔧 **Well-optimized** - Great quantization support

**Hardware Requirements:**
- **CPU-only:** ✅ Works well (default assumption - most businesses)
- **VRAM:** 8GB (quantized) or 14GB (full precision) - optional, faster if available
- **Best for:** 90% of businesses with limited resources (CPU-only is most common)

**Ollama Command:**
```bash
ollama pull mistral
```

**Example Usage:**
```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:11435/v1",
    api_key="not-needed"
)

# Default for most businesses
response = client.chat.completions.create(
    model="mistral",
    messages=[{"role": "user", "content": "Summarize this business report"}]
)
```

**Best For:** Fast responses, production workloads, resource-limited setups, most business use cases

---

### 🥈 **2. Llama 3.2 3B** - Lightweight Alternative

**Why Choose This:**
- 📱 **Smallest option** - Works on 6GB VRAM or CPU
- ⚡ **Fastest** - Very fast inference
- 💾 **Lowest memory** - Best for edge/limited resources
- 🔧 **Versatile** - Good for basic tasks

**Hardware Requirements:**
- **VRAM:** 6GB (quantized) or CPU
- **Best for:** Very limited resources, edge deployments

**Ollama Command:**
```bash
ollama pull llama3.2:3b
```

**Example Usage:**
```python
response = client.chat.completions.create(
    model="llama3.2:3b",
    messages=[{"role": "user", "content": "Answer this question"}]
)
```

**Best For:** Edge devices, very limited resources, fastest responses

---

### 🥉 **3. Qwen 2.5 7B** - Multilingual Option

**Why Choose This:**
- 🌐 **Excellent multilingual** support
- 📊 **Strong structured output** - JSON, tables, code
- 💻 **Great for coding** tasks
- 💾 **Works on 8GB VRAM** (quantized)

**Hardware Requirements:**
- **VRAM:** 8GB (quantized) or CPU
- **Best for:** International businesses, structured data needs

**Ollama Command:**
```bash
ollama pull qwen2.5:7b
```

**Example Usage:**
```python
response = client.chat.completions.create(
    model="qwen2.5:7b",
    messages=[{"role": "user", "content": "Generate a JSON schema for a user profile"}]
)
```

**Best For:** Multilingual applications, structured output, code generation

---

## 🏆 Top 5 Models (All Hardware Levels)

### 1. **Mistral 7B** ⚡ Speed Champion (Limited Resources) ⭐ DEFAULT

**Why Choose This:**
- 🚀 **Very fast** - Excellent speed-to-quality ratio
- 💾 **Low memory** - Runs on 8GB VRAM (quantized) or CPU
- ✅ **Good accuracy** - Strong instruction following
- 💰 **Cost-effective** - Efficient inference
- 🔧 **Well-optimized** - Great quantization support

**Hardware Requirements:**
- **VRAM:** 8GB (quantized) or 14GB (full precision)
- **CPU:** Works reasonably on CPU-only systems
- **Best for:** Resource-constrained environments, fast responses, **most businesses**

**Ollama Command:**
```bash
ollama pull mistral
```

**Example Usage:**
```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:11435/v1",
    api_key="not-needed"
)

response = client.chat.completions.create(
    model="mistral",  # Default for most businesses
    messages=[{"role": "user", "content": "Summarize this text"}]
)
```

**Best For:** Fast responses, low-latency applications, resource-limited setups, production workloads

---

### 2. **Llama 3.2** ⚡ Balanced Performance

**Why Choose This:**
- ✅ **Versatile** - Multiple sizes (3B, 11B, 70B, 405B)
- 👁️ **Vision support** - Vision variants available
- 🚀 **Good speed** - Faster than 3.3 on smaller variants
- 📱 **Edge-friendly** - 3B variant for mobile/edge devices
- 🔧 **Large ecosystem** - Well-supported, many tools

**Hardware Requirements:**
- **3B:** 6GB VRAM (quantized) - **Good for limited resources**
- **11B:** 12GB VRAM (quantized) - **Good for businesses with 16GB+ VRAM**
- **70B:** 48GB+ VRAM - **Rare in business environments**
- **Best for:** Various hardware configurations

**Ollama Commands:**
```bash
ollama pull llama3.2:3b    # Small, fast (limited resources)
ollama pull llama3.2:11b   # Balanced (if you have 16GB+ VRAM)
ollama pull llama3.2:70b   # High performance (rare)
```

**Example Usage:**
```python
# For limited resources (3B) - Most businesses
response = client.chat.completions.create(
    model="llama3.2:3b",
    messages=[{"role": "user", "content": "Write a Python function"}]
)

# For balanced performance (11B) - If you have 16GB+ VRAM
response = client.chat.completions.create(
    model="llama3.2:11b",
    messages=[{"role": "user", "content": "Write a Python function"}]
)
```

**Best For:** General tasks, vision applications, edge deployments, balanced performance

---

### 3. **Gemma 3 (12B-27B)** 💎 Quality & Speed

**Why Choose This:**
- ⭐ **Top-tier quality** - Very strong on general tasks
- 🌐 **Excellent translation** and summarization
- ✨ **Better creativity** than Gemma 2
- ⚡ **Good speed** - ~40-45 tok/s for 12B quantized
- 💰 **Cost-effective** - Great quality-to-cost ratio

**Hardware Requirements:**
- **12B:** ~16GB VRAM (quantized) - **Good for businesses with 16GB+ VRAM**
- **27B:** ~18GB VRAM (quantized) - **Requires more resources**
- **Best for:** Mid-range GPUs, desktops, businesses with dedicated AI hardware

**Ollama Commands:**
```bash
ollama pull gemma3:12b
ollama pull gemma3:27b
```

**Example Usage:**
```python
response = client.chat.completions.create(
    model="gemma3:12b",
    messages=[{"role": "user", "content": "Translate this to French: Hello"}]
)
```

**Best For:** General tasks, translation, summarization, creative writing

---

### 4. **Qwen 2.5** 🌍 Multilingual & Structured

**Why Choose This:**
- 🌐 **Excellent multilingual** support
- 📊 **Strong structured output** - JSON, tables, code
- 💻 **Great for coding** tasks
- 📚 **Long context** - Up to 128K tokens
- 🔧 **Multiple sizes** - 0.5B to 110B

**Hardware Requirements:**
- **0.5B:** 4GB VRAM - **Very limited resources**
- **7B:** 8GB VRAM (quantized) - **Good for limited resources** ⭐
- **14B:** 16GB VRAM (quantized) - **Good for businesses with 16GB+ VRAM**
- **72B:** 48GB+ VRAM - **Rare in business environments**
- **Best for:** Multilingual apps, structured data, coding

**Ollama Commands:**
```bash
ollama pull qwen2.5:7b     # Limited resources (most businesses)
ollama pull qwen2.5:14b    # Balanced (if you have 16GB+ VRAM)
ollama pull qwen2.5:72b    # High performance (rare)
```

**Example Usage:**
```python
# For limited resources (7B) - Most businesses
response = client.chat.completions.create(
    model="qwen2.5:7b",
    messages=[{"role": "user", "content": "Generate a JSON schema for a user profile"}]
)
```

**Best For:** Multilingual applications, structured output, code generation, long documents

---

### 5. **Llama 3.3 70B** ⭐ Best Overall (High-End Hardware Only)

**Why Choose This:**
- 🥇 **Best performance** on code generation (88.4% HumanEval)
- 🧠 **Superior reasoning** - 77% on MATH benchmarks
- 📝 **Excellent instruction following** - 92.1% IFEval score
- 🌍 **Strong multilingual** - 91.1% MGSM score
- ⚡ **Efficient** - ~6× less compute than 405B models
- 📚 **Large context** - 128K tokens

**Hardware Requirements:**
- **VRAM:** ~48GB (or quantized for less)
- **Best for:** High-end GPUs, servers, **rare in typical business environments**

**Ollama Command:**
```bash
ollama pull llama3.3:70b
```

**Example Usage:**
```python
response = client.chat.completions.create(
    model="llama3.3:70b",
    messages=[{"role": "user", "content": "Explain quantum computing"}]
)
```

**Best For:** General-purpose tasks, code generation, reasoning, instruction following (when you have the hardware)

---

## Quick Comparison Table

| Model | Quality | Speed | VRAM (Min) | Business Fit |
|-------|---------|-------|------------|--------------|
| **Mistral 7B** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 8GB | ⭐⭐⭐⭐⭐ Most businesses |
| **Llama 3.2 3B** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 6GB | ⭐⭐⭐⭐⭐ Very limited resources |
| **Qwen 2.5 7B** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | 8GB | ⭐⭐⭐⭐⭐ Multilingual businesses |
| **Llama 3.2 11B** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | 12GB | ⭐⭐⭐⭐ Businesses with 16GB+ VRAM |
| **Gemma 3 12B** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | 16GB | ⭐⭐⭐ Businesses with dedicated AI hardware |
| **Qwen 2.5 14B** | ⭐⭐⭐⭐ | ⭐⭐⭐ | 16GB | ⭐⭐⭐ Businesses with 16GB+ VRAM |
| **Llama 3.3 70B** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | 48GB+ | ⭐ Rare (enterprise only) |

⭐ = Recommended for typical business hardware

---

## Selection Guide by Business Hardware

### 🏢 Typical Business (32-64GB RAM, No GPU or 16GB VRAM) - 90% of Businesses

**Top 3 Choices:**
1. **Mistral 7B** - ⭐ **Best default** - Fast, efficient, works on CPU or 8GB VRAM
2. **Llama 3.2 3B** - Smallest, fastest option
3. **Qwen 2.5 7B** - Best for multilingual/structured needs

**Default Recommendation:** `mistral` or `llama3.2:3b`

---

### 💻 Business with 16GB+ VRAM GPU (10% of Businesses)

**Top 3 Choices:**
1. **Llama 3.2 11B** - ⭐ **Best balance** - Good quality, reasonable speed
2. **Gemma 3 12B** - Excellent quality
3. **Qwen 2.5 14B** - Multilingual powerhouse

**Default Recommendation:** `llama3.2:11b`

---

### 🚀 Enterprise/High-End (48GB+ VRAM) - Rare

**Top Choices:**
1. **Llama 3.3 70B** - Best overall quality
2. **Qwen 2.5 72B** - Multilingual + structured
3. **Llama 3.2 70B** - Balanced high performance

**Default Recommendation:** `llama3.3:70b`

---

## Selection Guide by Use Case

| Use Case | Typical Business (No GPU/16GB) | Business with 16GB+ VRAM | Enterprise (48GB+) |
|----------|-------------------------------|--------------------------|-------------------|
| **General Chat** | Mistral 7B | Llama 3.2 11B | Llama 3.3 70B |
| **Code Generation** | Qwen 2.5 7B | Qwen 2.5 14B | Llama 3.3 70B |
| **Multilingual** | Qwen 2.5 7B | Qwen 2.5 14B | Qwen 2.5 72B |
| **Fast Responses** | Mistral 7B | Llama 3.2 11B | Llama 3.2 70B |
| **Structured Output** | Qwen 2.5 7B | Qwen 2.5 14B | Qwen 2.5 72B |
| **Reasoning Tasks** | Llama 3.2 3B | Gemma 3 12B | Llama 3.3 70B |
| **Translation** | Qwen 2.5 7B | Gemma 3 12B | Gemma 3 27B |

---

## Installation & Setup

1. **Pull the model:**
   ```bash
   # For most businesses (default)
   ollama pull mistral
   # or
   ollama pull llama3.2:3b
   
   # For businesses with 16GB+ VRAM
   ollama pull llama3.2:11b
   
   # For enterprise (rare)
   ollama pull llama3.3:70b
   ```

2. **Verify it's available:**
   ```bash
   curl http://localhost:11435/v1/models
   ```

3. **Use in LlamaGate:**
   ```python
   from openai import OpenAI
   
   client = OpenAI(
       base_url="http://localhost:11435/v1",
       api_key="not-needed"
   )
   
   # Default for most businesses
   response = client.chat.completions.create(
       model="mistral",  # or llama3.2:3b
       messages=[{"role": "user", "content": "Hello!"}]
   )
   ```

---

## Default Model Recommendations

### For Documentation Examples

**Current:** Examples use `"mistral"` (Mistral 7B) as default  
**Recommended:** Update to use business-appropriate defaults:

```python
# Most businesses (32-64GB RAM, no GPU or 16GB VRAM) - DEFAULT
model="mistral"  # or llama3.2:3b

# Businesses with 16GB+ VRAM
model="llama3.2:11b"

# Enterprise (48GB+ VRAM) - Rare
model="llama3.3:70b"
```

### For New Users

**Start with:** `mistral` or `llama3.2:3b` (works on most business hardware)  
**Upgrade to:** `llama3.2:11b` if you have 16GB+ VRAM  
**Best quality:** `llama3.3:70b` if you have 48GB+ VRAM (rare)

---

## Benchmark Sources

- [ArtificialAnalysis.ai Open Source Models](https://artificialanalysis.ai/models/open-source) - Comprehensive open-source model benchmarks with intelligence scores, parameter counts, and performance metrics
- [Ollama Library](https://ollama.com/library) - Verified model availability and download commands
- All models in recommendations are verified to be available in Ollama and sourced from open-source models with downloadable weights

## Model Data Fields

Each recommended model includes:

- **Intelligence Score:** Artificial Analysis Intelligence Index (higher is better)
- **Parameters:** Model size in billions of parameters
- **Hardware Requirements:** Minimum RAM/VRAM needed
- **Quantization Status:** Whether the model is quantized for efficiency
- **Ollama Command:** Ready-to-use command to download the model
- **Use Cases:** Recommended applications for the model
- **Benchmark Links:** Direct links to detailed Artificial Analysis benchmarks

---

## Summary

**Top 5 Models for LlamaGate:**

1. **Mistral 7B** - ⭐ **Default for 90% of businesses** (8GB VRAM or CPU)
2. **Llama 3.2** - Versatile, multiple sizes (3B to 70B)
3. **Gemma 3** - Quality & speed balance (12B-27B)
4. **Qwen 2.5** - Multilingual & structured (7B-72B)
5. **Llama 3.3 70B** - Best overall quality (high-end hardware only)

**Quick Start for Most Businesses:**
- **CPU-only (most common - 90% of businesses):** `ollama pull mistral` ⭐ **Default**
- **If you have 8GB+ VRAM:** `ollama pull mistral` (same model, faster)
- **If you have 16GB+ VRAM:** `ollama pull llama3.2:11b`
- **Enterprise (rare):** `ollama pull llama3.3:70b`

**Key Takeaway:** Most businesses should start with **Mistral 7B** or **Llama 3.2 3B** - they work on typical business hardware (32-64GB RAM, **CPU-only is the default assumption**).

---

**Last Updated:** 2026-01-15  
**Next Review:** Quarterly or when major model releases occur
