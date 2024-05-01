package main

import (
	"net"
)

// OTWTSPacket represents a raw OTW TS packet.
type OTWTSPacket []byte

// InputStream represents an input stream read from UDP.
type InputStream struct {
	IPAddress string      // IP address of the UDP source
	Port      int         // Port number of the UDP source
	ServiceID int
	ID        string
	Name      string
}

// NewInputStream creates a new input stream with the given parameters.
func NewInputStream(config ReaderConfig) *InputStream {
	return &InputStream{
		ID:        config.ID,
		IPAddress: config.IPAddress
		Port:      config.Port,
		ServiceID: config.ServiceID,
		Name:      config.Name,
	}
}

// Start starts reading data from the input stream via UDP.
func (inputStream *InputStream) Start(ch chan<- OTWTSPacket, done <-chan struct{}) {
	addr := net.UDPAddr{
		IP:   net.ParseIP(inputStream.IPAddress),
		Port: inputStream.Port,
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Continuously read data from the connection
	for {
		select { // needs timeout
		case <-done:
			return
		default:
			packet := make([]byte, 188)
			_, err := conn.Read(packet)
			if err != nil {
				panic(err)
			}
			pa := DisassemblePacket(packet)
			if err := pa.AdjustBitrate(0); err != nil {
				// handle error
			}
			if err := pa.CalculateTimeValues([]byte{}); err != nil {
				// handle error
			}
			ch <- pa.ReassemblePacket()
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
func (outputStream *OutputStream) Write(p OTWTPacket) (int, error) {
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
	return outputStream.conn.Write(p)
}

// Close closes the output stream.
func (outputStream *OutputStream) Close() error {
	if outputStream.conn != nil {
		return outputStream.conn.Close()
	}
	return nil
}
