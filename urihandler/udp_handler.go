package uriHandler

import (
	"net"
	"sync"
	"time"
)

type UDPHandler struct {
	address        string
	conn           *net.UDPConn
	readDeadline   time.Duration
	writeDeadline  time.Duration
	mode           Mode // Only 'peer' mode as UDP is inherently peer-to-peer
	role           Role
	dataChan       chan []byte
	allowedSources map[string]struct{} // For Reader role to filter allowed sources
	destinations   []*net.UDPAddr      // For Writer role to define target addresses
	mu             sync.Mutex
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
		destinations:   make([]*net.UDPAddr, len(destinations)),
	}

	for _, src := range sources {
		handler.allowedSources[src] = struct{}{}
	}

	for i, dst := range destinations {
		addr, _ := net.ResolveUDPAddr("udp", dst)
		handler.destinations[i] = addr
	}

	return handler
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

	if h.role == Writer {
		go h.sendData()
	} else if h.role == Reader {
		go h.receiveData()
	}
	return nil
}

func (h *UDPHandler) sendData() {
	defer h.conn.Close()
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
	readBuffer := make([]byte, 2048)
	for {
		h.conn.SetReadDeadline(time.Now().Add(h.readDeadline))
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
