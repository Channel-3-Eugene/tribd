package channels

import (
	"fmt"
	"sync"
)

// Packet encapsulates the buffer and the pool reference to manage its lifecycle properly.
type Packet struct {
	buffer []byte
	pool   *sync.Pool
	mu     sync.Mutex // Protects buffer and pool fields
}

// Data retrieves the data and automatically releases the buffer back to the pool.
// It returns a copy of the data to ensure safety after buffer release.
func (p *Packet) Data() []byte {
	p.mu.Lock()
	defer p.mu.Unlock()

	data := make([]byte, len(p.buffer))
	copy(data, p.buffer)
	p.release() // Manually release the buffer to the pool
	return data
}

// release returns the buffer back to its pool, clearing the reference.
func (p *Packet) release() {
	p.pool.Put(p)
	p.buffer = nil // Clear the reference to prevent reuse
}

type PacketChan struct {
	ch     chan *Packet
	pool   *sync.Pool
	mu     sync.Mutex // Protects closing of channel and sending
	closed bool       // Indicates if the channel is closed
}

// NewPacketChan creates a new PacketChan with the specified buffer size.
func NewPacketChan(size int) *PacketChan {
	pool := &sync.Pool{
		New: func() interface{} {
			return &Packet{
				buffer: make([]byte, 0, 2048),
			}
		},
	}

	return &PacketChan{
		ch:   make(chan *Packet, size),
		pool: pool,
	}
}

// Send sends a packet to the channel and handles channel closure gracefully.
func (p *PacketChan) Send(data []byte) error {
	packet := p.pool.Get().(*Packet)
	packet.pool = p.pool // Assign the pool reference here

	packet.mu.Lock()
	packet.buffer = append(packet.buffer[:0], data...) // Reuse buffer, resetting and copying data
	packet.mu.Unlock()

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		packet.release() // Ensure buffer is released if channel is closed
		return fmt.Errorf("failed to send data: channel closed")
	}

	select {
	case p.ch <- packet: // Attempt to send the packet
		return nil // Successful send
	default:
		packet.release() // Ensure buffer is released if send fails
		return fmt.Errorf("failed to send data: buffer full")
	}
}

// Receive receives a packet from the channel, abstracting the data handling.
func (p *PacketChan) Receive() []byte {
	packet, ok := <-p.ch
	if !ok {
		return nil // Channel closed
	}
	return packet.Data() // This handles the release of the buffer
}

// Close closes the channel to prevent further sends.
func (p *PacketChan) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.closed {
		close(p.ch)
		p.closed = true
	}
}
