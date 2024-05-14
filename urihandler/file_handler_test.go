package uriHandler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewFileHandler verifies that a new FileHandler is correctly initialized with specified parameters.
func TestNewFileHandler(t *testing.T) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		fmt.Printf("TestNewFileHandler took %v\n", duration)
	}()

	filePath := randFileName()
	readTimeout := 5 * time.Millisecond
	writeTimeout := 5 * time.Millisecond

	handler := NewFileHandler(filePath, Writer, false, readTimeout, writeTimeout)
	dataChan := handler.dataChan

	// Assert that all properties are set as expected.
	assert.Equal(t, filePath, handler.filePath)
	assert.Equal(t, Writer, handler.role)
	assert.Equal(t, false, handler.isFIFO)
	assert.NotNil(t, dataChan)
	assert.Equal(t, readTimeout, handler.readTimeout)
	assert.Equal(t, writeTimeout, handler.writeTimeout)
}

// TestFileHandlerOpenAndClose tests the Open and Close methods of the FileHandler to ensure files are correctly managed.
func TestFileHandlerOpenAndClose(t *testing.T) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		fmt.Printf("TestFileHandlerOpenAndClose took %v\n", duration)
	}()

	filePath := randFileName()

	// Ensure any existing file with the same name is removed before starting the test.
	os.Remove(filePath)

	// Open the handler and verify that the file exists after opening.
	handler := NewFileHandler(filePath, Writer, false, 0, 0)
	err := handler.Open()
	assert.Nil(t, err)
	assert.FileExists(t, filePath)

	// Close the handler and check if the file still accessible, indicating proper closure.
	err = handler.Close()
	assert.Nil(t, err)

	// Check if file can be re-opened, indicating it was properly closed.
	file, err := os.OpenFile(filePath, os.O_WRONLY, 0666)
	assert.Nil(t, err)
	file.Close()

	// Clean up the created file after the test.
	os.Remove(filePath)
}

// TestFileHandlerFIFO checks the functionality of the FileHandler with FIFO specific operations.
func TestFileHandlerFIFO(t *testing.T) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		fmt.Printf("TestFileHandlerFIFO took %v\n", duration)
	}()

	filePath := randFileName()
	handler := NewFileHandler(filePath, Reader, true, 1, 1)

	// Open the handler as a FIFO and ensure the FIFO file exists.
	err := handler.Open()
	assert.Nil(t, err)
	_, err = os.Stat(filePath)
	assert.Nil(t, err)

	// Test opening the FIFO by another process to simulate a writer.
	writer, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	assert.Nil(t, err)
	writer.Close()
}

// TestFileHandlerDataFlow tests the complete cycle of writing to and reading from the file.
func TestFileHandlerDataFlow(t *testing.T) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		fmt.Printf("TestFileHandlerDataFlow took %v\n", duration)
	}()

	filePath := randFileName()

	// Initialize writer and reader handlers.
	writer := NewFileHandler(filePath, Writer, false, 0, 0)
	writer.Open()

	reader := NewFileHandler(filePath, Reader, false, 0, 0)
	reader.Open()

	// Write data to the file.
	testData := []byte("hello, world")
	go func() {
		err := writer.dataChan.Send(testData)
		assert.Nil(t, err)
		writer.dataChan.Close()
	}()

	// Attempt to read the data and check if matches what was written.
	select {
	case <-time.After(5 * time.Millisecond):
		assert.Fail(t, "Timeout waiting for data")
	default:
		receivedData := reader.dataChan.Receive()
		assert.Equal(t, testData, receivedData)
	}

	// Clean up resources and remove test file.
	writer.Close()
	reader.Close()
	os.Remove(filePath)
}

// randFileName generates a random filename for testing, reducing the chance of file conflicts.
func randFileName() string {
	randBytes := make([]byte, 8) // Generates a unique identifier of 16 hex characters.
	_, err := rand.Read(randBytes)
	if err != nil {
		return ""
	}
	return "/tmp/" + hex.EncodeToString(randBytes) + ".txt"
}
