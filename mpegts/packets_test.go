package mpegts

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// PIDs for video, audio, and data streams
const (
	packetLength = 188
	headerLength = 4

	VideoPID = 0x101
	AudioPID = 0x102
	DataPID  = 0x103
)

// PCR constants
const (
	// PCR frequency in Hz
	PCRFrequency = 27000000
	// Maximum PCR value
	MaxPCRValue = (1 << 33) - 1
)

func TestCalculateAdaptationFieldLength(t *testing.T) {
	tests := []struct {
		name   string
		packet EncodedPacket
		want   int
	}{
		{
			name:   "NoAdaptationField",
			packet: EncodedPacket{0x47, 0x40, 0x00, 0x10},
			want:   0,
		},
		{
			name:   "AdaptationFieldPresent",
			packet: EncodedPacket{0x47, 0x40, 0x00, 0x20, 0x05, 0x00, 0x00, 0x00, 0x00},
			want:   6, // Length of adaptation field is 5, plus 1 for the length field itself
		},
		{
			name:   "AdaptationFieldWithPCR",
			packet: EncodedPacket{0x47, 0x40, 0x00, 0x30, 0x0A, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			want:   11, // Length of adaptation field is 10, plus 1 for the length field itself
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateAdaptationFieldLength(tt.packet); got != tt.want {
				t.Errorf("calculateAdaptationFieldLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

// GenerateMPEGTSPackets generates a series of MPEG-TS packets representing a section of a stream containing one PES across multiple TS packets.
func GenerateMPEGTSPackets(count int) ([]EncodedPacket, error) {
	if count < 1 {
		return nil, errors.New("count must be greater than 0")
	}

	// Randomly select a PID from video, audio, or data types
	pids := []uint16{VideoPID, AudioPID, DataPID}
	pid := pids[randIntn(len(pids))]

	packets := make([]EncodedPacket, count)

	// Calculate PCR increment
	pcrIncrement := MaxPCRValue / uint64(count)

	// Generate packets
	for i := 0; i < count; i++ {
		packet := EncodedPacket{}

		// Set sync byte
		packet[0] = 0x47

		// Set adaptation field control bits (adaptation field present, payload present)
		packet[3] = 0x30 // Adaptation field present, payload present

		// Set PID
		packet[1] = byte(pid >> 8)   // Set PID high byte
		packet[2] = byte(pid & 0xFF) // Set PID low byte

		// Set PUSI bit only on the first packet
		if i == 0 {
			packet[1] |= 0x40
		}

		// Set continuity counter
		packet[3] |= byte(i & 0x0F)

		// Set PCR value
		pcr := uint64(i) * pcrIncrement
		SetPCR(&packet, pcr, PCRFrequency)

		// Calculate and set adaptation field length
		adaptationFieldLength := calculateAdaptationFieldLength(packet)
		packet[4] = byte(adaptationFieldLength - 1) // Set adaptation field length byte

		// Generate random payload
		payloadLength := packetLength - headerLength - adaptationFieldLength
		payload := make([]byte, payloadLength)
		if _, err := rand.Read(payload); err != nil {
			return nil, err
		}
		copy(packet[headerLength+adaptationFieldLength:], payload)

		packets[i] = packet
	}

	return packets, nil
}

// randIntn returns a random integer in the range [0, n)
func randIntn(n int) int {
	if n <= 0 {
		panic("invalid argument to randIntn")
	}
	b := make([]byte, 1)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return int(b[0]) % n
}

func TestGenerateMPEGTSPackets(t *testing.T) {
	t.Run("GenerateZeroPackets", func(t *testing.T) {
		count := 0

		packets, err := GenerateMPEGTSPackets(count)

		assert.Error(t, err)
		assert.Equal(t, 0, len(packets))
	})

	t.Run("GenerateSmallNumberOfPackets", func(t *testing.T) {
		count := 3

		packets, err := GenerateMPEGTSPackets(count)

		assert.NoError(t, err)
		assert.Len(t, packets, count)

		for i, packet := range packets {
			t.Run(fmt.Sprintf("Packet %d", i), func(t *testing.T) {
				assert.Len(t, packet, packetLength)
				assert.Equal(t, byte(0x47), packet[0])                                                                 // Check sync byte
				assert.Contains(t, []uint16{VideoPID, AudioPID, DataPID}, binary.BigEndian.Uint16(packet[1:3])&0x1FFF) // Check PID
				if i == 0 {
					assert.Equal(t, byte(0x40), packet[1]&0x40) // Check PUSI bit is set to 1 on the first packet
				} else {
					assert.Equal(t, byte(0x00), packet[1]&0x40) // Check PUSI bit is set to 0 on subsequent packets
				}
				assert.Equal(t, byte(i&0x0F), packet[3]&0x0F) // Check continuity counter

				// Check adaptations (if present)
				adaptationFieldLength := calculateAdaptationFieldLength(packet)
				if adaptationFieldLength > 0 {
					assert.True(t, packet[3]&0x20 != 0)                       // Check adaptation field control bit
					assert.Equal(t, byte(adaptationFieldLength-1), packet[4]) // Check adaptation field length
				} else {
					assert.False(t, packet[3]&0x20 != 0) // Check adaptation field control bit
					assert.Len(t, packet, packetLength)  // Check total length
				}
			})
		}
	})
}

func TestMPEGTSPacketIntegrity(t *testing.T) {
	t.Run("GeneratePacketsWithContinuityCounterWrapAround", func(t *testing.T) {
		count := 20

		packets, err := GenerateMPEGTSPackets(count)

		assert.NoError(t, err)
		assert.Len(t, packets, count)

		lastContinuityCounter := -1
		for i, packet := range packets {
			t.Run(fmt.Sprintf("Packet %d", i), func(t *testing.T) {
				// Get continuity counter
				continuityCounter := int(packet[3] & 0x0F)

				// Check continuity counter reset after wrap around
				if i > 0 && continuityCounter < lastContinuityCounter {
					assert.Equal(t, 0, continuityCounter)
				}

				lastContinuityCounter = continuityCounter
			})
		}
	})

	t.Run("GeneratePacketsWithIncrementingContinuityCounter", func(t *testing.T) {
		count := 20
		packets, err := GenerateMPEGTSPackets(count)

		assert.NoError(t, err)
		assert.Len(t, packets, count)

		var lastPID uint16
		lastContinuityCounter := -1
		for i, packet := range packets {
			t.Run(fmt.Sprintf("Packet %d", i), func(t *testing.T) {
				// Get continuity counter
				continuityCounter := int(packet[3] & 0x0F)

				// Check incrementing continuity counter
				if i > 0 {
					expectedContinuityCounter := (lastContinuityCounter + 1) % 16
					assert.Equal(t, expectedContinuityCounter, continuityCounter)
				}

				// Check PID consistency
				if continuityCounter == 0 && lastContinuityCounter != -1 {
					// Continuity counter rolled over, PID should remain the same unless PUSI bit is set
					if packet[1]&0x40 != 0 {
						// PUSI bit is set, PID might change
						assert.NotEqual(t, lastPID, binary.BigEndian.Uint16(packet[1:3])&0x1FFF)
					} else {
						// PUSI bit is not set, PID should remain the same
						assert.Equal(t, lastPID, binary.BigEndian.Uint16(packet[1:3])&0x1FFF)
					}
				}

				// Store current PID for comparison in the next iteration
				lastPID = binary.BigEndian.Uint16(packet[1:3]) & 0x1FFF

				lastContinuityCounter = continuityCounter
			})
		}
	})

	t.Run("GenerateSinglePacketWithAdaptation", func(t *testing.T) {
		packetCount := 1

		packets, err := GenerateMPEGTSPackets(packetCount)
		println("Single packet", packets)

		assert.NoError(t, err)
		assert.Len(t, packets, packetCount)

		packet := packets[0]

		t.Run("Packet", func(t *testing.T) {
			// Check adaptation field control bit
			assert.Equal(t, byte(0x20), packet[3]&0x20) // Adaptation field present

			// Check adaptation field length
			assert.Greater(t, calculateAdaptationFieldLength(packet), 0)
		})
	})

	t.Run("GenerateMultiplePacketsWithAdaptation", func(t *testing.T) {
		packetCount := 3

		packets, err := GenerateMPEGTSPackets(packetCount)

		assert.NoError(t, err)
		assert.Len(t, packets, packetCount)

		for i, packet := range packets {
			t.Run(fmt.Sprintf("Packet %d", i), func(t *testing.T) {
				// Check adaptation field control bit
				assert.Equal(t, byte(0x20), packet[3]&0x20) // Adaptation field present

				// Check adaptation field length
				assert.Greater(t, calculateAdaptationFieldLength(packet), 0)
			})
		}
	})

}
