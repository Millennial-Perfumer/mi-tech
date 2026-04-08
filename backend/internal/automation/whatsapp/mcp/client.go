package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
)

// JsonRpcRequest represents a standard MCP request
type JsonRpcRequest struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      interface{} `json:"id"`
}

// JsonRpcResponse represents a standard MCP response
type JsonRpcResponse struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JsonRpcError   `json:"error,omitempty"`
}

type JsonRpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Client struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	
	pending     map[interface{}]chan *JsonRpcResponse
	pendingLock sync.Mutex
	nextID      int
	
	stopChan chan struct{}
}

func NewClient(dir string, env []string) (*Client, error) {
	cmd := exec.Command("node", "shopify_bridge.js")
	cmd.Dir = dir
	cmd.Env = env

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	c := &Client{
		cmd:      cmd,
		stdin:    stdin,
		stdout:   stdout,
		stderr:   stderr,
		pending:  make(map[interface{}]chan *JsonRpcResponse),
		stopChan: make(chan struct{}),
	}

	// Read stderr in background
	go func() {
		scanner := bufio.NewScanner(c.stderr)
		for scanner.Scan() {
			log.Printf("MCP Stderr: %s", scanner.Text())
		}
	}()

	// Read responses in background
	go c.readResponses()

	return c, nil
}

func (c *Client) readResponses() {
	scanner := bufio.NewScanner(c.stdout)
	for scanner.Scan() {
		data := scanner.Bytes()
		var resp JsonRpcResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			log.Printf("MCP Parse Error: %v", err)
			continue
		}

		c.pendingLock.Lock()
		ch, ok := c.pending[resp.ID]
		if ok {
			ch <- &resp
			delete(c.pending, resp.ID)
		}
		c.pendingLock.Unlock()
	}
}

func (c *Client) Call(method string, params interface{}) (*JsonRpcResponse, error) {
	c.pendingLock.Lock()
	id := c.nextID
	c.nextID++
	ch := make(chan *JsonRpcResponse, 1)
	c.pending[id] = ch
	c.pendingLock.Unlock()

	req := JsonRpcRequest{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		ID:      id,
	}

	data, _ := json.Marshal(req)
	fmt.Fprintf(c.stdin, "%s\n", data)

	select {
	case resp := <-ch:
		return resp, nil
	case <-c.stopChan:
		return nil, fmt.Errorf("client stopped")
	}
}

func (c *Client) Close() error {
	close(c.stopChan)
	c.stdin.Close()
	return c.cmd.Process.Kill()
}
