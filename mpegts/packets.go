// Package mpegts implements functions to parse and manipulate MPEG-TS packets.
package mpegts

import (
	"encoding/binary"
	"errors"
)

// Error constants define common errors encountered during MPEG-TS parsing.
var (
	ErrInvalidSyncByte         = errors.New("mpegts: invalid sync byte")
	ErrInvalidPacketSize       = errors.New("mpegts: invalid packet size")
	ErrInvalidHeader           = errors.New("mpegts: invalid packet header")
	ErrInvalidAdaptationsField = errors.New("mpegts: invalid packet adaptations field")
	ErrInvalidPayload          = errors.New("mpegts: invalid packet payload")
	ErrUnexpectedEOF           = errors.New("mpegts: unexpected end of file")
	ErrInvalidPID              = errors.New("mpegts: invalid PID")
	ErrPATNotFound             = errors.New("mpegts: PAT not found")
	ErrPMTNotFound             = errors.New("mpegts: PMT not found")
	ErrProgramNotFound         = errors.New("mpegts: program not found")
	ErrStreamNotFound          = errors.New("mpegts: stream not found")
	ErrUnsupportedStream       = errors.New("mpegts: unsupported stream type")
)

// EncodedPacket represents a raw MPEG-TS packet.
type EncodedPacket [188]byte

// EncodedPackets represents a collection of raw MPEG-TS packets.
type EncodedPackets []*EncodedPacket

// NewMPEGTSPacket creates a new MPEG-TS packet from a 188-byte array.
// It validates the packet to ensure it starts with the correct sync byte (0x47).
func NewMPEGTSPacket(data [188]byte) (*EncodedPacket, error) {
	if data[0] != 0x47 {
		return nil, ErrInvalidSyncByte // Ensure sync byte is correct
	}
	packet := EncodedPacket(data) // Directly use the data as EncodedPacket
	return &packet, nil
}

// IsMPEGTS checks if the packet begins with the MPEG-TS sync byte.
func (ep *EncodedPacket) IsMPEGTS() bool {
	return ep[0] == 0x47
}

// GetSyncByte returns the sync byte of the MPEG-TS packet.
func (ep *EncodedPacket) GetSyncByte() byte {
	return ep[0]
}

// GetTEI returns the Transport Error Indicator (TEI) flag of the packet.
func (ep *EncodedPacket) GetTEI() bool {
	return ep[1]&0x80 != 0
}

// SetTEI sets the Transport Error Indicator (TEI) flag of the packet.
func (ep *EncodedPacket) SetTEI() {
	ep[1] |= 0x80
}

// ClearTEI clears the Transport Error Indicator (TEI) flag of the packet.
func (ep *EncodedPacket) ClearTEI() {
	ep[1] &= 0x7F
}

// GetPUSI returns the Payload Unit Start Indicator (PUSI) flag of the packet.
func (ep *EncodedPacket) GetPUSI() bool {
	return ep[1]&0x40 != 0
}

// SetPUSI sets the Payload Unit Start Indicator (PUSI) flag of the packet.
func (ep *EncodedPacket) SetPUSI() {
	ep[1] |= 0x40
}

// ClearPUSI clears the Payload Unit Start Indicator (PUSI) flag of the packet.
func (ep *EncodedPacket) ClearPUSI() {
	ep[1] &= 0xBF
}

// GetPID returns the Packet Identifier (PID) of the packet.
func (ep *EncodedPacket) GetPID() uint16 {
	return binary.BigEndian.Uint16(ep[1:3]) & 0x1FFF
}

// SetPID sets the Packet Identifier (PID) of the packet.
func (ep *EncodedPacket) SetPID(pid uint16) {
	if pid > 0x1FFF { // Ensure the PID does not exceed 13 bits
		// Handle error or assign a default value; depends on your use case
		pid = 0x1FFF
	}

	// Clear the existing PID bits (13 bits spanning bytes 1 and 2)
	ep[1] &= 0xE0 // Preserve the upper 3 bits (PUSI and TP) of the first byte
	ep[2] = 0x00  // Clear the second byte (will be fully set below)

	// Set the new PID
	ep[1] |= byte(pid >> 8)   // Set the upper 5 bits of the PID
	ep[2] |= byte(pid & 0xFF) // Set the lower 8 bits of the PID
}

