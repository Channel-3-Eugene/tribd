package uriHandler

import (
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func tryBindWithRetry(address string, maxRetries int) (net.Listener, error) {
	var listener net.Listener
	var err error
	for i := 0; i < maxRetries; i++ {
		listener, err = net.Listen("tcp", address)
		if err == nil {
			return listener, nil
		}
		if !strings.Contains(err.Error(), "address already in use") {
			return nil, err
		}
		time.Sleep(time.Duration(i+1) * 100 * time.Millisecond) // Exponential back-off could be considered
	}
	return nil, err
}

func startEchoServer(t *testing.T) (string, func()) {
	listener, err := tryBindWithRetry("localhost:0", 5)
	assert.NoError(t, err, "Failed to start echo server after retries")

	var wg sync.WaitGroup

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				break // Exit the loop on server close
			}
			wg.Add(1)
			go func(c net.Conn) {
				defer wg.Done()
				defer c.Close()
				io.Copy(c, c) // Simple echo server
			}(conn)
		}
	}()

	return listener.Addr().String(), func() {
		listener.Close()
		wg.Wait() // Ensure all connections are closed
	}
}

// Helper function to start the TCPHandler in server mode.
func startServerHandler(t *testing.T, address string, dataChan chan []byte) *TCPHandler {
	handler := NewTCPHandler(address, 1*time.Second, 1*time.Second, Server, Reader, dataChan)
	assert.NoError(t, handler.Open(), "Failed to open server handler")
	return handler
}

// TestTCPHandlerServerMode verifies the server functionality of TCPHandler to ensure it can receive data.
func TestTCPHandlerServerMode(t *testing.T) {
	// Setup a channel to capture data received by the server.
	dataChan := make(chan []byte, 1)
	defer close(dataChan)

	// Create server handler.
	serverAddr, cleanupServer := startEchoServer(t) // Ensure this function starts a simple echo server correctly
	defer cleanupServer()

	handler := startServerHandler(t, serverAddr, dataChan)
	defer handler.Close()

	// Start a client to send data to the server.
	conn, err := net.Dial("tcp", serverAddr)
	assert.NoError(t, err, "Failed to connect to server")
	defer conn.Close()

	// Send data to the server.
	sentMessage := "hello from client"
	_, err = conn.Write([]byte(sentMessage))
	assert.NoError(t, err, "Failed to send data to server")

	// Read and verify the data received by the server.
	receivedMessage := <-dataChan
	assert.Equal(t, sentMessage, string(receivedMessage), "Mismatch in data received by server")
}

func TestTCPHandlerClientMode(t *testing.T) {
	serverAddr, cleanupServer := startEchoServer(t)
	defer cleanupServer()

	handler := NewTCPHandler(serverAddr, 1*time.Second, 1*time.Second, Client, Writer, make(chan []byte, 1))
	assert.NoError(t, handler.Open(), "Failed to open TCP handler")
	defer handler.Close()

	// Additional testing logic to send/receive data to/from handler
}
