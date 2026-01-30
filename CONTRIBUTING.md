# Contributing to LlamaGate

Thank you for your interest in contributing to LlamaGate! This document provides guidelines and instructions for contributing.

## How to Contribute

### Reporting Bugs

If you find a bug, please open an issue with:
- A clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Your environment (OS, Go version, etc.)

### Suggesting Features

Feature suggestions are welcome! Please open an issue describing:
- The feature and its use case
- How it would benefit users
- Any implementation ideas (optional)

### Pull Requests

1. **Fork the repository**
2. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **Make your changes:**
   - Follow Go code style guidelines
   - Add tests for new features
   - Update documentation as needed
4. **Test your changes:**
   ```bash
   go build ./...   # Must succeed for downstream CI, E2E, and build-from-source tooling
   go test ./...
   go build -o llamagate ./cmd/llamagate
   ```
   For concurrent code, also run race detector tests (matches CI):
   - **Windows:** `.\scripts\windows\test-race.ps1`
   - **Unix/Linux/macOS:** `./scripts/unix/test-race.sh`
   - **Manual:** `CGO_ENABLED=1 go test -race -timeout=10m ./...`
   
   **Note:** Race detector tests require CGO_ENABLED=1 and match the CI configuration exactly.
5. **Commit your changes:**
   ```bash
   git commit -m "Add: description of your changes"
   ```
6. **Push to your fork:**
   ```bash
   git push origin feature/your-feature-name
   ```
7. **Open a Pull Request** with a clear description

## Code Style

- Use valid Go string literals: double-quoted strings `"..."` only need `\"` *inside* the string for a literal quote. Do not escape the delimiters in source (e.g. avoid `\"...\"` as the whole literal), so that `go build ./...` succeeds and downstream build-from-source (CI, E2E, forked automation) is not broken.
- Follow standard Go formatting (`go fmt`)
- Use `golangci-lint` v2.8.0 for linting:
  - **Windows:** `.\scripts\windows\install-golangci-lint.ps1` then `.\scripts\windows\lint.ps1`
  - **Unix/Linux/macOS:** `./scripts/unix/lint.sh`
- **Pre-commit hook:** Automatically runs linting on staged files before each commit
  - **Windows:** `.\scripts\windows\setup-pre-commit.ps1` (one-time setup)
  - **Unix/Linux/macOS:** Pre-commit hook is automatically created
  - To skip: `git commit --no-verify`
- Write clear, self-documenting code
- Add comments for exported functions
- Keep functions small and focused

### CI vs Local Linting

- **Local:** Full linting including test files (strict) - enforced by pre-commit hook
- **CI:** Production code only (faster, `tests: false` in `.golangci.yml`)
- **Why:** CI focuses on production code quality while maintaining fast feedback. Pre-commit hook ensures developers fix all issues locally before pushing.

## Testing

- Add unit tests for new features
- Ensure all tests pass: `go test ./...`
- Test on multiple platforms if possible

## Documentation

- Update README.md for user-facing changes
- Add code comments for complex logic
- Update relevant documentation files

## Questions?

Feel free to open an issue for questions or discussions!