// IsNullPacket checks if the packet is a null packet (PID 0x1FFF).
func (ep *EncodedPacket) IsNullPacket() bool {
	pid := int(ep[1]&0x1F)<<8 + int(ep[2])
	return pid == 0x1FFF
}

// GetTSC returns the Transport Scrambling Control (TSC) field of the packet.
func (ep *EncodedPacket) GetTSC() uint8 {
	return (ep[3] >> 6) & 0x03
}

// SetTSC sets the Transport Scrambling Control (TSC) field of the packet.
func (ep *EncodedPacket) SetTSC(tsc uint8) {
	ep[3] = (ep[3] & 0x3F) | (tsc << 6)
}

// GetAFC returns the Adaptation Field Control (AFC) field of the packet.
func (ep *EncodedPacket) GetAFC() uint8 {
	return (ep[3] >> 4) & 0x03
}

// SetAFC sets the Adaptation Field Control (AFC) field of the packet.
func (ep *EncodedPacket) SetAFC(afc uint8) {
	currentAFC := ep.GetAFC()
	ep[3] = (ep[3] & 0xCF) | (afc << 4)

	// If the new AFC indicates no adaptation field, clear the adaptation field length and data.
	if afc == 0x00 || afc == 0x10 {
		if currentAFC == 0x02 || currentAFC == 0x03 {
			// Zero the length of the adaptation field
			ep[4] = 0
			// Optionally clear the rest of the adaptation field data to maintain packet integrity
			// Here, we fill the adaptation field space with 0xFF (commonly used stuffing byte)
			for i := 5; i < 188; i++ {
				ep[i] = 0xFF
			}
		}
	}
}

// GetCC returns the Continuity Counter (CC) of the packet.
func (ep *EncodedPacket) GetCC() uint8 {
	return ep[3] & 0x0F
}

// SetCC sets the Continuity Counter (CC) of the packet.
func (ep *EncodedPacket) SetCC(cc uint8) {
	ep[3] = (ep[3] & 0xF0) | (cc & 0x0F)
}

// GetPayload returns the payload of the packet.
func (ep *EncodedPacket) GetPayload() []byte {
	if ep.GetAFC() == 0x01 {
		length := int(ep[4])
		return ep[5 : 5+length]
	}
	return ep[4:]
}

// SetPayload sets the payload of the packet.
func (ep *EncodedPacket) SetPayload(payload []byte) {
	if ep.GetAFC() == 0x01 {
		ep[4] = byte(len(payload))
		copy(ep[5:], payload)
	} else {
		copy(ep[4:], payload)
	}
}

// GetAdaptationField returns the adaptation field of the packet.
func (ep *EncodedPacket) GetAdaptationField() []byte {
	if ep.GetAFC() == 0x02 || ep.GetAFC() == 0x03 {
		length := int(ep[4])
		return ep[5 : 5+length]
	}
	return nil
}

// SetAdaptationField sets the adaptation field of the packet.
func (ep *EncodedPacket) SetAdaptationField(af []byte) {
	if ep.GetAFC() == 0x02 || ep.GetAFC() == 0x03 {
		ep[4] = byte(len(af))
		copy(ep[5:], af)
	}
}

// GetOPCR returns the Original Program Clock Reference (OPCR) value from the adaptation field.
func (ep *EncodedPacket) GetOPCR() uint64 {
	if ep.GetAFC() == 0x03 {
		return binary.BigEndian.Uint64(ep[13:21]) & 0x1FFFFFFFFFFFF
	}
	return 0
}

// SetOPCR sets the Original Program Clock Reference (OPCR) value in the adaptation field.
func (ep *EncodedPacket) SetOPCR(opcr uint64) {
	if ep.GetAFC() == 0x03 {
		binary.BigEndian.PutUint64(ep[13:21], opcr)
	}
}

// GetSpliceCountdown returns the Splice Countdown field from the adaptation field.
func (ep *EncodedPacket) GetSpliceCountdown() uint8 {
	if ep.GetAFC() == 0x03 {
		return ep[21]
	}
	return 0
}

// SetSpliceCountdown sets the Splice Countdown field in the adaptation field.
func (ep *EncodedPacket) SetSpliceCountdown(sc uint8) {
	if ep.GetAFC() == 0x03 {
		ep[21] = sc
	}
}

