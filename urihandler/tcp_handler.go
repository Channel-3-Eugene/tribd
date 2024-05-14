// Package uriHandler provides utilities for handling different types of socket communications.
package uriHandler

import (
	"net"
	"sync"
	"time"

	"github.com/Channel-3-Eugene/tribd/channels" // Correct import path
)

// TCPStatus represents the status of a TCPHandler instance.
type TCPStatus struct {
	Mode        Mode              // Mode represents the operational mode of the TCPHandler.
	Role        Role              // Role represents the role of the TCPHandler, whether it's a server or client.
	Address     string            // Address represents the network address the TCPHandler is bound to.
	Connections map[string]string // Connections holds a map of connection information (local address to remote address).
}

// GetMode returns the operational mode of the TCPHandler.
func (t TCPStatus) GetMode() Mode {
	return t.Mode
}

// GetRole returns the role of the TCPHandler.
func (t TCPStatus) GetRole() Role {
	return t.Role
}

// GetAddress returns the network address the TCPHandler is bound to.
func (t TCPStatus) GetAddress() string {
	return t.Address
}

// TCPHandler manages TCP connections and provides methods for handling TCP communication.
type TCPHandler struct {
	address       string                // address represents the network address the TCPHandler is bound to.
	readDeadline  time.Duration         // readDeadline represents the read deadline for incoming data.
	writeDeadline time.Duration         // writeDeadline represents the write deadline for outgoing data.
	mode          Mode                  // mode represents the operational mode of the TCPHandler.
	role          Role                  // role represents the role of the TCPHandler, whether it's a server or client.
	listener      net.Listener          // listener represents the TCP listener for server mode.
	dataChan      *channels.PacketChan  // dataChan is a PacketChan for sending and receiving data.
	connections   map[net.Conn]struct{} // connections holds a map of active TCP connections.
	mu            sync.RWMutex          // Use RWMutex to allow concurrent reads.

	status TCPStatus // status represents the current status of the TCPHandler.
}

// NewTCPHandler creates a new instance of TCPHandler with the specified configuration.
func NewTCPHandler(address string, readDeadline, writeDeadline time.Duration, mode Mode, role Role) *TCPHandler {
	h := &TCPHandler{
		address:       address,
		readDeadline:  readDeadline,
		writeDeadline: writeDeadline,
		mode:          mode,
		role:          role,
		dataChan:      channels.NewPacketChan(64 * 1024), // Initialize PacketChan with a buffer size
		connections:   make(map[net.Conn]struct{}),
	}

	// Initialize TCPStatus with default values.
	h.status = TCPStatus{
		Address:     address,
		Mode:        mode,
		Role:        role,
		Connections: make(map[string]string),
	}

	return h
}

// Open starts the TCPHandler instance based on its operational mode.
func (h *TCPHandler) Open() error {
	if h.mode == Client {
		return h.connectClient()
	} else if h.mode == Server {
		return h.startServer()
	}
	return nil
}

// Status returns the current status of the TCPHandler.
func (h *TCPHandler) Status() TCPStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Populate the connections map with active connection information.
	for c := range h.connections {
		h.status.Connections[c.LocalAddr().String()] = c.RemoteAddr().String()
	}

	return h.status
}

// connectClient establishes a client connection to the TCP server.
func (h *TCPHandler) connectClient() error {
	conn, err := net.Dial("tcp", h.address)
	if err != nil {
		return err
	}

	h.mu.Lock()
	h.connections[conn] = struct{}{}
	h.mu.Unlock()

	go h.manageStream(conn)
	return nil
}

// startServer starts the TCP server and listens for incoming client connections.
func (h *TCPHandler) startServer() error {
	ln, err := net.Listen("tcp", h.address)
	if err != nil {
		return err
	}
	h.listener = ln
	h.mu.Lock()
	h.status.Address = ln.Addr().String()
	h.mu.Unlock()
	go h.acceptClients()
	return nil
}

// acceptClients accepts incoming client connections and manages them concurrently.
func (h *TCPHandler) acceptClients() {
	for {
		conn, err := h.listener.Accept()
		if err != nil {
			continue
		}
		h.mu.Lock()
		h.connections[conn] = struct{}{}
		h.mu.Unlock()
		go h.manageStream(conn)
	}
}

// manageStream manages the TCP connection stream based on the role of the TCPHandler.
func (h *TCPHandler) manageStream(conn net.Conn) {
	defer func() {
		conn.Close()
		h.mu.Lock()
		delete(h.connections, conn)
		h.mu.Unlock()
	}()

	// Handle data transmission based on the role of the TCPHandler.
	if h.role == Writer {
		for {
			data := h.dataChan.Receive()
			if data == nil {
				break // Channel closed
			}
			if _, err := conn.Write(data); err != nil {
				break
			}
		}
	} else if h.role == Reader {
		readBuffer := make([]byte, 188*10)
		for {
			n, err := conn.Read(readBuffer)
			if err != nil {
				break
			}
			err = h.dataChan.Send(readBuffer[:n])
			if err != nil {
				break
			}
		}
	}
}

// Close closes the TCPHandler instance and all active connections.
func (h *TCPHandler) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.listener != nil {
		h.listener.Close()
	}
	for conn := range h.connections {
		conn.Close()
	}
	h.connections = nil
	h.dataChan.Close()
	return nil
}
