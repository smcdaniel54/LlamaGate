# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o llamagate ./cmd/llamagate

# Final stage: minimal image with only the binary (no shell; config via env)
FROM scratch

# Copy the binary from builder
COPY --from=builder /build/llamagate /llamagate

# LlamaGate listens on 11435 by default (override with PORT)
EXPOSE 11435

# Configure at runtime with env vars, e.g.:
#   OLLAMA_HOST   - Ollama URL (default http://localhost:11434); use host.docker.internal or service name if Ollama runs elsewhere
#   PORT          - Listen port (default 11435)
#   API_KEY       - Optional API key for auth (empty = no auth)
#   MCP_ENABLED   - Set true if using MCP (default false)
ENTRYPOINT ["/llamagate"]

