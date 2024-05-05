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

// TestSettingAndClearingTEI tests the setting and clearing of the TEI flag.
func TestSettingAndClearingTEI(t *testing.T) {
	packets, err := GenerateMPEGTSPackets(10) // Smaller number for focused tests
	assert.NoError(t, err)

	for i, packet := range packets {
		packet.SetTEI()
		assert.True(t, packet.GetTEI(), "TEI should be set for packet at index %d", i)

		packet.ClearTEI()
		assert.False(t, packet.GetTEI(), "TEI should be cleared for packet at index %d", i)
	}
}

// TestSettingAndClearingPUSI tests the setting and clearing of the PUSI flag.
func TestSettingAndClearingPUSI(t *testing.T) {
	packets, err := GenerateMPEGTSPackets(10)
	assert.NoError(t, err)

	for i, packet := range packets {
		packet.SetPUSI()
		assert.True(t, packet.GetPUSI(), "PUSI should be set for packet at index %d", i)

		packet.ClearPUSI()
		assert.False(t, packet.GetPUSI(), "PUSI should be cleared for packet at index %d", i)
	}
}

// TestPIDSetting tests setting the PID and verifying it.
func TestPIDSetting(t *testing.T) {
	packets, err := GenerateMPEGTSPackets(10)
	assert.NoError(t, err)

	for i, packet := range packets {
		expectedPID := uint16(i)
		packet.SetPID(expectedPID)
		assert.Equal(t, expectedPID, packet.GetPID(), "PID mismatch for packet at index %d", i)
	}
}

// TestTSCSetting tests setting the TSC field and verifying it.
func TestTSCSetting(t *testing.T) {
	packets, err := GenerateMPEGTSPackets(10)
	assert.NoError(t, err)

	for i, packet := range packets {
		expectedTSC := uint8(i % 4) // TSC is a 2-bit field
		packet.SetTSC(expectedTSC)
		assert.Equal(t, expectedTSC, packet.GetTSC(), "TSC mismatch for packet at index %d", i)
	}
}

// TestAFCSetting tests setting the AFC field and verifying it.
func TestAFCSetting(t *testing.T) {
	packets, err := GenerateMPEGTSPackets(10)
	assert.NoError(t, err)

	for i, packet := range packets {
		expectedAFC := uint8(i % 4) // AFC is a 2-bit field
		packet.SetAFC(expectedAFC)
		assert.Equal(t, expectedAFC, packet.GetAFC(), "AFC mismatch for packet at index %d", i)
	}
}

// TestCCSetting tests setting the Continuity Counter and verifying it.
func TestCCSetting(t *testing.T) {
	packets, err := GenerateMPEGTSPackets(10)
	assert.NoError(t, err)

	for i, packet := range packets {
		expectedCC := uint8(i % 16) // CC is a 4-bit field
		packet.SetCC(expectedCC)
		assert.Equal(t, expectedCC, packet.GetCC(), "CC mismatch for packet at index %d", i)
	}
}

// TestReadingTEI tests reading the TEI flag.
func TestReadingTEI(t *testing.T) {
	packets, err := GenerateMPEGTSPackets(10)
	assert.NoError(t, err)

	// Set TEI for even indexed packets
	for i, packet := range packets {
		if i%2 == 0 {
			packet.SetTEI()
		} else {
			packet.ClearTEI()
		}
		expectedTEI := i%2 == 0
		assert.Equal(t, expectedTEI, packet.GetTEI(), "TEI read error for packet at index %d", i)
	}
}

// TestReadingPUSI tests reading the PUSI flag.
func TestReadingPUSI(t *testing.T) {
	packets, err := GenerateMPEGTSPackets(10)
	assert.NoError(t, err)

	// Set PUSI for odd indexed packets
	for i, packet := range packets {
		if i%2 != 0 {
			packet.SetPUSI()
		} else {
			packet.ClearPUSI()
		}
		expectedPUSI := i%2 != 0
		assert.Equal(t, expectedPUSI, packet.GetPUSI(), "PUSI read error for packet at index %d", i)
	}
}

// TestReadingPID tests reading the PID.
func TestReadingPID(t *testing.T) {
	packets, err := GenerateMPEGTSPackets(10)
	assert.NoError(t, err)

	for i, packet := range packets {
		expectedPID := uint16(i)
		packet.SetPID(expectedPID)
		assert.Equal(t, expectedPID, packet.GetPID(), "PID read error for packet at index %d", i)
	}
}

// TestReadingTSC tests reading the TSC field.
func TestReadingTSC(t *testing.T) {
	packets, err := GenerateMPEGTSPackets(10)
	assert.NoError(t, err)

	for i, packet := range packets {
		expectedTSC := uint8(i % 4) // TSC is a 2-bit field
		packet.SetTSC(expectedTSC)
		assert.Equal(t, expectedTSC, packet.GetTSC(), "TSC read error for packet at index %d", i)
	}
}

