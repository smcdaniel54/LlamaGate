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
   go test ./...
   go build ./cmd/llamagate
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

- Follow standard Go formatting (`go fmt`)
- Use `golangci-lint` v2.8.0 for linting (matches CI):
  - **Windows:** `.\scripts\windows\install-golangci-lint.ps1` then `golangci-lint run`
  - **Unix/Linux/macOS:** `./scripts/unix/lint.sh`
- Write clear, self-documenting code
- Add comments for exported functions
- Keep functions small and focused

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

