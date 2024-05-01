package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	// read and parse config
	var c Config
	closer, err := c.Read(...)
	if err != nil {
		// handle error
	}

	// Close the resources when main exits
	defer closer()

	// Create a waitgroup to synchronize goroutines
	var wg sync.WaitGroup

	// Create a channel for receiving OTW packets from readers
	otwPacketCh := make(chan OTWTPacket)

	// Create a channel to handle OS signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Create a slice to collect the channels for stopping readers
	stopChannels := make([]chan struct{}, len(c.Readers))

	// Start a goroutine to close channels when all waitgroup members are gone
	go func() {
		wg.Wait()
		close(otwPacketCh)
	}()

	// Start a goroutine to handle OS signals
	go func() {
		for {
			select {
			case sig := <-sigCh:
				switch sig {
				case syscall.SIGINT, syscall.SIGTERM:
					// Shutdown gracefully
					for _, stopCh := range stopChannels {
						close(stopCh)
					}
					return
				case syscall.SIGHUP:
					// Handle SIGHUP (reload config, etc.)
				}
			}
		}
	}()

	// Start a goroutine for each reader config in c.Readers
	for _, readerConfig := range c.Readers {
		stopChannel := make(chan struct{}) // Create a stop channel for this reader
		stopChannels = append(stopChannels, stopChannel)
		for _, reader := range c.Readers {
			wg.Add(1)
			reader := NewInputStream(c.ReaderConfig)
			go func(reader InputStream, stopCh <-chan struct{}) {
				defer wg.Done()
				reader.Start(otwPacketCh, stopCh)
			}(reader, stopChannel)
		}
	}

	writer := NewWriter(...)
	// Switch for receiving OTW packets and writing them with a writer defined in c.Writer
	for {
		select {
		case packet := <-otwPacketCh:
			// Write the packet to a file
			if err := writer.Write(packet); err != nil {
				// handle error
			}
		}
	}
}
