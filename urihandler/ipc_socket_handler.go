// Package uriHandler provides utilities for handling different types of socket communications.
package uriHandler

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/Channel-3-Eugene/tribd/channels" // Correct import path
)

// SocketStatus defines the status of a SocketHandler including its mode, role, and current connections.
type SocketStatus struct {
	Mode          Mode
	Role          Role
	Address       string
	Connections   []string // List of connection identifiers for simplicity
	ReadDeadline  time.Duration
	WriteDeadline time.Duration
}

// GetMode returns the mode of the socket.
func (s SocketStatus) GetMode() Mode { return s.Mode }

// GetRole returns the role of the socket.
func (s SocketStatus) GetRole() Role { return s.Role }

// GetAddress returns the address the socket is bound to.
func (s SocketStatus) GetAddress() string { return s.Address }

// SocketHandler manages socket connections, providing methods to open, close, and manage streams.
type SocketHandler struct {
	socketPath    string
	readDeadline  time.Duration
	writeDeadline time.Duration
	mode          Mode
	role          Role
	listener      net.Listener
	dataChan      *channels.PacketChan
	connections   map[net.Conn]struct{}
	mu            sync.RWMutex // Use RWMutex to allow concurrent reads
	status        SocketStatus
}

// NewSocketHandler creates and initializes a new SocketHandler with the specified parameters.
func NewSocketHandler(socketPath string, readDeadline, writeDeadline time.Duration, mode Mode, role Role) *SocketHandler {
	return &SocketHandler{
		socketPath:    socketPath,
		readDeadline:  readDeadline,
		writeDeadline: writeDeadline,
		mode:          mode,
		role:          role,
		dataChan:      channels.NewPacketChan(64 * 1024), // Initialize PacketChan with a buffer size
		connections:   make(map[net.Conn]struct{}),
		status: SocketStatus{
			Address:       socketPath,
			Mode:          mode,
			Role:          role,
			Connections:   []string{},
			ReadDeadline:  readDeadline,
			WriteDeadline: writeDeadline,
		},
	}
}

// Open initializes the socket's server or client based on its mode.
func (h *SocketHandler) Open() error {
	if h.mode == Client {
		go h.connectClient()
	} else if h.mode == Server {
		go h.startServer()
	}
	return nil
}

// Status returns the current status of the socket.
func (h *SocketHandler) Status() SocketStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()

	connections := []string{} // Reset the list
	for conn := range h.connections {
		// Using remote address or local if remote not available
		connDesc := conn.RemoteAddr().String()
		if connDesc == "" {
			connDesc = conn.LocalAddr().String()
		}
		connections = append(connections, connDesc)
	}
	h.status.Connections = connections

	return h.status
}

// connectClient manages the client connection to the server.
func (h *SocketHandler) connectClient() {
	conn, err := net.Dial("unix", h.socketPath)
	if err != nil {
		fmt.Printf("Error connecting to socket: %#v %s", err, err.Error())
		return
	}
	h.mu.Lock()
	h.connections[conn] = struct{}{}
	h.mu.Unlock()

	h.manageStream(conn)
}

// startServer starts the socket server and listens for incoming connections.
func (h *SocketHandler) startServer() {
	ln, err := net.Listen("unix", h.socketPath)
	if err != nil {
		fmt.Printf("Error creating socket: %#v %s", err, err.Error())
		return
	}
	h.mu.Lock()
	h.listener = ln
	h.status.Address = ln.Addr().String()
	h.mu.Unlock()

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

// manageStream handles data transmission over the connection based on the socket's role.
func (h *SocketHandler) manageStream(conn net.Conn) {
	defer func() {
		conn.Close()
		h.mu.Lock()
		delete(h.connections, conn)
		h.mu.Unlock()
	}()

	if h.role == Writer {
		h.handleWrite(conn)
	} else if h.role == Reader {
		h.handleRead(conn)
	}
}

// handleWrite manages writing data to the connection.
func (h *SocketHandler) handleWrite(conn net.Conn) {
	if h.writeDeadline > 0 {
		conn.SetWriteDeadline(time.Now().Add(h.writeDeadline))
	}
	for {
		data := h.dataChan.Receive()
		if data == nil {
			break // Channel closed
		}
		_, err := conn.Write(data)
		if err != nil {
			fmt.Println("Error writing to connection:", err)
			break // Exit if there is an error writing
		}
	}
}

// handleRead manages reading data from the connection.
func (h *SocketHandler) handleRead(conn net.Conn) {
	readBuffer := make([]byte, 4096) // Buffer size can be adjusted as needed
	if h.readDeadline > 0 {
		conn.SetReadDeadline(time.Now().Add(h.readDeadline))
	}
	for {
		n, err := conn.Read(readBuffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading from connection:", err)
			}
			break // Exit on error or when EOF is reached
		}
		// Send the data to the data channel for further processing
		err = h.dataChan.Send(readBuffer[:n])
		if err != nil {
			break
		}
	}
}

// Close shuts down the socket and cleans up resources.
func (h *SocketHandler) Close() error {
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
