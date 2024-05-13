package uriHandler

import (
	"io"
	"os"
	"syscall"
	"time"
)

type FileHandler struct {
	filePath     string
	file         *os.File
	dataChan     chan []byte
	role         Role
	isFIFO       bool
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func NewFileHandler(filePath string, role Role, isFIFO bool, dataChan chan []byte, readTimeout, writeTimeout time.Duration) *FileHandler {
	return &FileHandler{
		filePath:     filePath,
		role:         role,
		isFIFO:       isFIFO,
		dataChan:     dataChan,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
	}
}

func (h *FileHandler) Open() error {
	var err error
	if _, err = os.Stat(h.filePath); os.IsNotExist(err) {
		if h.isFIFO {
			if err = syscall.Mkfifo(h.filePath, 0666); err != nil {
				return err
			}
		} else {
			if h.file, err = os.Create(h.filePath); err != nil {
				return err
			}
		}
	}

	if h.role == Reader {
		go h.readData()
	} else {
		go h.writeData()
	}
	return nil
}

func (h *FileHandler) readData() {
	var err error
	if h.file == nil {
		h.file, err = os.Open(h.filePath)
		if err != nil {
			return
		}
	}
	defer h.file.Close()
	defer close(h.dataChan)

	buffer := make([]byte, 4096)
	for {
		if h.readTimeout > 0 {
			select {
			case <-time.After(h.readTimeout):
				return // Timeout reached, exit goroutine
			default:
				n, err := h.file.Read(buffer)
				if err != nil {
					if err == io.EOF || err == syscall.EINTR {
						continue
					}
					return
				}
				h.dataChan <- buffer[:n]
			}
		} else {
			n, err := h.file.Read(buffer)
			if err != nil {
				if err == io.EOF || err == syscall.EINTR {
					continue
				}
				return
			}
			h.dataChan <- buffer[:n]
		}
	}
}

func (h *FileHandler) writeData() {
	var err error
	h.file, err = os.OpenFile(h.filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer h.file.Close()

	for data := range h.dataChan {
		if h.writeTimeout > 0 {
			select {
			case <-time.After(h.writeTimeout):
				return // Timeout reached, exit goroutine
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

func (h *FileHandler) Close() error {
	if h.file != nil {
		if h.isFIFO {
			err := h.file.Close()
			syscall.Unlink(h.filePath)
			return err
		}
		return h.file.Close()
	}
	return nil
}