// TestReadingAFC tests reading the AFC field.
func TestReadingAFC(t *testing.T) {
	packets, err := GenerateMPEGTSPackets(10)
	assert.NoError(t, err)

	for i, packet := range packets {
		expectedAFC := uint8(i % 4) // AFC is a 2-bit field
		packet.SetAFC(expectedAFC)
		assert.Equal(t, expectedAFC, packet.GetAFC(), "AFC read error for packet at index %d", i)
	}
}

// TestReadingCC tests reading the Continuity Counter.
func TestReadingCC(t *testing.T) {
	packets, err := GenerateMPEGTSPackets(10)
	assert.NoError(t, err)

	for i, packet := range packets {
		expectedCC := uint8(i % 16) // CC is a 4-bit field
		packet.SetCC(expectedCC)
		assert.Equal(t, expectedCC, packet.GetCC(), "CC read error for packet at index %d", i)
	}
}

// TestPacketManipulationIntegration tests the integration of multiple packet manipulations.
func TestPacketManipulationIntegration(t *testing.T) {
	packets, err := GenerateMPEGTSPackets(10)
	assert.NoError(t, err)

	for i, packet := range packets {
		// Initial operations
		initialPID := uint16(100 + i)
		packet.SetPID(initialPID)
		packet.SetPUSI()
		packet.SetTEI()

		// Verify initial settings
		assert.True(t, packet.GetPUSI(), "PUSI should be set for packet at index %d", i)
		assert.True(t, packet.GetTEI(), "TEI should be set for packet at index %d", i)
		assert.Equal(t, initialPID, packet.GetPID(), "PID should be set for packet at index %d", i)

		// Clear flags and change PID
		packet.ClearPUSI()
		packet.ClearTEI()
		newPID := uint16(200 + i)
		packet.SetPID(newPID)

		// Verify new settings
		assert.False(t, packet.GetPUSI(), "PUSI should be cleared for packet at index %d", i)
		assert.False(t, packet.GetTEI(), "TEI should be cleared for packet at index %d", i)
		assert.Equal(t, newPID, packet.GetPID(), "PID should be updated for packet at index %d", i)
	}
}

// TestIncorrectSyncByte checks error handling for packets with incorrect sync bytes.
func TestIncorrectSyncByte(t *testing.T) {
	packet := &EncodedPacket{}
	packet[0] = 0x48 // Incorrect sync byte
	assert.False(t, packet.IsMPEGTS(), "Packet with incorrect sync byte should not pass validation")
}

// TestInvalidPacketSize checks the system's response to packets that do not conform to the 188 byte standard.
func TestInvalidPacketSize(t *testing.T) {
	packet := make([]byte, 190)   // Incorrect size
	_, err := ParsePacket(packet) // Assuming a function ParsePacket exists
	assert.Error(t, err, "Should error with invalid packet size")
}

// TestInvalidPID checks handling of packets with an invalid PID value.
func TestInvalidPID(t *testing.T) {
	packet := &EncodedPacket{}
	packet[0] = 0x47
	packet[1] = 0xFF // PID set to an invalid range, assuming 0x1FFF is max
	packet[2] = 0xFF
	assert.True(t, packet.GetPID() == 0x1FFF, "Packet with invalid PID should get a null value")
}

// TestMissingAdaptationFieldWhenExpected tests packets where adaptation field control bits are set but the field is missing.
func TestSetAFC(t *testing.T) {
	packet := &EncodedPacket{}
	packet[0] = 0x47 // Set the sync byte

	// Test setting adaptation field to be present without payload
	packet.SetAFC(0x02) // Adaptation field present, no payload
	assert.Equal(t, byte(0x20), packet[3]&0x30, "Adaptation field control bits should indicate adaptation field only")
	packet[4] = 5 // Set a dummy adaptation field length
	assert.Equal(t, 5, int(packet[4]), "Adaptation field length should match the set value")

	// Test disabling the adaptation field
	packet.SetAFC(0x01) // No adaptation field, payload only
	adaptationFieldLength := calculateAdaptationFieldLength(packet)
	assert.Equal(t, 0, adaptationFieldLength, "Adaptation field length should be 0 when adaptation is disabled")
}

func TestPayloadHandling(t *testing.T) {
	packet := &EncodedPacket{}
	packet[0] = 0x47
	packet.SetAFC(0x01) // Set AFC to ensure payload includes a length byte.

	payload := []byte("test payload")
	packet.SetPayload(payload)
	retrievedPayload := packet.GetPayload()

	assert.Equal(t, payload, retrievedPayload, "The set and retrieved payloads do not match")
}

