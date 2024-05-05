package mpegts

import (
	"crypto/rand"
	"errors"
)

// PIDs for video, audio, and data streams
const (
	packetLength = 188
	headerLength = 4
	MaxPCRValue  = (1 << 33) - 1

	VideoPID = 0x101
	AudioPID = 0x102
	DataPID  = 0x103
)

// Generator with tests for MPEG-TS packets
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
		packet.SetPCR(pcr)

		// Calculate and set adaptation field length
		adaptationFieldLength := calculateAdaptationFieldLength(&packet)
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
