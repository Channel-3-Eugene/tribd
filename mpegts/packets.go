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

// GetPCR returns the Program Clock Reference (PCR) value from the adaptation field.
func (ep *EncodedPacket) GetPCR() uint64 {
	if ep.GetAFC() == 0x02 || ep.GetAFC() == 0x03 {
		return binary.BigEndian.Uint64(ep[5:13]) & 0x1FFFFFFFFFFFF
	}
	return 0
}

// SetPCR sets the Program Clock Reference (PCR) value in the adaptation field.
func (ep *EncodedPacket) SetPCR(pcr uint64) {
	if ep.GetAFC() == 0x02 || ep.GetAFC() == 0x03 {
		binary.BigEndian.PutUint64(ep[5:13], pcr)
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

// SetPCR sets the PCR value in the adaptation field of the MPEG-TS packet header.
// PCR frequency is the frequency of the Program Clock Reference in Hz.
func SetPCR(packet *EncodedPacket, pcr uint64, pcrFrequency uint64) {
	// PCR base is 33 bits, divided into two 33-bit fields in the MPEG-TS header
	// PCR extension is 9 bits
	// The PCR is divided by the PCR frequency to get the value in PCR units (in Hz)
	pcrBase := pcr / pcrFrequency
	pcrExt := pcr % pcrFrequency
	adaptationFieldStart := 4

	// Set adaptation field control to indicate PCR presence
	packet[adaptationFieldStart] |= 0x10 // Set PCR flag

	// Set PCR fields in adaptation field
	packet[adaptationFieldStart+1] = byte(pcrBase >> 25)
	packet[adaptationFieldStart+2] = byte(pcrBase >> 17)
	packet[adaptationFieldStart+3] = byte(pcrBase >> 9)
	packet[adaptationFieldStart+4] = byte(pcrBase >> 1)
	packet[adaptationFieldStart+5] = byte(pcrBase<<7 | (pcrExt>>8)&0x7F)
	packet[adaptationFieldStart+6] = byte(pcrExt & 0xFF)
}
