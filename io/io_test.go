package main

import (
	"net"
	"testing"
	"time"

	"github.com/Channel-3-Eugene/tribd/config"
	"github.com/Channel-3-Eugene/tribd/mpegts"

	"github.com/stretchr/testify/assert"
)

// Function to start the server and return a channel for sending batches of packets
func setupTestServer(t *testing.T, port int) (chan []mpegts.EncodedPacket, func()) {
	addr := net.UDPAddr{
		IP:   net.ParseIP("::1"),
		Port: port,
	}
	conn, err := net.DialUDP("udp6", nil, &addr)
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	packetChan := make(chan []mpegts.EncodedPacket)

	go func() {
		defer conn.Close()
		for packets := range packetChan {
			buffer := make([]byte, 0, len(packets)*188)
			for _, packet := range packets {
				buffer = append(buffer, packet[:]...)
			}
			if _, err := conn.Write(buffer); err != nil {
				t.Logf("Failed to send batch of packets: %v", err)
			}
		}
	}()

	return packetChan, func() {
		close(packetChan)
		conn.Close()
	}
}

func TestInputStream_Integration(t *testing.T) {
	packets, err := mpegts.GenerateMPEGTSPackets(10)
	if err != nil {
		t.Fatalf("Error generating MPEG-TS packets: %v", err)
	}

	inputStream := NewInputStream(config.ReaderConfig{
		IPAddress: "::1",
		Port:      8788,
		ID:        "test",
	})

	ch := make(chan *mpegts.EncodedPacket)
	done := make(chan struct{})
	defer close(done)

	go inputStream.Start(ch, done)
	packetChan, cleanup := setupTestServer(t, 8788)
	defer cleanup()

	// Send data to the server
	packetChan <- packets

	receivedPackets := make([]*mpegts.EncodedPacket, 0)
	timeout := time.After(5 * time.Second)
	for i := 0; i < len(packets); i++ {
		select {
		case packet := <-ch:
			receivedPackets = append(receivedPackets, packet)
		case <-timeout:
			t.Fatal("Test timed out waiting for packets")
		}
	}

	assert.Len(t, receivedPackets, len(packets), "Expected number of packets")
}
