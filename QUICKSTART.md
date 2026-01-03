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
scripts\windows\install.cmd
```

**Mac/Linux:**

```bash
chmod +x scripts/unix/install.sh && ./scripts/unix/install.sh
```

### Step 2: Start (30 seconds)

**Windows:**

```cmd
scripts\windows\run.cmd
```

**Mac/Linux:**

```bash
./scripts/unix/run.sh
```

### Step 3: Run the Demo (30 seconds)

**Windows:**

```cmd
scripts\windows\demo.cmd
```

**Mac/Linux:**

```bash
chmod +x scripts/unix/demo.sh && ./scripts/unix/demo.sh
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

## 🔄 Migration Examples

### Example 1: Python with OpenAI SDK

**Before (Using OpenAI):**

```python
from openai import OpenAI

client = OpenAI(
    api_key="sk-your-openai-key-here"  # Requires paid API key
)

response = client.chat.completions.create(
    model="gpt-3.5-turbo",
    messages=[
        {"role": "user", "content": "Hello! Explain quantum computing in one sentence."}
    ]
)

print(response.choices[0].message.content)
```

**After (Using LlamaGate) ✨:**

```python
from openai import OpenAI

# Just change the base_url - that's it!
client = OpenAI(
    base_url="http://localhost:8080/v1",  # ← LlamaGate endpoint
    api_key="sk-llamagate"  # Your API_KEY from .env (or omit if not set)
)

# Same code, different model!
response = client.chat.completions.create(
    model="llama2",  # ← Any Ollama model you have installed
    messages=[
        {"role": "user", "content": "Hello! Explain quantum computing in one sentence."}
    ]
)

print(response.choices[0].message.content)
```

### Example 2: JavaScript/Node.js

**Before:**

```javascript
import OpenAI from 'openai';

const openai = new OpenAI({
  apiKey: 'sk-your-openai-key-here',
});

const response = await openai.chat.completions.create({
  model: 'gpt-3.5-turbo',
  messages: [
    { role: 'user', content: 'Write a haiku about coding' }
  ],
});

console.log(response.choices[0].message.content);
```

**After:**

```javascript
import OpenAI from 'openai';

const openai = new OpenAI({
  baseURL: 'http://localhost:8080/v1',  // ← LlamaGate endpoint
  apiKey: 'sk-llamagate',  // Your API_KEY from .env (or omit if not set)
});

const response = await openai.chat.completions.create({
  model: 'llama2',  // ← Any Ollama model
  messages: [
    { role: 'user', content: 'Write a haiku about coding' }
  ],
});

console.log(response.choices[0].message.content);
```

### Example 3: cURL

**Before (OpenAI API):**

```bash
curl https://api.openai.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-your-openai-key" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

**After (LlamaGate) ✨:**

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "model": "llama2",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

**Changes:**

1. URL: `https://api.openai.com/v1` → `http://localhost:8080/v1`
2. Header: `Authorization: Bearer` → `X-API-Key` (or keep `Authorization: Bearer`)
3. Model: `gpt-3.5-turbo` → `llama2` (or any Ollama model)

---

## 🔀 Model Switching

One of LlamaGate's superpowers is **instant model switching**. You can switch between different models with zero code changes - just change the model name!

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="sk-llamagate"
)

# Use Llama 2
response1 = client.chat.completions.create(
    model="llama2",
    messages=[{"role": "user", "content": "Explain AI in simple terms"}]
)

# Switch to Mistral (if you have it installed)
response2 = client.chat.completions.create(
    model="mistral",  # ← Just change the model name!
    messages=[{"role": "user", "content": "Explain AI in simple terms"}]
)

# Switch to CodeLlama
response3 = client.chat.completions.create(
    model="codellama",  # ← Another model, same code!
    messages=[{"role": "user", "content": "Write a Python function to sort a list"}]
)
```

### Installing More Models

```bash
# Install different models
ollama pull llama2
ollama pull mistral
ollama pull codellama
ollama pull phi
ollama pull gemma

# List installed models
ollama list

# Then use any of them in your code - just change the model name!
```

---

## 🎨 Real-World Migration Examples

### Example A: LangChain Application

**Before:**

```python
from langchain.chat_models import ChatOpenAI

llm = ChatOpenAI(
    model="gpt-3.5-turbo",
    temperature=0.7
)

response = llm.invoke("What is machine learning?")
```

**After:**

```python
from langchain.chat_models import ChatOpenAI