// GetTransportPrivateData returns the Transport Private Data from the adaptation field.
func (ep *EncodedPacket) GetTransportPrivateData() []byte {
	if ep.GetAFC() == 0x03 {
		length := ep[22]          // Get the length of the transport private data
		return ep[23 : 23+length] // Correct slicing to exclude the length byte
	}
	return nil
}

// SetTransportPrivateData sets the Transport Private Data in the adaptation field.
func (ep *EncodedPacket) SetTransportPrivateData(tpd []byte) {
	if ep.GetAFC() == 0x03 {
		ep[22] = byte(len(tpd))
		copy(ep[23:], tpd)
	}
}

// GetAdaptationFieldExtension returns the Adaptation Field Extension from the adaptation field.
func (ep *EncodedPacket) GetAdaptationFieldExtension() []byte {
	if ep.GetAFC() == 0x03 {
		start := 22 + ep[22]                // Calculate the starting position correctly
		length := ep[start]                 // Get the length of the extension from the start position
		return ep[start+1 : start+1+length] // Correct slicing to exclude the length byte
	}
	return nil
}

// SetAdaptationFieldExtension sets the Adaptation Field Extension in the adaptation field.
func (ep *EncodedPacket) SetAdaptationFieldExtension(afe []byte) {
	if ep.GetAFC() == 0x03 {
		ep[22+ep[22]] = byte(len(afe))
		copy(ep[23+ep[22]:], afe)
	}
}

// calculateAdaptationFieldLength calculates the length of the adaptation field based on whether it is present.
func calculateAdaptationFieldLength(packet *EncodedPacket) int {
	adaptationFieldControl := packet[3] >> 4 & 0x03
	if adaptationFieldControl == 0x02 || adaptationFieldControl == 0x03 {
		// Adaptation field is present
		return int(packet[4]) + 1
	}
	return 0
}

// GetPCR returns the Program Clock Reference (PCR) value from the adaptation field.
// We are assuming a standard 27 MHz clock frequency for the PCR.
func (ep *EncodedPacket) SetPCR(pcr uint64) {
	pcrBase := pcr / 300      // Divide by 300 to convert to base unit
	pcrExtension := pcr % 300 // Modulus 300 to find the extension

	afc := ep.GetAFC()
	if afc != 0x02 && afc != 0x03 {
		ep.SetAFC(0x03) // Ensure the adaptation field is set to contain a PCR
	}

	// Ensure adaptation field length is enough to hold PCR (minimum 7 bytes)
	if ep[4] < 7 {
		ep[4] = 7
		for i := 5; i < 12; i++ {
			ep[i] = 0
		}
	}

	ep[5] |= 0x10 // Set the PCR flag

	// Set PCR base and extension
	ep[6] = byte(pcrBase >> 25)
	ep[7] = byte(pcrBase >> 17)
	ep[8] = byte(pcrBase >> 9)
	ep[9] = byte(pcrBase >> 1)
	ep[10] = (byte(pcrBase<<7) & 0x80) | (byte(pcrExtension>>8) & 0x01)
	ep[11] = byte(pcrExtension & 0xFF)
}

// GetPCR returns the Program Clock Reference (PCR) value from the adaptation field.
func (ep *EncodedPacket) GetPCR() uint64 {
	if (ep.GetAFC() == 0x02 || ep.GetAFC() == 0x03) && (ep[5]&0x10 == 0x10) {
		pcrBase := uint64(ep[6])<<25 | uint64(ep[7])<<17 | uint64(ep[8])<<9 | uint64(ep[9])<<1 | uint64(ep[10]>>7)
		pcrExtension := uint64(ep[10]&0x01)<<8 | uint64(ep[11])
		return pcrBase*300 + pcrExtension
	}
	return 0
}

// ClearPCR removes the PCR data from the packet if it exists.
func (ep *EncodedPacket) ClearPCR() {
	afc := ep.GetAFC()
	if afc == 0x02 || afc == 0x03 { // Check if the adaptation field is present
		afLength := int(ep[4])
		if afLength > 0 && (ep[5]&0x10) != 0 { // Check if PCR flag is set
			ep[5] &= 0xEF // Clear the PCR flag
			// Zero out PCR fields (6 bytes following the adaptation field length)
			for i := 6; i <= 11; i++ {
				ep[5+i] = 0
			}
		}
	}
}
