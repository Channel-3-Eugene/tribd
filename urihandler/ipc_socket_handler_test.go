package uriHandler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewSocketHandler checks the initialization of a new SocketHandler to ensure all fields are set as expected.
func TestNewSocketHandler(t *testing.T) {
	dataChan := make(chan []byte)
	socketPath := randSocketPath()
	handler := NewSocketHandler(socketPath, 0, 0, Server, Reader, dataChan)
	assert.Equal(t, socketPath, handler.socketPath)
	assert.Equal(t, 0*time.Second, handler.readDeadline)
	assert.Equal(t, 0*time.Second, handler.writeDeadline)
	assert.Equal(t, Server, handler.mode)
	assert.Equal(t, Reader, handler.role)
	assert.Equal(t, dataChan, handler.dataChan)
	assert.NotNil(t, handler.connections)
}

// TestSocketServerWriterClientReader tests the interaction between a server set to write and a client set to read.
func TestSocketServerWriterClientReader(t *testing.T) {
	writerChan := make(chan []byte)
	readerChan := make(chan []byte)

	randomSocketPath := randSocketPath()

	// Initialize server to write data.
	serverWriter := NewSocketHandler(randomSocketPath, 0, 0, Server, Writer, writerChan)
	serverWriter.Open()
	// Ensure the server is ready.
	for {
		if serverWriter.listener != nil {
			break
		}
		time.Sleep(100 * time.Microsecond)
	}

	// Initialize client to read data.
	clientReader := NewSocketHandler(randomSocketPath, 10*time.Millisecond, 10*time.Millisecond, Client, Reader, readerChan)
	clientReader.Open()

	t.Run("TestNewSocketHandler", func(t *testing.T) {
		status := serverWriter.Status()
		assert.Equal(t, serverWriter.socketPath, status.Address)
		assert.Equal(t, Server, status.Mode)
		assert.Equal(t, Writer, status.Role)
	})

	t.Run("TestWriteData", func(t *testing.T) {
		randBytes := make([]byte, 188)
		_, err := rand.Read(randBytes)
		if err != nil {
			t.Fatal("Failed to generate random bytes:", err)
		}
		fmt.Println("Sending data...")
		writerChan <- randBytes
		fmt.Println("Sent data")

		select {
		case data := <-readerChan:
			fmt.Println("Data received.")
			assert.Equal(t, randBytes, data)
		case <-time.After(100 * time.Millisecond):
			t.Error("Timeout waiting for data")
		}
	})
}

// TestSocketServerReaderClientWriter tests the interaction between a server set to read and a client set to write.
func TestSocketServerReaderClientWriter(t *testing.T) {
	writerChan := make(chan []byte)
	readerChan := make(chan []byte)

	// Initialize server to read data.
	serverReader := NewSocketHandler(randSocketPath(), 0, 0, Server, Reader, readerChan)
	serverReader.Open()
	// Ensure the server is ready.
	for {
		if serverReader.listener != nil {
			break
		}
		time.Sleep(100 * time.Microsecond)
	}

	// Initialize client to write data.
	clientWriter := NewSocketHandler(serverReader.socketPath, 10*time.Millisecond, 10*time.Millisecond, Client, Writer, writerChan)
	clientWriter.Open()

	t.Run("TestNewSocketHandler", func(t *testing.T) {
		status := serverReader.Status()
		assert.Equal(t, Server, status.Mode)
		assert.Equal(t, Reader, status.Role)

		status = clientWriter.Status()
		assert.Equal(t, serverReader.socketPath, status.Address)
		assert.Equal(t, Client, status.Mode)
		assert.Equal(t, Writer, status.Role)
	})

	t.Run("TestWriteData", func(t *testing.T) {
		randBytes := make([]byte, 188)
		_, _ = rand.Read(randBytes)
		writerChan <- randBytes

		select {
		case data := <-readerChan:
			assert.Equal(t, randBytes, data)
		case <-time.After(5 * time.Millisecond):
			assert.Fail(t, "Timeout waiting for data")
		}
	})
}

// randSocketPath generates a random path for a Unix socket used in testing.
func randSocketPath() string {
	randBytes := make([]byte, 8)
	_, err := rand.Read(randBytes)
	if err != nil {
		panic(err) // It's better to handle the error properly in real applications.
	}
	return "/tmp/" + hex.EncodeToString(randBytes) + ".sock"
}
