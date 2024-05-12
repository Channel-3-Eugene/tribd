package uriHandler

type Mode string
type Role string

const (
	Peer   Mode = "peer" // Unified mode for simplicity, denoting peer-to-peer behavior
	Server Mode = "server"
	Client Mode = "client"
	Reader Role = "reader"
	Writer Role = "writer"
)

type DataHandler interface {
	Open() error
	Close() error
	SendData(data []byte) error
	ReceiveData() <-chan []byte
}

// Possible IO handlers we may eventually support:
// - SRT
// - RTMP
// - DVB
// - ASI
// - SCTE-35
