package uriHandler

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewFileHandler(t *testing.T) {
	dataChan := make(chan []byte, 1)
	filePath := randFileName()
	readTimeout := 5 * time.Millisecond
	writeTimeout := 5 * time.Millisecond

	handler := NewFileHandler(filePath, Writer, false, dataChan, readTimeout, writeTimeout)

	assert.Equal(t, filePath, handler.filePath)
	assert.Equal(t, Writer, handler.role)
	assert.Equal(t, false, handler.isFIFO)
	assert.Equal(t, dataChan, handler.dataChan)
	assert.Equal(t, readTimeout, handler.readTimeout)
	assert.Equal(t, writeTimeout, handler.writeTimeout)
}

func TestFileHandlerOpenAndClose(t *testing.T) {
	dataChan := make(chan []byte)
	filePath := randFileName()

	// Cleanup before test
	os.Remove(filePath)

	handler := NewFileHandler(filePath, Writer, false, dataChan, 0, 0)
	err := handler.Open()
	assert.Nil(t, err)
	assert.FileExists(t, filePath)

	// Test closure
	err = handler.Close()
	assert.Nil(t, err)

	// Check if file can be re-opened, indicating it was properly closed
	file, err := os.OpenFile(filePath, os.O_WRONLY, 0666)
	assert.Nil(t, err)
	file.Close()

	// Cleanup after test
	os.Remove(filePath)
}

func TestFileHandlerFIFO(t *testing.T) {
	dataChan := make(chan []byte)
	filePath := randFileName()
	handler := NewFileHandler(filePath, Reader, true, dataChan, 1, 1)

	// Defer cleanup
	defer handler.Close()
	defer os.Remove(filePath)

	err := handler.Open()
	assert.Nil(t, err)

	// FIFO file should exist
	_, err = os.Stat(filePath)
	assert.Nil(t, err)

	// Ensure FIFO can be opened by another process as Writer
	writer, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	assert.Nil(t, err)
	writer.Close()
}

func TestFileHandlerDataFlow(t *testing.T) {
	filePath := randFileName()

	// Initialize handlers
	writeChan := make(chan []byte)
	writer := NewFileHandler(filePath, Writer, false, writeChan, 0, 0)
	writer.Open()

	readChan := make(chan []byte)
	reader := NewFileHandler(filePath, Reader, false, readChan, 0, 0)
	reader.Open()

	// Write data
	testData := []byte("hello, world")
	writeChan <- testData
	close(writeChan)

	// Read data
	select {
	case receivedData := <-readChan:
		assert.Equal(t, testData, receivedData)
	case <-time.After(5 * time.Millisecond):
		assert.Fail(t, "Timeout waiting for data")
	}

	// Clean up
	writer.Close()
	reader.Close()
	os.Remove(filePath)
}

func randFileName() string {
	randBytes := make([]byte, 8) // 8 bytes -> 16 characters in hex
	_, err := rand.Read(randBytes)
	if err != nil {
		return ""
	}
	return "/tmp/" + hex.EncodeToString(randBytes) + ".txt"
}
