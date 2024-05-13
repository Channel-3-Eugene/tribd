package uriHandler

import (
	"net"
	"sync"
	"time"
)

type UDPStatus struct {
	Mode    Mode
	Role    Role
	Address string
}

type UDPHandler struct {
	address        string
	conn           *net.UDPConn
	readDeadline   time.Duration
	writeDeadline  time.Duration
	mode           Mode // Only 'peer' mode as UDP is inherently peer-to-peer
	role           Role
	dataChan       chan []byte
	allowedSources map[string]struct{}
	destinations   map[string]*net.UDPAddr
	mu             sync.Mutex

	status UDPStatus
}

func NewUDPHandler(address string, readDeadline, writeDeadline time.Duration, role Role, dataChan chan []byte, sources, destinations []string) *UDPHandler {
	handler := &UDPHandler{
		address:        address,
		readDeadline:   readDeadline,
		writeDeadline:  writeDeadline,
		mode:           Peer,
		role:           role,
		dataChan:       dataChan,
		allowedSources: make(map[string]struct{}),
		destinations:   make(map[string]*net.UDPAddr),
	}

	for _, src := range sources {
		if _, err := net.ResolveUDPAddr("udp", src); err == nil {
			handler.allowedSources[src] = struct{}{}
		}
	}

	for _, dst := range destinations {
		if addr, err := net.ResolveUDPAddr("udp", dst); err == nil {
			handler.destinations[dst] = addr
		}
	}

	handler.status = UDPStatus{
		Mode: Peer,
		Role: role,
	}

	return handler
}

func (h *UDPHandler) Status() UDPStatus {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.status
}

func (h *UDPHandler) Open() error {
	udpAddr, err := net.ResolveUDPAddr("udp", h.address)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	h.conn = conn

	h.status.Address = conn.LocalAddr().String()

	if h.role == Writer {
		go h.sendData()
	} else if h.role == Reader {
		go h.receiveData()
	}
	return nil
}

func (h *UDPHandler) AddSource(addr string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.allowedSources[addr] = struct{}{}
	return nil
}

func (h *UDPHandler) RemoveSource(addr string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.allowedSources, addr)
	return nil
}

func (h *UDPHandler) AddDestination(addr string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	h.destinations[addr] = udpAddr
	return nil
}

func (h *UDPHandler) RemoveDestination(addr string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.destinations, addr)
	return nil
}

func (h *UDPHandler) sendData() {
	defer h.conn.Close()

	if h.writeDeadline > 0 {
		h.conn.SetWriteDeadline(time.Now().Add(h.writeDeadline))
	}

	for batch := range h.dataChan {

		for _, addr := range h.destinations {
			_, err := h.conn.WriteToUDP(batch, addr)
			if err != nil {
				break
			}
		}
	}
}

func (h *UDPHandler) receiveData() {
	defer h.conn.Close()

	if h.readDeadline > 0 {
		h.conn.SetReadDeadline(time.Now().Add(h.readDeadline))
	}

	readBuffer := make([]byte, 2048)

	for {
		n, addr, err := h.conn.ReadFromUDP(readBuffer)
		if err != nil {
			continue // Handle or log errors appropriately
		}
		if _, ok := h.allowedSources[addr.IP.String()]; !ok {
			continue // Ignore packets not from allowed sources
		}
		h.dataChan <- readBuffer[:n]
	}
}

func (h *UDPHandler) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conn != nil {
		h.conn.Close()
	}
	return nil
}
