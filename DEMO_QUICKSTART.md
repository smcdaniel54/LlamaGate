# 🚀 LlamaGate Demo & Quick-Start Showcase

**Status:** ⚙️ Included in LlamaGate docs (not sold separately)

## What This Guide Does

- ✅ Shows how to swap a ChatGPT/OpenAI app to LlamaGate in **under 5 minutes**
- ✅ Demonstrates seamless model switching between different LLMs
- ✅ Helps developers onboard and start using local models immediately
- ✅ Reduces friction, increases trust, and improves adoption

---

## 🎯 The Big Picture: Why LlamaGate?

**Before LlamaGate:**
- Locked into OpenAI's API pricing and rate limits
- No control over your data or models
- Dependent on external services
- Limited model selection

**After LlamaGate:**
- ✅ **100% local** - Run models on your own infrastructure
- ✅ **Zero API costs** - No per-token charges
- ✅ **Complete privacy** - Your data never leaves your machine
- ✅ **Model flexibility** - Switch between any Ollama model instantly
- ✅ **OpenAI-compatible** - Drop-in replacement, no code rewrites needed

---

## ⚡ 5-Minute Quick Start

### Step 1: Install LlamaGate (2 minutes)

**Windows:**
```cmd
install\windows\install.cmd
```

**Unix/Linux/macOS:**
```bash
chmod +x install.sh
./install.sh
```

The installer will:
- Check and install Go if needed
- Check and guide Ollama installation
- Build LlamaGate
- Create configuration file

### Step 2: Start LlamaGate (30 seconds)

**Windows:**
```cmd
run.cmd
```

**Unix/Linux/macOS:**
```bash
./run.sh
```

You should see:
```
{"level":"info","message":"Starting LlamaGate","ollama_host":"http://localhost:11434","port":"8080"}
{"level":"info","message":"Server starting","address":":8080"}
```

### Step 3: Verify It Works (30 seconds)

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{"status":"healthy"}
```

### Step 4: List Available Models (30 seconds)

```bash
curl http://localhost:8080/v1/models
```

Expected response:
```json
{
  "object": "list",
  "data": [
    {
      "id": "llama2",
      "object": "model",
      "created": 0,
      "owned_by": "ollama"
    }
  ]
}
```

**🎉 You're ready!** Now let's migrate your existing OpenAI app.

---

## 🔄 Migration Guide: From OpenAI to LlamaGate

### The Magic: It's Just One Line Change!

LlamaGate is a **drop-in replacement** for OpenAI's API. You only need to change the `base_url` (or `api_base`) in your code.

---

## 📝 Example 1: Python with OpenAI SDK

### Before (Using OpenAI)

```python
from openai import OpenAI

# Original OpenAI client
client = OpenAI(
    api_key="sk-your-openai-key-here"  # Requires paid API key
)

# Make a request
response = client.chat.completions.create(
    model="gpt-3.5-turbo",  # OpenAI model
    messages=[
        {"role": "user", "content": "Hello! Explain quantum computing in one sentence."}
    ]
)

print(response.choices[0].message.content)
```

### After (Using LlamaGate) ✨

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

**That's it!** Your code works exactly the same. No other changes needed.

---

## 📝 Example 2: JavaScript/Node.js with OpenAI SDK

### Before (Using OpenAI)

```javascript
import OpenAI from 'openai';

const openai = new OpenAI({
  apiKey: 'sk-your-openai-key-here',  // Requires paid API key
});

const response = await openai.chat.completions.create({
  model: 'gpt-3.5-turbo',
  messages: [
    { role: 'user', content: 'Write a haiku about coding' }
  ],
});

console.log(response.choices[0].message.content);
```

### After (Using LlamaGate) ✨

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

---

## 📝 Example 3: cURL (Direct API Calls)

### Before (OpenAI API)

```bash
curl https://api.openai.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-your-openai-key" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### After (LlamaGate) ✨

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

## 🔀 Model Switching Demo

One of LlamaGate's superpowers is **instant model switching**. You can switch between different models with zero code changes - just change the model name!

### Example: Switching Between Models

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
print(f"Llama 2: {response1.choices[0].message.content}\n")

# Switch to Mistral (if you have it installed)
response2 = client.chat.completions.create(
    model="mistral",  # ← Just change the model name!
    messages=[{"role": "user", "content": "Explain AI in simple terms"}]
)
print(f"Mistral: {response2.choices[0].message.content}\n")

