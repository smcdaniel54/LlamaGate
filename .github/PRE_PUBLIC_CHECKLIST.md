# Pre-Public Release Checklist

This checklist ensures LlamaGate is ready to go public safely and professionally.

## ✅ 1. Leak Check (MOST IMPORTANT)

### A. Secrets & Credentials
- [x] **`.env` files**: Real `.env` is in `.gitignore`, only `.env.example` exists
- [x] **API keys/tokens**: Only example keys found (e.g., `sk-llamagate`), no real secrets
- [x] **Private keys/certs**: No `.key`, `.pem`, `.p12` files in repo
- [x] **Database wallets**: None found
- [x] **Installer scripts**: No embedded credentials
- [x] **Git history**: No secrets found in commit history

**Action taken**: Ran `git grep` for common secret patterns - all results are examples/placeholders.

### B. Git History
- [x] Checked git history for sensitive files - clean
- [x] No secrets found in past commits

**Status**: ✅ **CLEAN** - No secrets detected

---

## ✅ 2. Public Posture & Documentation

### A. License
- [x] **LICENSE** file exists (MIT License)
- [x] License is appropriate for open-source core with paid add-ons

### B. Essential Documentation
- [x] **README.md** - Comprehensive with Quick Start
- [x] **LICENSE** - MIT License
- [x] **SECURITY.md** - Security policy and reporting process
- [x] **CONTRIBUTING.md** - Contribution guidelines
- [x] **CODE_OF_CONDUCT.md** - Community standards (just created)
- [x] **CHANGELOG.md** - Version history
- [x] **.env.example** - Configuration template (exists)

### C. Paid Tier Boundary
- [x] **README updated** with "Project Scope & Paid Tier Boundary" section
- [x] Clear separation between open-source core and paid modules

**Status**: ✅ **COMPLETE**

---

## ✅ 3. Repository Polish

### A. Quick Start
- [x] Quick Start guide exists (QUICKSTART.md)
- [x] Installation instructions are clear
- [x] Example code provided

**Action needed**: Manually verify Quick Start works end-to-end on a fresh machine

### B. Known Limitations
- [x] **README updated** with "Known Limitations" section
- [x] Lists supported platforms, backends, MCP status
- [x] Clearly states what's NOT included

### C. Issue Templates
- [x] Bug report template exists (`.github/ISSUE_TEMPLATE/bug_report.md`)
- [x] Feature request template exists (`.github/ISSUE_TEMPLATE/feature_request.md`)
- [x] Config file exists (`.github/ISSUE_TEMPLATE/config.yml`)

**Status**: ✅ **COMPLETE**

---

## ⚠️ 4. GitHub Repository Settings (Manual Steps Required)

### A. Branch Protection
**Action required**: Enable in GitHub → Settings → Branches

- [ ] Require pull request reviews before merging
- [ ] Require status checks to pass before merging
  - [ ] Require CI workflow to pass
- [ ] Require branches to be up to date before merging
- [ ] Do not allow bypassing the above settings
- [ ] Restrict who can push to matching branches (if applicable)

### B. Security Features
**Action required**: Enable in GitHub → Settings → Security

- [ ] Enable Dependabot alerts
- [ ] Enable Dependabot security updates
- [ ] Enable Secret scanning (if available on your plan)
- [ ] Enable Code scanning (optional, if available)

### C. Repository Features
**Action required**: Review in GitHub → Settings → General

- [ ] Disable Wiki (if you won't maintain it)
- [ ] Choose merge strategy (squash, merge, or rebase)
- [ ] Enable/disable auto-delete branches after merge
- [ ] Set default branch to `main`

**Status**: ⚠️ **MANUAL ACTION REQUIRED**

---

## ✅ 5. CI/CD & Release Automation

### A. GitHub Actions Workflows
- [x] **CI workflow** (`.github/workflows/ci.yml`) - Clean, no secrets exposed
- [x] **Release workflow** (`.github/workflows/release.yml`) - Uses `GITHUB_TOKEN` only
- [x] No hardcoded secrets in workflows
- [x] Workflows use encrypted secrets where needed

**Status**: ✅ **SECURE**

### B. Versioning
- [x] **CHANGELOG.md** updated to 0.9.0
- [x] Version references updated throughout codebase
- [x] Ready for first public release tag: `v0.9.0`

**Status**: ✅ **READY**

---

## ⚠️ 6. Pre-Flip Verification (Do Before Going Public)

### A. Test Quick Start
- [ ] Test on a fresh machine/VM:
  - [ ] Windows installation
  - [ ] Linux installation
  - [ ] macOS installation (if available)
- [ ] Verify all installation methods work
- [ ] Verify "Hello World" example works

### B. Documentation Review
- [ ] Open README in incognito window (not logged in)
- [ ] Verify all links work
- [ ] Check that no internal-only docs are visible
- [ ] Verify Quick Start is understandable

**Status**: ⚠️ **MANUAL TESTING REQUIRED**

---

## 🚀 7. Going Public (Final Steps)

### A. Flip Repository Visibility
1. Go to GitHub → Repository Settings
2. Scroll to "Danger Zone"
3. Click "Change visibility"
4. Select "Public"
5. Confirm by typing repository name

### B. First Public Release
1. Create a release tag: `git tag v0.9.0`
2. Push tag: `git push origin v0.9.0`
3. GitHub Actions will automatically build and create release
4. Review the release page to ensure binaries are attached

### C. Post-Public Checklist
- [ ] Verify repository is accessible in incognito window
- [ ] Check that all links work
- [ ] Verify release page looks correct
- [ ] Test downloading a binary from releases
- [ ] (Optional) Announce on social media/forums

**Status**: ⚠️ **READY TO EXECUTE**

---

## 📋 Summary

### ✅ Completed
- [x] Leak check (no secrets found)
- [x] All essential documentation files
- [x] CODE_OF_CONDUCT.md created
- [x] README updated with paid tier boundary
- [x] README updated with known limitations
- [x] CI/CD workflows verified secure
- [x] Version updated to 0.9.0

### ⚠️ Manual Actions Required
- [ ] Enable branch protection rules
- [ ] Enable security features (Dependabot, secret scanning)
- [ ] Test Quick Start on fresh machines
- [ ] Review repository settings
- [ ] Test documentation in incognito mode

### 🚀 Ready to Go Public
Once manual actions are complete, you're ready to flip the repository to public!

---

## 🔒 Security Reminder

**If you ever committed a secret (even if deleted later):**
1. **Rotate the secret immediately** (treat as compromised)
2. Consider rewriting git history (advanced, use `git filter-branch` or BFG Repo-Cleaner)
3. Review GitHub's secret scanning alerts after going public

---

**Last Updated**: 2026-01-05
**Version**: 0.9.0-pre-release

