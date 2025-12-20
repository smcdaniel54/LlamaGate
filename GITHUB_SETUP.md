# GitHub Repository Setup Guide

This guide walks you through creating and setting up the LlamaGate GitHub repository.

## Step 1: Prepare Your Local Repository

### 1.1 Initialize Git (if not already done)

```bash
git init
```

### 1.2 Verify .gitignore

Make sure `.gitignore` is properly configured. It should exclude:
- Build artifacts (`llamagate`, `llamagate.exe`)
- Log files (`*.log`)
- Environment files (`.env`)
- IDE files
- OS-specific files

### 1.3 Check What Will Be Committed

```bash
git status
```

You should see your source files, but NOT:
- `llamagate.exe` or `llamagate` (binaries)
- `llamagate.log` (log files)
- `.env` (environment files)

## Step 2: Create Initial Commit

```bash
# Add all files
git add .

# Create initial commit
git commit -m "Initial commit: LlamaGate - OpenAI-compatible proxy for Ollama

- Full MVP implementation with caching, auth, rate limiting
- Cross-platform support (Windows, Linux, macOS)
- Comprehensive documentation and installers
- Unit tests and test scripts"
```

## Step 3: Create GitHub Repository

### Option A: Using GitHub Web Interface

1. **Go to GitHub:**
   - Visit https://github.com
   - Sign in to your account

2. **Create New Repository:**
   - Click the "+" icon in the top right
   - Select "New repository"

3. **Repository Settings:**
   - **Name:** `llamagate` (or your preferred name)
   - **Description:** "OpenAI-compatible HTTP proxy/gateway for local Ollama instances"
   - **Visibility:** Choose Public or Private
   - **Initialize repository:** 
     - ❌ DO NOT check "Add a README file" (you already have one)
     - ❌ DO NOT check "Add .gitignore" (you already have one)
     - ❌ DO NOT check "Choose a license" (we'll add it manually)
   - Click "Create repository"

4. **Copy the repository URL:**
   - You'll see a page with setup instructions
   - Copy the repository URL (e.g., `https://github.com/yourusername/llamagate.git`)

### Option B: Using GitHub CLI

If you have GitHub CLI installed:

```bash
gh repo create llamagate --public --description "OpenAI-compatible HTTP proxy/gateway for local Ollama instances"
```

## Step 4: Connect Local Repository to GitHub

```bash
# Add remote (replace with your repository URL)
git remote add origin https://github.com/yourusername/llamagate.git

# Verify remote
git remote -v
```

## Step 5: Push to GitHub

```bash
# Push to main branch (or master if that's your default)
git branch -M main
git push -u origin main
```

If you're using `master` as your default branch:

```bash
git push -u origin master
```

## Step 6: Set Up Repository Settings

### 6.1 Add Repository Topics

Go to your repository → Settings → Topics, add:
- `go`
- `golang`
- `ollama`
- `openai`
- `proxy`
- `api-gateway`
- `llm`
- `local-llm`

### 6.2 Add Repository Description

Update the description to:
```
OpenAI-compatible HTTP proxy/gateway for local Ollama instances. Features caching, authentication, rate limiting, and structured logging.
```

### 6.3 Enable GitHub Actions

The `.github/workflows/ci.yml` file will automatically set up CI/CD:
- Tests run on push/PR
- Builds on multiple platforms
- Linting checks

### 6.4 Add Repository Badges (Optional)

Add to the top of your README.md:

```markdown
![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey)
```

## Step 7: Create First Release (Optional)

1. Go to your repository → Releases → "Create a new release"
2. **Tag version:** `v0.1.0`
3. **Release title:** `v0.1.0 - Initial Release`
4. **Description:**
   ```markdown
   ## Initial Release
   
   First stable release of LlamaGate with full MVP features.
   
   ### Features
   - OpenAI-compatible API endpoints
   - In-memory caching
   - API key authentication
   - Rate limiting
   - Structured logging
   - Cross-platform support
   - Comprehensive documentation
   ```
5. Click "Publish release"

## Step 8: Verify Everything Works

1. **Check repository:**
   - Visit your GitHub repository
   - Verify all files are present
   - Check that README displays correctly

2. **Test cloning:**
   ```bash
   cd /tmp
   git clone https://github.com/yourusername/llamagate.git
   cd llamagate
   ls -la
   ```

3. **Verify .gitignore:**
   - Make sure no binaries or sensitive files are in the repo

## Step 9: Set Up Branch Protection (Optional, for teams)

If working with a team:

1. Go to Settings → Branches
2. Add branch protection rule for `main`/`master`
3. Enable:
   - Require pull request reviews
   - Require status checks to pass
   - Require branches to be up to date

## Step 10: Add Additional Files (Optional)

Consider adding:

- **CODE_OF_CONDUCT.md** - Community guidelines
- **SECURITY.md** - Security policy
- **CHANGELOG.md** - Version history
- **.github/ISSUE_TEMPLATE/** - Issue templates
- **.github/PULL_REQUEST_TEMPLATE.md** - PR template

## Troubleshooting

### "Repository not found" error
- Check repository URL is correct
- Verify you have push access
- Check authentication (use SSH or HTTPS with token)

### "Permission denied" error
- Set up SSH keys or use HTTPS with personal access token
- For HTTPS: `git remote set-url origin https://YOUR_TOKEN@github.com/username/llamagate.git`

### Large file warnings
- Make sure binaries are in .gitignore
- If committed by mistake: `git rm --cached llamagate.exe`

## Next Steps

After setting up the repository:

1. **Share the repository** with others
2. **Monitor issues** and respond to questions
3. **Accept contributions** via pull requests
4. **Create releases** as you add features
5. **Update documentation** as needed

## Useful Commands Reference

```bash
# Check status
git status

# Add files
git add .

# Commit
git commit -m "Your commit message"

# Push
git push origin main

# Pull latest
git pull origin main

# Create new branch
git checkout -b feature/new-feature

# View remote
git remote -v

# Update remote URL
git remote set-url origin NEW_URL
```