# Switch to CodeLlama
response3 = client.chat.completions.create(
    model="codellama",  # ← Another model, same code!
    messages=[{"role": "user", "content": "Write a Python function to sort a list"}]
)
print(f"CodeLlama: {response3.choices[0].message.content}\n")
```

### Installing More Models

To use different models, just install them with Ollama:

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

## 🎬 Video Demo Script Outline

> **Note:** This section provides a script outline for creating a video demonstration. The actual video should be created separately.

### Video Structure (5-7 minutes)

#### Segment 1: Introduction (30 seconds)
- Show a working ChatGPT/OpenAI application
- Highlight the cost and dependency concerns
- Introduce LlamaGate as the solution

#### Segment 2: Installation (1 minute)
- Show the installer running
- Demonstrate the automated setup
- Verify LlamaGate is running

#### Segment 3: The Migration (2 minutes)
- Show the "before" code (OpenAI)
- Show the "after" code (LlamaGate)
- Highlight: **Only one line changed!**
- Run both versions side-by-side
- Show identical results

#### Segment 4: Model Switching (1.5 minutes)
- Show listing available models
- Demonstrate switching between 3-4 different models
- Show how easy it is - just change the model name
- Compare outputs from different models

#### Segment 5: Benefits Recap (1 minute)
- Zero API costs
- Complete privacy
- Model flexibility
- Production-ready features (caching, rate limiting, auth)

#### Segment 6: Call to Action (30 seconds)
- "Get started in 5 minutes"
- Link to documentation
- GitHub repository

### Key Visual Elements
- Split screen showing before/after code
- Terminal showing LlamaGate running
- Code editor with highlighted changes
- Side-by-side API responses
- Model switching demonstration

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

## 🚦 Troubleshooting Quick Reference

### Issue: "Connection refused" or "Failed to connect to Ollama"

**Solution:**
```bash
# Make sure Ollama is running
ollama serve

# Or check if it's running
curl http://localhost:11434/api/tags
```

### Issue: "Model not found"

**Solution:**
```bash
# List installed models
ollama list

# Install the model you need
ollama pull llama2  # or mistral, codellama, etc.
```

### Issue: "401 Unauthorized"

**Solution:**
- Check if `API_KEY` is set in your `.env` file
- Include the API key in your requests: `X-API-Key: sk-llamagate`
- Or remove `API_KEY` from `.env` to disable authentication

### Issue: "429 Too Many Requests"

**Solution:**
- Adjust `RATE_LIMIT_RPS` in your `.env` file
- Default is 10 requests per second

---

## 📊 Comparison Table

| Feature | OpenAI API | LlamaGate |
|---------|-----------|-----------|
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

## 🎓 Next Steps

1. **Try the examples above** - Copy and paste them into your environment
2. **Install more models** - Explore different Ollama models
3. **Read the full documentation** - See [README.md](README.md) for advanced features
4. **Join the community** - Contribute, report issues, or ask questions

---

## 💡 Pro Tips

1. **Start without authentication** - Remove `API_KEY` from `.env` for quick testing
2. **Use caching** - LlamaGate automatically caches identical requests
3. **Monitor with logs** - Enable `DEBUG=true` to see detailed request logs
4. **Experiment with models** - Different models excel at different tasks
5. **Combine with OpenAI** - Use LlamaGate for some requests, OpenAI for others

---

## ✅ Success Checklist

After following this guide, you should be able to:

- [ ] Install and run LlamaGate
- [ ] Migrate an existing OpenAI app to LlamaGate
- [ ] Switch between different models
- [ ] Understand the benefits of using local models
- [ ] Troubleshoot common issues

---

## 🎉 You're All Set!

You now know how to:
- ✅ Swap any ChatGPT/OpenAI app to LlamaGate in minutes
- ✅ Switch between models effortlessly
- ✅ Run AI applications locally with zero API costs
- ✅ Maintain complete privacy and control

**Ready to build?** Start with the examples above and customize them for your needs!

---

## 📚 Additional Resources

- [Full Documentation](README.md) - Complete feature reference
- [Installation Guide](INSTALL.md) - Detailed installation instructions
- [Testing Guide](TESTING.md) - How to test your setup
- [Project Structure](STRUCTURE.md) - Understanding the codebase

---

**Questions? Issues?** Check the troubleshooting section above or refer to the main documentation.

**Happy coding! 🚀**

