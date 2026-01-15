# Top 5 Model Recommendations for LlamaGate

**Based on ArtificialAnalysis.ai Benchmarks (v4.0) & Typical Business Hardware**  
**Date:** 2026-01-15  
**Source:** [ArtificialAnalysis.ai](https://artificialanalysis.ai)

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

- [ArtificialAnalysis.ai](https://artificialanalysis.ai) - Real-world capability benchmarks (v4.0)
- [Ollama Models](https://ollama.com/library) - Available models and sizes
- Performance data from community benchmarks and testing

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
