# URI Handler Package

The uriHandler package provides a unified interface for handling different types of URI-based data streams including TCP, UDP, and file operations. This package enables applications to interact seamlessly with various data sources and destinations, supporting a wide range of use cases from network communication to file management and web streaming.

## Installation

To install the library, use `go get github.com/Channel-3-Eugene/tribd/urihandler`.

To use the uriHandler package, import it into your Go project:

```go
import "github.com/Channel-3-Eugene/tribd/urihandler"
```

## Features

### Handlers provided 

- File Handler: Perform file operations, including reading from and writing to files, with support for named pipes (FIFOs).
- IPC Handler: Send and receive data using Unix domain sockets.
- UDP Handler: Handle UDP data transmission with support for both sending and receiving data.
- TCP Handler: Manage TCP connections for sending and receiving data.

### Broader features

- Flexible Data Streaming: The handler is capable of continuous data reading and writing, making it suitable for streaming applications. Data is handled through channels, allowing for smooth integration with concurrent Go routines and operations.
- Configurable Timeouts: Users can specify read and write timeouts, providing control over blocking operations. This feature is critical for ensuring responsiveness in systems where timely data processing is essential.
- Non-Blocking Options: The handler can be configured to operate in a non-blocking mode, particularly useful when working with named pipes. This prevents the handler from being stuck in operations where no data is available or no recipients are ready to receive data.
- Role-Based Functionality: The handler operates based on specified roles — either as a 'reader' or a 'writer', tailoring its behavior to fit the needs of the application, whether it's consuming or producing data.

### File Handler

The FileHandler within the uriHandler package is designed to handle various file operations in a unified and efficient manner. It supports reading from and writing to different types of file-like endpoints, which makes it highly versatile for applications that require handling standard files, named pipes (FIFOs), and potentially other special file types.

- Standard File Operations: The handler can open, read from, and write to plain files stored on disk. This is useful for applications that need to process or generate data stored in a file system.
- Standard Streams: The handler will open, read from, and write to standard streams such as stdin, stdout, and numbered file descriptors beyond stderr.
- Unix Domain Sockets: The URI handler can create, read from, write to, and destroy Unix domain sockets (also called Interprocess Communication sockets.)
- Named Pipes (FIFOs): It offers support for named pipes, which allows for inter-process communication using file-like interfaces. This is particularly beneficial in scenarios where processes need to exchange data in real-time without the overhead of network communication.

#### Usage

The FileHandler can be instantiated with parameters that define its role, the path of the file or named pipe, and settings for timeouts and blocking behavior. Here's a simple example of how to set up and use the FileHandler:

```go
package main

import (
    "github.com/Channel-3-Eugene/tribd/urihandler"
    "time"
)

func main() {
    dataChan := make(chan []byte)

    // Create a FileHandler to read from a named pipe
    fileHandler := uriHandler.NewFileHandler("/tmp/myfifo", uriHandler.Reader, true, dataChan, 0, 0)

    if err := fileHandler.Open(); err != nil {
        panic(err)
    }

    // Use a separate goroutine to handle incoming data
    go func() {
        for data := range dataChan {
            process(data)
        }
    }()

    // Close the handler when done
    defer fileHandler.Close()
}

func process(data []byte) {
    // Process data received from the file or named pipe
}
```

#### Integration

The FileHandler is designed to be easily integrated into larger systems that require file-based data input/output, making it an essential tool for applications ranging from data processing pipelines to system utilities that need to interact with the file system or other processes via named pipes.

### TCP Handler

#### Features

- Bidirectional Communication: The TCP Handler supports full-duplex communication, allowing data to be sent and received over the same connection.
- Connection Management: It handles the setup, maintenance, and teardown of TCP connections, ensuring reliable and ordered data transmission.
- Concurrency Support: The handler can manage multiple connections simultaneously, suitable for server applications that need to handle many clients.
- Configurable Timeouts: Users can set connection timeouts, read, and write timeouts to handle network delays and ensure responsiveness.
- Scalability: Built to efficiently handle high volumes of traffic and scalable across multiple cores, leveraging Go’s goroutines for concurrent connection handling.

#### Usage

The TCP Handler can be used to set up a client or server for TCP-based communication, handling connection establishment, data transmission, and connection closure cleanly within your application.

```go
package main

import (
    "github.com/Channel-3-Eugene/tribd/urihandler"
    "log"
)

func main() {
    dataChan := make(chan []byte, 1024) // Buffer size as needed

    // Setup a TCP server handler
    tcpHandler := uriHandler.NewTCPHandler("localhost:9999", uriHandler.Server, dataChan)
    if err := tcpHandler.Open(); err != nil {
        log.Fatal(err)
    }

    // Example of handling incoming data
    go func() {
        for data := range dataChan {
            log.Println("Received data:", string(data))
        }
    }()

    // Clean up on exit
    defer tcpHandler.Close()
}
```

### UDP Handler

#### Features

- Non-connection-based Communication: Unlike TCP, UDP does not establish a connection, which means it can send messages with lower initial latency. However, for ongoing transmission, TCP can achieve similar or even lower latency due to network optimizations such as large send offload and driver/hardware segmentation.
- Broadcast and Multicast: Supports broadcasting messages to multiple recipients and multicasting to a selected group of listeners.
- Lightweight Protocol: Ideal for applications that require fast, efficient communication, such as real-time services.
- Configurable Buffer Sizes: Allows adjustment of read and write buffer sizes to optimize for throughput or memory usage.

#### Usage

The UDP Handler is used for sending and receiving datagrams over UDP, supporting both unicast and multicast transmissions.

```go
package main

import (
    "github.com/Channel-3-Eugene/tribd/urihandler"
    "log"
)

func main() {
    dataChan := make(chan []byte, 1024) // Buffer size as needed

    // Setup a UDP endpoint
    udpHandler := uriHandler.NewUDPHandler("localhost:9998", uriHandler.Reader, dataChan)
    if err := udpHandler.Open(); err != nil {
        log.Fatal(err)
    }

    // Example of handling incoming data
    go func() {
        for data := range dataChan {
            log.Println("Received data:", string(data))
        }
    }()

    // Clean up on exit
    defer udpHandler.Close()
}
```

## Tests

A comprehensive test suite is provided, which may also be used as a reference for usage.

To run the test suite:

```bash
cd urihandler
go test . --race
```

### License

This project is licensed under the BSD 3 clause license.
