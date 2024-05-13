package uriHandler

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewUDPHandler(t *testing.T) {
	writerChan := make(chan []byte)
	sources := []string{}
	destinations := []string{}

	handler := NewUDPHandler(":0", 10*time.Second, 5*time.Second, Reader, writerChan, sources, destinations)

	assert.Equal(t, ":0", handler.address)
	assert.Equal(t, 10*time.Second, handler.readDeadline)
	assert.Equal(t, 5*time.Second, handler.writeDeadline)
	assert.Equal(t, Peer, handler.mode)
	assert.Equal(t, Reader, handler.role)
	assert.Equal(t, writerChan, handler.dataChan)
	assert.Len(t, handler.allowedSources, 0)
	assert.Len(t, handler.destinations, 0)
}

func TestUDPHandlerDataFlow(t *testing.T) {
	writerChan := make(chan []byte, 1)
	readChan := make(chan []byte, 1)

	// UDP writer
	writer := NewUDPHandler("[::1]:0", 0, 0, Writer, writerChan, nil, nil)
	err := writer.Open()
	assert.Nil(t, err)

	// UDP reader
	reader := NewUDPHandler("[::1]:0", 0, 0, Reader, readChan, nil, nil)
	err = reader.Open()
	assert.Nil(t, err)

	// Ensure reader and writer are listening on different ports
	assert.NotEmpty(t, reader.conn.LocalAddr().String())
	assert.NotEmpty(t, writer.conn.LocalAddr().String())
	assert.NotEqual(t, reader.conn.LocalAddr().String(), writer.conn.LocalAddr().String())

	// Writer is a reader source
	err = reader.AddSource("::1")
	assert.Nil(t, err)

	// Reader is a writer destination
	err = writer.AddDestination(reader.conn.LocalAddr().String())
	assert.Nil(t, err)

	t.Run("TestWriteAndReceiveData", func(t *testing.T) {
		randBytes := make([]byte, 188)
		_, _ = rand.Read(randBytes)
		writerChan <- randBytes

		select {
		case data := <-readChan:
			assert.Equal(t, randBytes, data)
		case <-time.After(100 * time.Millisecond):
			assert.Fail(t, "Timeout waiting for data")
		}
	})

	// Ensure connections are properly closed after tests
	assert.Nil(t, writer.Close())
	assert.Nil(t, reader.Close())
}
