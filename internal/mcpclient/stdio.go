package mcpclient

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// StdioTransport implements MCP communication over stdio
type StdioTransport struct {
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	stdout   io.ReadCloser
	stderr   io.ReadCloser
	scanner  *bufio.Scanner
	mu       sync.Mutex
	requests map[interface{}]chan *JSONRPCResponse
	nextID   int64
	closed   bool
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(command string, args []string, env map[string]string) (*StdioTransport, error) {
	ctx, cancel := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Env = make([]string, 0, len(env))
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		_ = stdin.Close() // Ignore error - we're already returning an error
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		_ = stdin.Close()   // Ignore error - we're already returning an error
		_ = stdout.Close()   // Ignore error - we're already returning an error
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		_ = stdin.Close()   // Ignore error - we're already returning an error
		_ = stdout.Close()  // Ignore error - we're already returning an error
		_ = stderr.Close()  // Ignore error - we're already returning an error
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	transport := &StdioTransport{
		cmd:      cmd,
		stdin:    stdin,
		stdout:   stdout,
		stderr:   stderr,
		scanner:  bufio.NewScanner(stdout),
		requests: make(map[interface{}]chan *JSONRPCResponse),
		ctx:      ctx,
		cancel:   cancel,
	}

	// Start reading responses
	go transport.readResponses()
	// Start reading stderr for logging
	go transport.readStderr()

	return transport, nil
}

// readResponses reads JSON-RPC responses from stdout
func (t *StdioTransport) readResponses() {
	for t.scanner.Scan() {
		line := t.scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var resp JSONRPCResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			log.Warn().
				Err(err).
				Bytes("line", line).
				Msg("Failed to parse JSON-RPC response")
			continue
		}

		t.mu.Lock()
		ch, ok := t.requests[resp.ID]
		if ok {
			delete(t.requests, resp.ID)
		}
		t.mu.Unlock()

		if ok && ch != nil {
			select {
			case ch <- &resp:
			case <-t.ctx.Done():
				return
			}
		}
	}

	// Scanner stopped, mark as closed
	t.mu.Lock()
	t.closed = true
	// Close all pending request channels
	for id, ch := range t.requests {
		close(ch)
		delete(t.requests, id)
	}
	t.mu.Unlock()
}

// readStderr reads from stderr and logs it
func (t *StdioTransport) readStderr() {
	scanner := bufio.NewScanner(t.stderr)
	for scanner.Scan() {
		line := scanner.Text()
		log.Debug().
			Str("stderr", line).
			Msg("MCP server stderr")
	}
}

// SendRequest sends a JSON-RPC request and waits for a response
func (t *StdioTransport) SendRequest(ctx context.Context, method string, params interface{}) (*JSONRPCResponse, error) {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return nil, ErrConnectionClosed
	}

	id := t.nextID
	t.nextID++
	t.mu.Unlock()

	var paramsJSON json.RawMessage
	if params != nil {
		var err error
		paramsJSON, err = json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
	}

	req := JSONRPCRequest{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Method:  method,
		Params:  paramsJSON,
	}

	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create response channel
	ch := make(chan *JSONRPCResponse, 1)

	t.mu.Lock()
	t.requests[id] = ch
	t.mu.Unlock()

	// Write request
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return nil, ErrConnectionClosed
	}
	_, err = t.stdin.Write(append(reqJSON, '\n'))
	t.mu.Unlock()

	if err != nil {
		t.mu.Lock()
		delete(t.requests, id)
		t.mu.Unlock()
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// Wait for response
	select {
	case resp := <-ch:
		if resp == nil {
			return nil, ErrConnectionClosed
		}
		if resp.Error != nil {
			return nil, resp.Error
		}
		return resp, nil
	case <-ctx.Done():
		t.mu.Lock()
		delete(t.requests, id)
		t.mu.Unlock()
		return nil, fmt.Errorf("request timeout: %w", ctx.Err())
	case <-t.ctx.Done():
		return nil, ErrConnectionClosed
	}
}

// Close closes the transport
func (t *StdioTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true
	t.cancel()

	// Close stdin to signal shutdown
	if t.stdin != nil {
		_ = t.stdin.Close() // Ignore error - we're shutting down
	}

	// Wait for process to exit with timeout
	done := make(chan error, 1)
	go func() {
		done <- t.cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			log.Warn().Err(err).Msg("MCP server process exited with error")
		}
	case <-time.After(5 * time.Second):
		log.Warn().Msg("MCP server process did not exit, killing")
		if killErr := t.cmd.Process.Kill(); killErr != nil {
			log.Warn().Err(killErr).Msg("Failed to kill MCP server process")
		}
		<-done
	}

	// Close pipes
	if t.stdout != nil {
		_ = t.stdout.Close() // Ignore error - we're shutting down
	}
	if t.stderr != nil {
		_ = t.stderr.Close() // Ignore error - we're shutting down
	}

	return nil
}

// IsClosed returns whether the transport is closed
func (t *StdioTransport) IsClosed() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.closed
}
