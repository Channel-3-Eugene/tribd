package uriHandler

import (
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/Channel-3-Eugene/tribd/channels"
	"github.com/stretchr/testify/assert"
)

// TestNewTCPHandler tests the creation of a new TCPHandler instance.
func TestNewTCPHandler(t *testing.T) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		fmt.Printf("TestNewTCPHandler took %v\n", duration)
	}()

	dataChan := channels.NewPacketChan(1)
	handler := NewTCPHandler(":0", 0, 0, Server, Reader)
	handler.dataChan = dataChan
	assert.Equal(t, ":0", handler.address)
	assert.Equal(t, 0*time.Second, handler.readDeadline)
	assert.Equal(t, 0*time.Second, handler.writeDeadline)
	assert.Equal(t, Server, handler.mode)
	assert.Equal(t, Reader, handler.role)
	assert.Equal(t, dataChan, handler.dataChan)
	assert.NotNil(t, handler.connections)
}

// TestTCPServerWriterClientReader tests the TCPHandler functionality with a server in writer role and a client in reader role.
func TestTCPServerWriterClientReader(t *testing.T) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		fmt.Printf("TestTCPServerWriterClientReader took %v\n", duration)
	}()

	writerChan := channels.NewPacketChan(1)
	readerChan := channels.NewPacketChan(1)

	serverWriter := NewTCPHandler(":0", 0, 0, Server, Writer)
	serverWriter.dataChan = writerChan
	err := serverWriter.Open()
	assert.Nil(t, err)

	serverWriterAddr := serverWriter.Status().Address

	clientReader := NewTCPHandler(serverWriterAddr, 0, 0, Client, Reader)
	clientReader.dataChan = readerChan
	err = clientReader.Open()
	assert.Nil(t, err)

	t.Run("TestNewTCPHandler", func(t *testing.T) {
		status := serverWriter.Status()
		assert.Equal(t, serverWriterAddr, status.Address)
		assert.Equal(t, Server, status.Mode)
		assert.Equal(t, Writer, status.Role)
	})

	t.Run("TestWriteData", func(t *testing.T) {
		randBytes := make([]byte, 188)
		_, _ = rand.Read(randBytes)
		writerChan.Send(randBytes)

		select {
		case <-time.After(5 * time.Millisecond):
			assert.Fail(t, "Timeout waiting for data")
		default:
			data := readerChan.Receive()
			assert.Equal(t, randBytes, data)
		}

		status := serverWriter.Status()
		assert.Equal(t, 1, len(status.Connections))
	})
}

// TestTCPServerReaderClientWriter tests the TCPHandler functionality with a server in reader role and a client in writer role.
func TestTCPServerReaderClientWriter(t *testing.T) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		fmt.Printf("TestTCPServerReaderClientWriter took %v\n", duration)
	}()

	writerChan := channels.NewPacketChan(1)
	readerChan := channels.NewPacketChan(1)

	serverReader := NewTCPHandler(":0", 0, 0, Server, Reader)
	serverReader.dataChan = readerChan
	err := serverReader.Open()
	assert.Nil(t, err)

	serverReaderAddr := serverReader.Status().Address

	clientWriter := NewTCPHandler(serverReaderAddr, 0, 0, Client, Writer)
	clientWriter.dataChan = writerChan
	err = clientWriter.Open()
	assert.Nil(t, err)

	clientWriterAddr := clientWriter.Status().Address

	t.Run("TestNewTCPHandler", func(t *testing.T) {
		status := serverReader.Status()
		assert.Equal(t, Server, status.Mode)
		assert.Equal(t, Reader, status.Role)

		status = clientWriter.Status()
		assert.Equal(t, serverReaderAddr, status.Address)
		assert.Equal(t, Client, status.Mode)
		assert.Equal(t, Writer, status.Role)
	})

	t.Run("TestWriteData", func(t *testing.T) {
		randBytes := make([]byte, 188)
		_, _ = rand.Read(randBytes)
		writerChan.Send(randBytes)

		select {
		case <-time.After(5 * time.Millisecond):
			assert.Fail(t, "Timeout waiting for data")
		default:
			data := readerChan.Receive()
			assert.Equal(t, randBytes, data)
		}

		assert.Equal(t, 1, len(serverReader.Status().Connections))
		assert.Equal(t, clientWriter.Status().Connections[clientWriterAddr], clientWriter.Status().Connections[clientWriterAddr])

		assert.Equal(t, 1, len(clientWriter.Status().Connections))
		assert.Equal(t, serverReader.Status().Connections[serverReaderAddr], clientWriter.Status().Connections[clientWriterAddr])
	})
}

// TestServerWriterClientReaderTCP tests the TCPHandler functionality with a server in writer role and a client in reader role, ensuring proper data transfer.
func TestServerWriterClientReaderTCP(t *testing.T) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		fmt.Printf("TestServerWriterClientReaderTCP took %v\n", duration)
	}()

	writerChan := channels.NewPacketChan(1)
	readerChan := channels.NewPacketChan(1)

	serverWriter := NewTCPHandler(":0", 0, 0, Server, Writer)
	serverWriter.dataChan = writerChan
	err := serverWriter.Open()
	assert.Nil(t, err)

	serverWriterAddr := serverWriter.Status().Address

	clientReader := NewTCPHandler(serverWriterAddr, 0, 0, Client, Reader)
	clientReader.dataChan = readerChan
	err = clientReader.Open()
	assert.Nil(t, err)

	t.Run("TestNewTCPHandler", func(t *testing.T) {
		serverStatus := serverWriter.Status()
		assert.Equal(t, Server, serverStatus.Mode)
		assert.Equal(t, Writer, serverStatus.Role)
		assert.Equal(t, serverWriterAddr, serverStatus.Address)

		clientStatus := clientReader.Status()
		assert.Equal(t, Client, clientStatus.Mode)
		assert.Equal(t, Reader, clientStatus.Role)
		assert.Equal(t, serverWriterAddr, clientStatus.Address)
	})

	t.Run("TestWriteData", func(t *testing.T) {
		randBytes := make([]byte, 188)
		_, _ = rand.Read(randBytes)
		writerChan.Send(randBytes)

		select {
		case <-time.After(5 * time.Second):
			assert.Fail(t, "Timeout waiting for data")
		default:
			data := readerChan.Receive()
			assert.Equal(t, randBytes, data)
		}
	})

	assert.Nil(t, serverWriter.Close())
	assert.Nil(t, clientReader.Close())
}
