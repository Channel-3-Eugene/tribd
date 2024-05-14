package uriHandler

import (
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/Channel-3-Eugene/tribd/channels"
	"github.com/stretchr/testify/assert"
)

// TestNewUDPHandler verifies the creation and initialization of a new UDPHandler with specific configuration settings.
func TestNewUDPHandler(t *testing.T) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		fmt.Printf("TestNewUDPHandler took %v\n", duration)
	}()

	sources := []string{}      // No allowed sources specified for this test.
	destinations := []string{} // No destinations specified for this test.

	// Create a new UDPHandler with a listening address, read and write timeouts, and roles specified.
	handler := NewUDPHandler(":0", 10*time.Second, 5*time.Second, Reader, sources, destinations)

	// Assert the handler is initialized with the correct configurations.
	assert.Equal(t, ":0", handler.address)
	assert.Equal(t, 10*time.Second, handler.readDeadline)
	assert.Equal(t, 5*time.Second, handler.writeDeadline)
	assert.Equal(t, Peer, handler.mode)
	assert.Equal(t, Reader, handler.role)
	assert.NotNil(t, handler.dataChan)
	assert.Len(t, handler.allowedSources, 0)
	assert.Len(t, handler.destinations, 0)
}

// TestUDPHandlerDataFlow tests the UDP data flow from a writer to a reader to ensure data sent by the writer is correctly received by the reader.
func TestUDPHandlerDataFlow(t *testing.T) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		fmt.Printf("TestUDPHandlerDataFlow took %v\n", duration)
	}()

	readChan := channels.NewPacketChan(1)

	// Setup writer and reader handlers, each on separate local UDP addresses.
	writer := NewUDPHandler("[::1]:0", 0, 0, Writer, nil, nil)
	err := writer.Open()
	assert.Nil(t, err)

	reader := NewUDPHandler("[::1]:0", 0, 0, Reader, nil, nil)
	err = reader.Open()
	assert.Nil(t, err)
	reader.dataChan = readChan

	// Ensure both the reader and the writer have unique, valid local addresses.
	assert.NotEmpty(t, reader.conn.LocalAddr().String())
	assert.NotEmpty(t, writer.conn.LocalAddr().String())
	assert.NotEqual(t, reader.conn.LocalAddr().String(), writer.conn.LocalAddr().String())

	// Configure the reader to accept data from the writer's address.
	err = reader.AddSource("::1")
	assert.Nil(t, err)

	// Configure the writer to send data to the reader's address.
	err = writer.AddDestination(reader.conn.LocalAddr().String())
	assert.Nil(t, err)

	// Subtest to write data from the writer and verify it is received by the reader.
	t.Run("TestWriteAndReceiveData", func(t *testing.T) {
		randBytes := make([]byte, 188) // Prepare a random byte slice for testing.
		_, _ = rand.Read(randBytes)    // Read random data into randBytes.

		go func() {
			err := writer.dataChan.Send(randBytes) // Send data to the writer channel.
			assert.Nil(t, err)
		}()

		// Wait to receive data on the reader channel or fail after a timeout.
		select {
		case <-time.After(10 * time.Millisecond):
			assert.Fail(t, "Timeout waiting for data") // Fail test if no data is received in time.
		default:
			data := readChan.Receive()
			assert.Equal(t, randBytes, data) // Check if received data matches the sent data.
		}
	})

	// Clean up: Close both handlers.
	assert.Nil(t, writer.Close())
	assert.Nil(t, reader.Close())
}
