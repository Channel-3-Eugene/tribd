package uriHandler

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTCPHandler(t *testing.T) {
	dataChan := make(chan []byte)
	handler := NewTCPHandler(":0", 0, 0, Server, Reader, dataChan)
	assert.Equal(t, ":0", handler.address)
	assert.Equal(t, 0*time.Second, handler.readDeadline)
	assert.Equal(t, 0*time.Second, handler.writeDeadline)
	assert.Equal(t, Server, handler.mode)
	assert.Equal(t, Reader, handler.role)
	assert.Equal(t, dataChan, handler.dataChan)
	assert.NotNil(t, handler.connections)
}

func TestServerWriterClientReader(t *testing.T) {
	writerChan := make(chan []byte)
	readerChan := make(chan []byte)

	serverWriter := NewTCPHandler(":0", 0, 0, Server, Writer, writerChan)
	err := serverWriter.Open()
	assert.Nil(t, err)

	serverWriterAddr := serverWriter.Status().Address

	clientReader := NewTCPHandler(serverWriterAddr, 0, 0, Client, Reader, readerChan)
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
		writerChan <- randBytes

		select {
		case data := <-readerChan:
			assert.Equal(t, randBytes, data)
		case <-time.After(5 * time.Millisecond):
			assert.Fail(t, "Timeout waiting for data")
		}

		status := serverWriter.Status()
		assert.Equal(t, 1, len(status.Connections))
	})
}

func TestServerReaderClientWriter(t *testing.T) {
	writerChan := make(chan []byte)
	readerChan := make(chan []byte)

	serverReader := NewTCPHandler(":0", 0, 0, Server, Reader, readerChan)
	err := serverReader.Open()
	assert.Nil(t, err)

	serverReaderAddr := serverReader.Status().Address

	clientWriter := NewTCPHandler(serverReaderAddr, 0, 0, Client, Writer, writerChan)
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
		writerChan <- randBytes

		select {
		case data := <-readerChan:
			assert.Equal(t, randBytes, data)
		case <-time.After(5 * time.Millisecond):
			assert.Fail(t, "Timeout waiting for data")
		}

		assert.Equal(t, 1, len(serverReader.Status().Connections))
		assert.Equal(t, clientWriter.Status().Connections[clientWriterAddr], clientWriter.Status().Connections[clientWriterAddr])

		assert.Equal(t, 1, len(clientWriter.Status().Connections))
		assert.Equal(t, serverReader.Status().Connections[serverReaderAddr], clientWriter.Status().Connections[clientWriterAddr])
	})
}

func TestServerWriterClientReaderTCP(t *testing.T) {
	writerChan := make(chan []byte, 1)
	readerChan := make(chan []byte, 1)

	serverWriter := NewTCPHandler(":0", 0, 0, Server, Writer, writerChan)
	err := serverWriter.Open()
	assert.Nil(t, err)

	serverWriterAddr := serverWriter.Status().Address

	clientReader := NewTCPHandler(serverWriterAddr, 0, 0, Client, Reader, readerChan)
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
		writerChan <- randBytes

		select {
		case data := <-readerChan:
			assert.Equal(t, randBytes, data)
		case <-time.After(5 * time.Second):
			assert.Fail(t, "Timeout waiting for data")
		}
	})

	assert.Nil(t, serverWriter.Close())
	assert.Nil(t, clientReader.Close())
}
