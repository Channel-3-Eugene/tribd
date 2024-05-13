package uriHandler

import (
	"net"
	"sync"
	"time"
)

type TCPStatus struct {
	Mode        Mode
	Role        Role
	Address     string
	Connections map[string]string
}

type TCPHandler struct {
	address       string
	readDeadline  time.Duration
	writeDeadline time.Duration
	mode          Mode
	role          Role
	listener      net.Listener
	dataChan      chan []byte
	connections   map[net.Conn]struct{}
	mu            sync.Mutex

	status TCPStatus
}

func NewTCPHandler(address string, readDeadline, writeDeadline time.Duration, mode Mode, role Role, dataChan chan []byte) *TCPHandler {
	h := &TCPHandler{
		address:       address,
		readDeadline:  readDeadline,
		writeDeadline: writeDeadline,
		mode:          mode,
		role:          role,
		dataChan:      dataChan,
		connections:   make(map[net.Conn]struct{}),
	}

	h.status = TCPStatus{
		Address:     address,
		Mode:        mode,
		Role:        role,
		Connections: map[string]string{},
	}

	return h
}

func (h *TCPHandler) Open() error {
	if h.mode == Client {
		return h.connectClient()
	} else if h.mode == Server {
		return h.startServer()
	}
	return nil
}

func (h *TCPHandler) Status() TCPStatus {
	h.mu.Lock()
	defer h.mu.Unlock()

	for c := range h.connections {
		h.status.Connections[c.LocalAddr().String()] = c.RemoteAddr().String()
	}

	return h.status
}

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

func (h *TCPHandler) startServer() error {
	ln, err := net.Listen("tcp", h.address)
	if err != nil {
		return err
	}
	h.listener = ln
	h.status.Address = ln.Addr().String()
	go h.acceptClients()
	return nil
}

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

func (h *TCPHandler) manageStream(conn net.Conn) {
	defer func() {
		conn.Close()
		h.mu.Lock()
		delete(h.connections, conn)
		h.mu.Unlock()
	}()

	if h.role == Writer {
		for batch := range h.dataChan {
			if _, err := conn.Write(batch); err != nil {
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
			h.dataChan <- readBuffer[:n]
		}
	}
}

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
	return nil
}
