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

# Final stage
FROM scratch

# Copy the binary from builder
COPY --from=builder /build/llamagate /llamagate

# Expose port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/llamagate"]

