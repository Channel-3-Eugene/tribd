package uriHandler

import (
	"net"
	"sync"
	"time"

	"github.com/Channel-3-Eugene/tribd/channels" // Correct import path
)

// UDPStatus represents the status of a UDPHandler, detailing its configuration and state.
type UDPStatus struct {
	Mode           Mode
	Role           Role
	Address        string
	ReadDeadline   time.Duration
	WriteDeadline  time.Duration
	AllowedSources []string // List of source addresses allowed to send data
	Destinations   []string // List of destination addresses to send data
}

// Getter methods for UDPStatus
func (u UDPStatus) GetMode() Mode      { return u.Mode }
func (u UDPStatus) GetRole() Role      { return u.Role }
func (u UDPStatus) GetAddress() string { return u.Address }

// UDPHandler manages UDP network communication, supporting roles as sender (writer) or receiver (reader).
type UDPHandler struct {
	address        string
	conn           *net.UDPConn
	readDeadline   time.Duration
	writeDeadline  time.Duration
	mode           Mode
	role           Role
	dataChan       *channels.PacketChan
	allowedSources map[string]struct{}     // Set of source IPs allowed to send data to this handler
	destinations   map[string]*net.UDPAddr // UDP addresses for sending data
	mu             sync.RWMutex            // Mutex to protect concurrent access to handler state

	status UDPStatus
}

// NewUDPHandler initializes a new UDPHandler with specified settings.
func NewUDPHandler(address string, readDeadline, writeDeadline time.Duration, role Role, sources, destinations []string) *UDPHandler {
	handler := &UDPHandler{
		address:        address,
		readDeadline:   readDeadline,
		writeDeadline:  writeDeadline,
		mode:           Peer,
		role:           role,
		dataChan:       channels.NewPacketChan(64 * 1024), // TODO: get this from config
		allowedSources: make(map[string]struct{}),
		destinations:   make(map[string]*net.UDPAddr),
	}

	// Populate allowed sources.
	for _, src := range sources {
		if _, err := net.ResolveUDPAddr("udp", src); err == nil {
			handler.allowedSources[src] = struct{}{}
		}
	}

	// Populate destinations.
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

// Status returns the current status of the UDPHandler.
func (h *UDPHandler) Status() UDPStatus {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Convert internal maps to slices for easier external consumption.
	sources := make([]string, 0, len(h.allowedSources))
	for src := range h.allowedSources {
		sources = append(sources, src)
	}

	destinations := make([]string, 0, len(h.destinations))
	for dst := range h.destinations {
		destinations = append(destinations, dst)
	}

	return UDPStatus{
		Mode:           h.mode,
		Role:           h.role,
		Address:        h.conn.LocalAddr().String(),
		ReadDeadline:   h.readDeadline,
		WriteDeadline:  h.writeDeadline,
		AllowedSources: sources,
		Destinations:   destinations,
	}
}

// Open starts the UDPHandler, setting up a UDP connection for sending or receiving data.
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

// Methods for managing allowed sources and destinations.
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

// sendData handles sending data to the configured destinations.
func (h *UDPHandler) sendData() {
	defer h.conn.Close()

	if h.writeDeadline > 0 {
		h.conn.SetWriteDeadline(time.Now().Add(h.writeDeadline))
	}

	for {
		data := h.dataChan.Receive()
		if data == nil {
			break // Channel closed
		}

		for _, addr := range h.destinations {
			_, err := h.conn.WriteToUDP(data, addr)
			if err != nil {
				break
			}
		}
	}
}

// receiveData continues to handle data reception and uses PacketChan.
func (h *UDPHandler) receiveData() {
	defer h.conn.Close()

	if h.readDeadline > 0 {
		h.conn.SetReadDeadline(time.Now().Add(h.readDeadline))
	}

	bufferPool := sync.Pool{
		New: func() interface{} {
			return new([]byte)
		},
	}

	for {
		rawBuffer := bufferPool.Get().(*[]byte) // Get a buffer from the pool
		if cap(*rawBuffer) < 2048 {
			*rawBuffer = make([]byte, 2048)
		}

		n, addr, err := h.conn.ReadFromUDP(*rawBuffer)
		if err != nil {
			bufferPool.Put(rawBuffer)
			continue
		}

		h.mu.RLock()
		_, ok := h.allowedSources[addr.IP.String()]
		h.mu.RUnlock()

		if !ok {
			bufferPool.Put(rawBuffer)
			continue
		}

		// Send the data through the channel.
		err = h.dataChan.Send((*rawBuffer)[:n])
		if err != nil {
			bufferPool.Put(rawBuffer)
		}
	}
}

// Close terminates the handler's operations and closes the UDP connection.
func (h *UDPHandler) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conn != nil {
		h.conn.Close()
	}
	h.dataChan.Close()
	return nil
}