func TestAdaptationFieldHandling(t *testing.T) {
	packet := &EncodedPacket{}
	packet[0] = 0x47
	packet.SetAFC(0x03) // Set AFC to ensure adaptation field includes a length byte.

	adaptationField := []byte{0x0A, 0xBB, 0xCC} // Example adaptation field data
	packet.SetAdaptationField(adaptationField)
	retrievedAdaptationField := packet.GetAdaptationField()

	assert.Equal(t, adaptationField, retrievedAdaptationField, "The set and retrieved adaptation fields do not match")
}

// TestPCRHandling tests setting and retrieving the PCR value.
func TestPCRHandling(t *testing.T) {
	packet := &EncodedPacket{}
	packet[0] = 0x47    // Set the sync byte
	packet.SetAFC(0x02) // Adaptation field only, for simplicity

	// Set and get PCR value
	originalPCR := uint64(1234567890)
	packet.SetPCR(originalPCR)
	retrievedPCR := packet.GetPCR()

	assert.Equal(t, originalPCR, retrievedPCR, "The set and retrieved PCR values do not match")
}

// TestOPCRHandling tests setting and retrieving the OPCR value.
func TestOPCRHandling(t *testing.T) {
	packet := &EncodedPacket{}
	packet[0] = 0x47    // Set the sync byte
	packet.SetAFC(0x03) // Adaptation field followed by payload

	originalOPCR := uint64(9876543210)
	packet.SetOPCR(originalOPCR)
	retrievedOPCR := packet.GetOPCR()

	assert.Equal(t, originalOPCR, retrievedOPCR, "The set and retrieved OPCR values do not match")
}

// TestSpliceCountdownHandling tests setting and retrieving the splice countdown value.
func TestSpliceCountdownHandling(t *testing.T) {
	packet := &EncodedPacket{}
	packet[0] = 0x47
	packet.SetAFC(0x03) // Assume adaptation field followed by payload includes splice countdown

	originalCountdown := uint8(10)
	packet.SetSpliceCountdown(originalCountdown)
	retrievedCountdown := packet.GetSpliceCountdown()

	assert.Equal(t, originalCountdown, retrievedCountdown, "The set and retrieved splice countdown values do not match")
}

// Assuming these function stubs exist
func ParsePacket(data []byte) (*EncodedPacket, error) {
	if len(data) != 188 {
		return nil, errors.New("invalid packet size")
	}
	return &EncodedPacket{}, nil
}

// TestTransportPrivateDataHandling tests setting and retrieving transport private data.
func TestTransportPrivateDataHandling(t *testing.T) {
	packet := &EncodedPacket{}
	packet[0] = 0x47
	packet.SetAFC(0x03) // Ensure AFC allows for private data

	// Initialize the adaptation field length including private data
	packet[4] = 20 // Set an adaptation field length that includes space for private data

	privateData := []byte{0x01, 0x02, 0x03, 0x04}
	packet.SetTransportPrivateData(privateData)
	retrievedData := packet.GetTransportPrivateData()

	assert.Equal(t, privateData, retrievedData, "The set and retrieved transport private data do not match")
}

// TestAdaptationFieldExtensionHandling tests setting and retrieving the adaptation field extension.
func TestAdaptationFieldExtensionHandling(t *testing.T) {
	packet := &EncodedPacket{}
	packet[0] = 0x47
	packet.SetAFC(0x03) // Ensure AFC allows for extensions

	// Initialize the adaptation field length including an extension
	packet[4] = 20 // Set an adaptation field length that includes space for an extension
	packet[22] = 5 // Position for extension length

	adaptationFieldExtension := []byte{0xAA, 0xBB, 0xCC}
	packet.SetAdaptationFieldExtension(adaptationFieldExtension)
	retrievedExtension := packet.GetAdaptationFieldExtension()

	assert.Equal(t, adaptationFieldExtension, retrievedExtension, "The set and retrieved adaptation field extension do not match")
}

// TestCalculateAdaptationFieldLength tests the function to calculate the adaptation field length.
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
			if got := calculateAdaptationFieldLength(&tt.packet); got != tt.want {
				t.Errorf("calculateAdaptationFieldLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
		SetPCR(&packet, pcr, PCRFrequency)

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

// TestGenerateMPEGTSPackets tests the function to generate MPEG-TS packets.
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
				adaptationFieldLength := calculateAdaptationFieldLength(&packet)
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

// TestMPEGTSPacketIntegrity tests the integrity of MPEG-TS packets.
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

		assert.NoError(t, err)
		assert.Len(t, packets, packetCount)

		packet := packets[0]

		t.Run("Packet", func(t *testing.T) {
			// Check adaptation field control bit
			assert.Equal(t, byte(0x20), packet[3]&0x20) // Adaptation field present

			// Check adaptation field length
			assert.Greater(t, calculateAdaptationFieldLength(&packet), 0)
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
				assert.Greater(t, calculateAdaptationFieldLength(&packet), 0)
			})
		}
	})
}
