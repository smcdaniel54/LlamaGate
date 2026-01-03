# ⚡ LlamaGate Quick Start

**Get from zero to running in 2 minutes.**

## What is LlamaGate?

LlamaGate is an **OpenAI-compatible API gateway** for local LLMs. Switch from OpenAI to local models with **one line of code**.

```python
# Before: OpenAI (costs money, sends data externally)
client = OpenAI(api_key="sk-...")

# After: LlamaGate (free, 100% local, same code!)
client = OpenAI(base_url="http://localhost:8080/v1", api_key="sk-llamagate")
```

That's it. Your existing OpenAI code works immediately.

---

## 🚀 2-Minute Setup

### Step 1: Install (1 minute)

**Windows:**
```cmd
install\windows\install.cmd
```

**Mac/Linux:**
```bash
chmod +x install.sh && ./install.sh
```

### Step 2: Start (30 seconds)

**Windows:**
```cmd
run.cmd
```

**Mac/Linux:**
```bash
./run.sh
```

### Step 3: Run the Demo (30 seconds)

**Windows:**
```cmd
demo.cmd
```

**Mac/Linux:**
```bash
chmod +x demo.sh && ./demo.sh
```

**🎉 Done!** You now have a local OpenAI-compatible API running.

---

## 💡 The Magic: One-Line Migration

Your existing OpenAI code works without changes:

```python
from openai import OpenAI

# Just change the base_url - that's it!
client = OpenAI(
    base_url="http://localhost:8080/v1",  # ← Only change needed
    api_key="sk-llamagate"  # Or leave empty if auth disabled
)

# Everything else stays the same
response = client.chat.completions.create(
    model="llama2",  # Use any Ollama model
    messages=[{"role": "user", "content": "Hello!"}]
)

print(response.choices[0].message.content)
```

---

## ⚠️ Important: Model Loading Behavior

**Understanding Model Loading Latency:**

- **First request to a model:** 5-30+ seconds (Ollama loads model weights into memory)
- **Subsequent requests:** Fast (model stays loaded in memory)
- **Switching to a different model:** First request is slow again (new model needs loading)

**Why?** Ollama needs to load model weights (often several GB) into RAM/VRAM. This is a one-time cost per model.

**Best Practices:**

1. **Pre-warm models** - Make a simple request at app startup:
   ```python
   # Pre-warm on startup
   client.chat.completions.create(
       model="llama2",
       messages=[{"role": "user", "content": "hi"}]
   )
   ```

2. **Stick to one model** - For consistent performance, use one model per app instance

3. **Use caching** - LlamaGate caches responses, so identical requests are instant

4. **Monitor memory** - Large models need significant RAM/VRAM

**Note:** This is normal behavior for local LLMs. The trade-off is privacy and zero API costs!

---

## 🎯 What You Get

- ✅ **Zero API costs** - Run models locally
- ✅ **100% private** - Your data never leaves your machine
- ✅ **OpenAI-compatible** - Drop-in replacement
- ✅ **Smart caching** - Identical requests are instant
- ✅ **Any model** - Switch between Ollama models (with loading time on first use)

---

## 📚 Next Steps

- **Full Documentation:** See [README.md](README.md)
- **Migration Guide:** See [DEMO_QUICKSTART.md](DEMO_QUICKSTART.md)
- **Configuration:** Copy `.env.example` to `.env` and customize settings

---

**Questions?** Check the [README](README.md) or open an issue.