llm = ChatOpenAI(
    model="llama2",
    openai_api_base="http://localhost:8080/v1",  # ← Add this
    openai_api_key="sk-llamagate"  # ← Add this
)

response = llm.invoke("What is machine learning?")
```

### Example B: FastAPI Application

**Before:**

```python
from fastapi import FastAPI
from openai import OpenAI

app = FastAPI()
client = OpenAI(api_key="sk-your-key")

@app.post("/chat")
async def chat(message: str):
    response = client.chat.completions.create(
        model="gpt-3.5-turbo",
        messages=[{"role": "user", "content": message}]
    )
    return {"response": response.choices[0].message.content}
```

**After:**

```python
from fastapi import FastAPI
from openai import OpenAI

app = FastAPI()
client = OpenAI(
    base_url="http://localhost:8080/v1",  # ← Change this
    api_key="sk-llamagate"  # ← Change this
)

@app.post("/chat")
async def chat(message: str):
    response = client.chat.completions.create(
        model="llama2",  # ← Change model name
        messages=[{"role": "user", "content": message}]
    )
    return {"response": response.choices[0].message.content}
```

**That's it!** Your FastAPI app now uses local models.

---

## 🎯 What You Get

- ✅ **Zero API costs** - Run models locally
- ✅ **100% private** - Your data never leaves your machine
- ✅ **OpenAI-compatible** - Drop-in replacement
- ✅ **Smart caching** - Identical requests are instant
- ✅ **Any model** - Switch between Ollama models (with loading time on first use)

---

## 🎯 Common Use Cases

### Use Case 1: Development & Testing

**Scenario:** You're building an AI app but don't want to pay for OpenAI API during development.

**Solution:** Use LlamaGate with local models for free development and testing. Switch to OpenAI (or keep using LlamaGate) in production.

### Use Case 2: Privacy-Sensitive Applications

**Scenario:** You're building a healthcare or financial app that can't send data to external APIs.

**Solution:** LlamaGate keeps everything local. Your data never leaves your infrastructure.

### Use Case 3: Cost Optimization

**Scenario:** Your OpenAI API costs are getting high with high usage.

**Solution:** Use LlamaGate with local models for non-critical requests, saving API costs while maintaining functionality.

### Use Case 4: Model Experimentation

**Scenario:** You want to test different models to find the best one for your use case.

**Solution:** Install multiple Ollama models and switch between them instantly with LlamaGate - no code changes needed beyond the model name.

---

## 🚦 Troubleshooting

### Issue: "Connection refused" or "Failed to connect to Ollama"

```bash
# Make sure Ollama is running
ollama serve

# Or check if it's running
curl http://localhost:11434/api/tags
```

### Issue: "Model not found"

```bash
# List installed models
ollama list

# Install the model you need
ollama pull llama2  # or mistral, codellama, etc.
```

### Issue: "401 Unauthorized"

- Check if `API_KEY` is set in your `.env` file
- Include the API key in your requests: `X-API-Key: sk-llamagate`
- Or remove `API_KEY` from `.env` to disable authentication

### Issue: "429 Too Many Requests"

- Adjust `RATE_LIMIT_RPS` in your `.env` file
- Default is 10 requests per second

---

## 📊 Comparison Table

| Feature | OpenAI API | LlamaGate |
| ------- | ---------- | --------- |
| **Cost** | Per-token pricing | Free (local) |
| **Privacy** | Data sent to OpenAI | 100% local |
| **Model Selection** | Limited to OpenAI models | Any Ollama model |
| **API Compatibility** | OpenAI format | ✅ OpenAI-compatible |
| **Rate Limits** | OpenAI's limits | Your own limits |
| **Offline Use** | ❌ Requires internet | ✅ Fully offline |
| **Caching** | ❌ No | ✅ Built-in |
| **Authentication** | Required | Optional |
| **Setup Time** | Minutes | Minutes |

---

## 💡 Pro Tips

1. **Start without authentication** - Remove `API_KEY` from `.env` for quick testing
2. **Use caching** - LlamaGate automatically caches identical requests
3. **Monitor with logs** - Enable `DEBUG=true` to see detailed request logs
4. **Experiment with models** - Different models excel at different tasks
5. **Combine with OpenAI** - Use LlamaGate for some requests, OpenAI for others

---

## 📚 Next Steps

- **Full Documentation:** See [README.md](README.md)
- **Configuration:** Copy `.env.example` to `.env` and customize settings
- **Advanced Usage:** See [docs/](docs/) for detailed guides

---

**Questions?** Check the [README](README.md) or open an issue.
