// Package uriHandler provides a set of tools for file handling operations that can interact with regular files and FIFOs.
package uriHandler

import (
	"io"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/Channel-3-Eugene/tribd/channels" // Correct import path
)

// FileStatus represents the current status of a FileHandler, including operational configuration and state.
type FileStatus struct {
	FilePath     string
	IsFIFO       bool
	Mode         Mode
	Role         Role
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IsOpen       bool
}

// GetMode returns the operation mode of the file handler.
func (f FileStatus) GetMode() Mode { return f.Mode }

// GetRole returns the operational role of the file handler.
func (f FileStatus) GetRole() Role { return f.Role }

// GetAddress returns the file path associated with the file handler.
func (f FileStatus) GetAddress() string { return f.FilePath }

// FileHandler manages the operations for a file, supporting both regular file operations and FIFO-based interactions.
type FileHandler struct {
	filePath     string
	file         *os.File
	dataChan     *channels.PacketChan
	mode         Mode
	role         Role
	isFIFO       bool
	readTimeout  time.Duration
	writeTimeout time.Duration
	isOpen       bool       // Tracks the open or closed state of the file.
	mu           sync.Mutex // mu is a mutex for synchronizing access to connections and other shared resources.
}

// NewFileHandler creates a new FileHandler with specified configurations.
func NewFileHandler(filePath string, role Role, isFIFO bool, readTimeout, writeTimeout time.Duration) *FileHandler {
	return &FileHandler{
		filePath:     filePath,
		mode:         Peer,
		role:         role,
		isFIFO:       isFIFO,
		dataChan:     channels.NewPacketChan(64 * 1024), // Initialize PacketChan with a buffer size
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
		isOpen:       false,
	}
}

// Status provides the current status of the FileHandler.
func (h *FileHandler) Status() FileStatus {
	return FileStatus{
		FilePath:     h.filePath,
		IsFIFO:       h.isFIFO,
		Mode:         h.mode,
		Role:         h.role,
		ReadTimeout:  h.readTimeout,
		WriteTimeout: h.writeTimeout,
		IsOpen:       h.isOpen,
	}
}

// Open initializes the file handler by opening or creating the file and starting the appropriate data processing goroutines.
func (h *FileHandler) Open() error {
	var err error
	// Check if the file exists; if not, create it or initialize a FIFO.
	if _, err = os.Stat(h.filePath); os.IsNotExist(err) {
		if h.isFIFO {
			if err = syscall.Mkfifo(h.filePath, 0666); err != nil {
				return err
			}
		} else {
			h.mu.Lock()
			h.file, err = os.Create(h.filePath)
			h.mu.Unlock()
			if err != nil {
				return err
			}
		}
	} else {
		h.mu.Lock()
		h.file, err = os.OpenFile(h.filePath, os.O_APPEND|os.O_CREATE, 0666)
		h.mu.Unlock()
		if err != nil {
			return err
		}
	}

	h.mu.Lock()
	h.isOpen = true // Mark the file as open.
	h.mu.Unlock()

	if h.role == Reader {
		go h.readData()
	} else {
		go h.writeData()
	}
	return nil
}

// Close terminates the file handler's operations and closes the file.
func (h *FileHandler) Close() error {
	if h.file != nil {
		h.mu.Lock()
		err := h.file.Close()
		h.isOpen = false // Update the state to closed.
		h.mu.Unlock()
		if h.isFIFO {
			syscall.Unlink(h.filePath) // Remove the FIFO file.
		}
		h.dataChan.Close()
		return err
	}
	return nil
}

// readData handles the data reading operations from the file based on configured timeouts.
func (h *FileHandler) readData() {
	var err error
	if h.file == nil {
		h.mu.Lock()
		h.file, err = os.Open(h.filePath)
		h.mu.Unlock()
		if err != nil {
			return
		}
	}
	defer h.file.Close()

	buffer := make([]byte, 4096)
	for {
		if h.readTimeout > 0 {
			select {
			case <-time.After(h.readTimeout):
				return // Exit the goroutine after a timeout.
			default:
				n, err := h.file.Read(buffer)
				if err != nil {
					if err == io.EOF || err == syscall.EINTR {
						continue
					}
					return
				}
				err = h.dataChan.Send(buffer[:n])
				if err != nil {
					return
				}
			}
		} else {
			n, err := h.file.Read(buffer)
			if err != nil {
				if err == io.EOF || err == syscall.EINTR {
					continue
				}
				return
			}
			err = h.dataChan.Send(buffer[:n])
			if err != nil {
				return
			}
		}
	}
}

// writeData handles the data writing operations to the file based on configured timeouts.
func (h *FileHandler) writeData() {
	var err error
	h.mu.Lock()
	h.file, err = os.OpenFile(h.filePath, os.O_WRONLY|os.O_CREATE, 0666)
	h.mu.Unlock()
	if err != nil {
		return
	}
	defer h.file.Close()

	for {
		data := h.dataChan.Receive()
		if data == nil {
			return // Channel closed
		}

		if h.writeTimeout > 0 {
			select {
			case <-time.After(h.writeTimeout):
				return // Exit the goroutine after a timeout.
			default:
				_, err := h.file.Write(data)
				if err != nil {
					return
				}
			}
		} else {
			_, err := h.file.Write(data)
			if err != nil {
				return
			}
		}
	}
}
