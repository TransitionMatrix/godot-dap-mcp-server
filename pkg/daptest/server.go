package daptest

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-dap"
)

// MockServer is a programmable DAP server for testing
type MockServer struct {
	t        *testing.T
	listener net.Listener
	addr     string
	conn     net.Conn
	mu       sync.Mutex
	closed   bool
	seq      int

	// Channels for coordination
	receivedMsgs chan dap.Message
	errors       chan error
}

// NewServer starts a new mock DAP server on a random port
func NewServer(t *testing.T) *MockServer {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}

	s := &MockServer{
		t:            t,
		listener:     listener,
		addr:         listener.Addr().String(),
		receivedMsgs: make(chan dap.Message, 100),
		errors:       make(chan error, 10),
	}

	go s.acceptLoop()

	return s
}

// Address returns the address the server is listening on
func (s *MockServer) Address() string {
	return s.addr
}

// Port returns the port number
func (s *MockServer) Port() int {
	addr := s.listener.Addr().(*net.TCPAddr)
	return addr.Port
}

// Close shuts down the server
func (s *MockServer) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	if s.conn != nil {
		s.conn.Close()
	}
	s.listener.Close()
}

// NextSeq returns the next sequence number
func (s *MockServer) NextSeq() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	return s.seq
}

func (s *MockServer) acceptLoop() {
	conn, err := s.listener.Accept()
	if err != nil {
		s.mu.Lock()
		if !s.closed {
			s.errors <- fmt.Errorf("accept error: %v", err)
		}
		s.mu.Unlock()
		return
	}

	s.mu.Lock()
	s.conn = conn
	s.mu.Unlock()

	go s.readLoop(conn)
}

func (s *MockServer) readLoop(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		// Read Content-Length header
		contentLength := -1
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					s.errors <- fmt.Errorf("read error: %v", err)
				}
				return
			}
			line = strings.TrimSpace(line)
			if line == "" {
				break // End of headers
			}
			if strings.HasPrefix(line, "Content-Length: ") {
				cl, err := strconv.Atoi(strings.TrimPrefix(line, "Content-Length: "))
				if err != nil {
					s.errors <- fmt.Errorf("invalid content length: %v", err)
					return
				}
				contentLength = cl
			}
		}

		if contentLength == -1 {
			s.errors <- fmt.Errorf("missing Content-Length header")
			return
		}

		// Read body
		body := make([]byte, contentLength)
		_, err := io.ReadFull(reader, body)
		if err != nil {
			s.errors <- fmt.Errorf("read body error: %v", err)
			return
		}

		// Decode message to determine type
		var raw map[string]interface{}
		if err := json.Unmarshal(body, &raw); err != nil {
			s.errors <- fmt.Errorf("json unmarshal error: %v", err)
			continue
		}

		// Decode into dap.Message using go-dap
		// To simulate ReadProtocolMessage's input, we need headers + body.
		fullMsg := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), string(body))
		decodedMsg, err := dap.ReadProtocolMessage(bufio.NewReader(strings.NewReader(fullMsg)))
		if err != nil {
			s.errors <- fmt.Errorf("dap decode error: %v", err)
			continue
		}

		s.receivedMsgs <- decodedMsg
	}
}

// Send sends a message to the connected client
func (s *MockServer) Send(msg dap.Message) error {
	s.mu.Lock()
	conn := s.conn
	s.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("no client connected")
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	_, err = conn.Write([]byte(header))
	if err != nil {
		return err
	}
	_, err = conn.Write(body)
	return err
}

// ExpectRequest waits for a request of a specific command type
func (s *MockServer) ExpectRequest(command string) (dap.Message, error) {
	select {
	case msg := <-s.receivedMsgs:
		// Extract Type using reflection
		val := reflect.ValueOf(msg)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		msgType := ""
		typeField := val.FieldByName("Type")
		if typeField.IsValid() && typeField.Kind() == reflect.String {
			msgType = typeField.String()
		} else {
			// Try embedded ProtocolMessage
			pmField := val.FieldByName("ProtocolMessage")
			if pmField.IsValid() {
				typeField = pmField.FieldByName("Type")
				if typeField.IsValid() && typeField.Kind() == reflect.String {
					msgType = typeField.String()
				}
			}
		}

		// Check if it's a request
		if msgType != "request" {
			return nil, fmt.Errorf("expected request, got %s (type: %T)", msgType, msg)
		}

		// Extract command using reflection since go-dap returns specific structs
		// (e.g. *dap.InitializeRequest) which don't share a common Command interface
		// val is already the value

		cmdField := val.FieldByName("Command")
		if !cmdField.IsValid() || cmdField.Kind() != reflect.String {
			// Try embedded Request struct
			reqField := val.FieldByName("Request")
			if reqField.IsValid() {
				cmdField = reqField.FieldByName("Command")
			}
		}

		actualCmd := ""
		if cmdField.IsValid() && cmdField.Kind() == reflect.String {
			actualCmd = cmdField.String()
		}

		if actualCmd == command {
			return msg, nil
		}
		return nil, fmt.Errorf("expected command %s, got %s (type: %T)", command, actualCmd, msg)

	case err := <-s.errors:
		return nil, err
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout waiting for request %s", command)
	}
}
