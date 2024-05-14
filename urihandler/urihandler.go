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

type URIHandler interface {
	Open() error
	Close() error
	Status() interface{}
}

// Common status interface for all handlers
type Status interface {
	GetMode() Mode
	GetRole() Role
	GetAddress() string
}

// Possible IO handlers we may eventually support:
// - SRT
// - RTMP
// - DVB
// - ASI
// - SCTE-35

// TODO: Add IPC socket support using net.UnixConn.

// Writer:

// func main() {
// 	address := "/tmp/mysocket.sock"
// 	os.Remove(address) // Ensure the socket does not already exist

// 	// Setup the *net.UnixAddr for a Unix socket at the specified address
// 	unixAddr, err := net.ResolveUnixAddr("unix", address)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Listen for connections using Unix domain sockets
// 	listener, err := net.ListenUnix("unix", unixAddr)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer listener.Close()
// 	defer os.Remove(address) // Clean up the socket file on exit

// 	log.Println("Server listening at", address)
// 	for {
// 		conn, err := listener.AcceptUnix()
// 		if err != nil {
// 			log.Print(err)
// 			continue
// 		}
// 		go handleUnixConnection(conn)
// 	}
// }

// func handleUnixConnection(conn *net.UnixConn) {
// 	defer conn.Close()
// 	buffer := make([]byte, 512)
// 	for {
// 		n, _, err := conn.ReadFromUnix(buffer)
// 		if err != nil {
// 			log.Print(err)
// 			return
// 		}
// 		log.Printf("Received: %s", string(buffer[:n]))
// 		// Echo the data back to the client
// 		conn.WriteToUnix(buffer[:n], conn.RemoteAddr().(*net.UnixAddr))
// 	}
// }

// Reader:

// 	address := "/tmp/mysocket.sock"
// 	unixAddr, err := net.ResolveUnixAddr("unix", address)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	conn, err := net.DialUnix("unix", nil, unixAddr)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer conn.Close()

// 	// Send data to the server
// 	message := "Hello Unix Domain Socket!"
// 	_, err = conn.Write([]byte(message))
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Read response from the server
// 	buffer := make([]byte, len(message))
// 	_, err = conn.Read(buffer)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
