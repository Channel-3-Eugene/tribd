package channels

import (
	"bytes"
	"sync"
	"testing"
	"time"
)

func TestPacketChan_SendReceive(t *testing.T) {
	pc := NewPacketChan(10)

	// Data to send
	expected := []byte("test data")

	// Send data
	err := pc.Send(expected)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	// Receive data
	received := pc.Receive()
	if !bytes.Equal(received, expected) {
		t.Errorf("Expected %v, got %v", expected, received)
	}

	// Clean up
	pc.Close()
}

func TestPacketChan_ConcurrentAccess(t *testing.T) {
	pc := NewPacketChan(100)
	wg := sync.WaitGroup{}
	wg.Add(2)

	// Concurrent sender
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			err := pc.Send([]byte{byte(i)})
			if err != nil {
				t.Errorf("Send failed: %v", err)
				return
			}
		}
	}()

	// Concurrent receiver
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			data := pc.Receive()
			if data == nil {
				t.Errorf("Receive failed: channel closed")
				return
			}
		}
	}()

	wg.Wait()

	// Clean up
	pc.Close()
}

func TestPacketChan_Close(t *testing.T) {
	pc := NewPacketChan(10)

	// Close the channel before the test
	pc.Close()

	// Attempt to send data should fail
	err := pc.Send([]byte("data"))
	if err == nil {
		t.Errorf("Send did not fail on closed channel")
	}

	// Attempt to receive data should not hang and should return nil
	select {
	case data := <-pc.ch:
		if data != nil {
			t.Errorf("Expected nil, got %v", data)
		}
	case <-time.After(1 * time.Second):
		t.Errorf("Receive hanged on closed channel")
	}
}

func TestPacketChan_StressTest(t *testing.T) {
	pc := NewPacketChan(100)
	wg := sync.WaitGroup{}
	data := []byte("data")
	const workers = 10

	// Start multiple senders
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				for {
					if err := pc.Send(data); err != nil {
						if err.Error() == "failed to send data: buffer full" {
							time.Sleep(10 * time.Millisecond) // Backoff
							continue
						}
						t.Errorf("Send failed: %v", err)
						return
					}
					break
				}
			}
		}()
	}

	// Start multiple receivers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				data := pc.Receive()
				if data == nil {
					t.Errorf("Receive failed: channel closed")
					return
				}
			}
		}()
	}

	wg.Wait()

	// Clean up
	pc.Close()
}
