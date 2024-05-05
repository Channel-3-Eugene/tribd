package main

import (
	"log"
	"net"

	"github.com/Channel-3-Eugene/tribd/config"
	"github.com/Channel-3-Eugene/tribd/mpegts"
)

// InputStream represents an input stream read from UDP.
type InputStream struct {
	IPAddress string // IP address of the UDP source
	Port      int    // Port number of the UDP source
	ServiceID int
	ID        string
	Name      string
}

// NewInputStream creates a new input stream with the given parameters.
func NewInputStream(config config.ReaderConfig) *InputStream {
	return &InputStream{
		ID:        config.ID,
		IPAddress: config.IPAddress,
		Port:      config.Port,
		ServiceID: config.ServiceID,
		Name:      config.Name,
	}
}

func (inputStream *InputStream) Start(ch chan<- *mpegts.EncodedPacket, done <-chan struct{}) {
	addr := net.UDPAddr{
		IP:   net.ParseIP(inputStream.IPAddress),
		Port: inputStream.Port,
	}
	conn, err := net.ListenUDP("udp6", &addr) // Connect to the server, do not bind
	if err != nil {
		log.Fatalf("Failed to connect to UDP server: %v", err)
		return
	}
	defer conn.Close()

	// Assume it reads from conn to receive packets sent by the server
	buffer := make([]byte, 2048) // Adjust size as needed
	for {
		select {
		case <-done:
			return
		default:
			n, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				log.Printf("Error reading from UDP: %v", err)
				continue
			}
			if n%188 != 0 {
				log.Printf("Received misaligned data from UDP: %d bytes", n)
				continue // Handle or log the misalignment
			}
			// Process each TS packet within the buffer
			for i := 0; i < n; i += 188 {
				rawPacket := *(*[188]byte)(buffer[i : i+188]) // Convert slice to array
				packet, err := mpegts.NewMPEGTSPacket(rawPacket)
				if err != nil {
					log.Printf("Error validating MPEG-TS packet: %v", err)
					continue // Skip invalid packets
				}

				if packet.IsNullPacket() {
					continue // Skip null packets
				}
				packet.ClearPCR() // Strip PCR if needed
				ch <- packet      // Send packet to the channel
			}
		}

	}
}

// OutputStream represents an output stream pushed to UDP.
type OutputStream struct {
	IPAddress string // IP address of the UDP destination
	Port      int    // Port number of the UDP destination
	// ...
	conn *net.UDPConn // UDP connection
}

// Write writes data to the output stream via UDP.
func (outputStream *OutputStream) Write(p mpegts.EncodedPacket) (int, error) {
	if outputStream.conn == nil {
		addr := net.UDPAddr{
			IP:   net.ParseIP(outputStream.IPAddress),
			Port: outputStream.Port,
		}
		conn, err := net.DialUDP("udp", nil, &addr)
		if err != nil {
			return 0, err
		}
		outputStream.conn = conn
	}
	return outputStream.conn.Write(p[:])
}

// Close closes the output stream.
func (outputStream *OutputStream) Close() error {
	if outputStream.conn != nil {
		return outputStream.conn.Close()
	}
	return nil
}
