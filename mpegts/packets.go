package mpegts

import (
	"encoding/binary"
	"errors"
)

// Error constants
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

// Adaptations represents various adaptation field parameters in an MPEG-TS packet.
type Adaptations struct {
	PCR                      uint64 // Program Clock Reference (PCR)
	DTS                      uint64 // Decode Time Stamp (DTS)
	DiscontinuityIndicator   uint8  // Discontinuity Indicator
	TransportPrivateData     []byte // Transport Private Data
	AdaptationFieldExtension []byte // Adaptation Field Extension
	UnusedData               []byte // Carried forward unused data
}

// DecodedPacket represents the decoded headers of an MPEG-TS packet.
type DecodedPacket struct {
	SyncByte          byte        // Sync byte indicating the start of the packet
	TransportError    bool        // Transport Error Indicator (TEI)
	PayloadUnitStart  bool        // Payload Unit Start Indicator (PUSI)
	TransportPriority bool        // Transport Priority
	PID               uint16      // Packet Identifier (PID)
	ScramblingControl uint8       // Transport Scrambling Control
	Adaptation        Adaptations // Adaptation field data
	ContinuityCounter uint8       // Continuity Counter
}

// Packet represents an MPEG-TS packet along with its headers, adaptations, payload, and error status.
type Packet struct {
	Encoded     EncodedPacket  // Raw packet
	Decoded     *DecodedPacket // Headers
	Adaptations *Adaptations   // Adaptation Controls
	Payload     []byte         // TS payload
	Error       error          // Error, or nil
}

// NewPacket creates a new Packet instance from the given encoded packet.
func NewPacket(encodedPacket EncodedPacket) (*Packet, error) {
	if encodedPacket[0] != 0x47 {
		return nil, ErrInvalidSyncByte
	}

	if len(encodedPacket) != 188 {
		return nil, ErrInvalidPacketSize
	}

	adaptationFieldLength := int((encodedPacket[4] & 0b00000011) + 1)
	payloadStart := 4 + adaptationFieldLength

	packet := &Packet{
		Encoded: encodedPacket,
		Decoded: &DecodedPacket{
			SyncByte:          encodedPacket[0],
			TransportError:    encodedPacket[1]&0b10000000 != 0,
			PayloadUnitStart:  encodedPacket[1]&0b01000000 != 0,
			TransportPriority: encodedPacket[1]&0b00100000 != 0,
			PID:               binary.BigEndian.Uint16(encodedPacket[1:]) & 0x1FFF,
			ScramblingControl: (encodedPacket[3] >> 6) & 0b00000011,
			ContinuityCounter: encodedPacket[3] & 0b00001111,
		},
	}

	// Extract adaptation field
	if adaptationFieldLength > 0 {
		adaptations, err := NewAdaptations(encodedPacket[4:payloadStart])
		if err != nil {
			return nil, err
		}
		packet.Adaptations = adaptations
	}

	// Extract payload
	packet.Payload = encodedPacket[payloadStart:]

	return packet, nil
}

// Encode encodes the Packet into a complete MPEG-TS packet.
func (p *Packet) Encode() {
	encoded := make([]byte, 188)

	// Copy the header bytes
	copy(encoded[:4], p.Decoded.Encode())

	// Encode adaptations if present
	var adaptationBytes []byte
	if p.Adaptations != nil {
		adaptationBytes = p.Adaptations.Encode()
		copy(encoded[4:4+len(adaptationBytes)], adaptationBytes)
	}

	// Copy the payload
	copy(encoded[4+len(adaptationBytes):], p.Payload)

	// Copy encoded to p.Encoded
	copy(p.Encoded[:], encoded)

	return
}

// Encode encodes the DecodedPacket into a byte slice.
func (d *DecodedPacket) Encode() []byte {
	encoded := make([]byte, 4)
	encoded[0] = d.SyncByte
	if d.TransportError {
		encoded[1] |= 0b10000000
	}
	if d.PayloadUnitStart {
		encoded[1] |= 0b01000000
	}
	if d.TransportPriority {
		encoded[1] |= 0b00100000
	}
	binary.BigEndian.PutUint16(encoded[1:], d.PID)
	encoded[3] = d.ScramblingControl<<6 | d.ContinuityCounter
	return encoded
}

// NewAdaptations decodes the adaptation field from a byte slice
func NewAdaptations(data []byte) (*Adaptations, error) {
	if len(data) < 1 {
		println("missing data")
		return nil, ErrInvalidAdaptationsField
	}

	adaptations := &Adaptations{
		PCR:                      0, // Placeholder, update with actual PCR value if present
		DTS:                      0, // Placeholder, update with actual DTS value if present
		DiscontinuityIndicator:   data[0] & 0b10000000 >> 7,
		TransportPrivateData:     nil, // Placeholder, update with actual data if present
		AdaptationFieldExtension: nil, // Placeholder, update with actual data if present
		UnusedData:               nil, // Placeholder, update with remaining data
	}

	// Process the adaptation field data if it's present
	if len(data) > 1 {
		// Process PCR and DTS values if present
		if (data[0] & 0b00010000) != 0 {
			if len(data) < 10 {
				println("data too short for payload only")
				return nil, ErrInvalidAdaptationsField
			}
			adaptations.PCR = binary.BigEndian.Uint64(data[1:9])
			data = data[9:]
		}
		if (data[0] & 0b00100000) != 0 {
			if len(data) < 5 {
				println("data too short for adaptations only")
				return nil, ErrInvalidAdaptationsField
			}
			adaptations.DTS = binary.BigEndian.Uint64(data[1:5])
			data = data[5:]
		}

		// Handle Transport Private Data if present
		if (data[0] & 0b01000000) != 0 {
			// Extract Transport Private Data
			// The first byte indicates the length of the Transport Private Data
			// Adjust as per the specific format of Transport Private Data
			length := int(data[1])
			if len(data) < 2+length {
				println("data too short for transport private data")
				return nil, ErrInvalidAdaptationsField
			}
			adaptations.TransportPrivateData = data[2 : 2+length]
			data = data[2+length:]
		}

		// Handle Adaptation Field Extension if present
		if (data[0] & 0b00000001) != 0 {
			// Extract Adaptation Field Extension
			// Implement the decoding logic for Adaptation Field Extension
			// Adjust as per the specific format of Adaptation Field Extension
			// Placeholder for demonstration
			adaptations.AdaptationFieldExtension = data[1:]
		}
	}

	// Any remaining data is considered unused and carried forward
	adaptations.UnusedData = data

	return adaptations, nil
}

// Encode encodes the Adaptations into a byte slice.
func (a *Adaptations) Encode() []byte {
	var encoded []byte

	// Add discontinuity indicator
	encoded = append(encoded, (a.DiscontinuityIndicator&0b00000001)<<7)

	// Add PCR if present
	if a.PCR != 0 {
		pcrBytes := make([]byte, 6)
		binary.BigEndian.PutUint64(pcrBytes, a.PCR)
		encoded = append(encoded, 0b00010000)
		encoded = append(encoded, pcrBytes[1:]...)
	}

	// Add DTS if present
	if a.DTS != 0 {
		dtsBytes := make([]byte, 5)
		binary.BigEndian.PutUint64(dtsBytes, a.DTS)
		encoded = append(encoded, 0b00100000)
		encoded = append(encoded, dtsBytes[1:]...)
	}

	// Add Transport Private Data if present
	if len(a.TransportPrivateData) > 0 {
		encoded = append(encoded, 0b01000000)
		encoded = append(encoded, byte(len(a.TransportPrivateData)))
		encoded = append(encoded, a.TransportPrivateData...)
	}

	// Add Adaptation Field Extension if present
	if len(a.AdaptationFieldExtension) > 0 {
		encoded = append(encoded, 0b00000001)
		encoded = append(encoded, a.AdaptationFieldExtension...)
	}

	// Add remaining unused data
	encoded = append(encoded, a.UnusedData...)
	return encoded
}

// calculateAdaptationFieldLength calculates the length of the adaptation field based on whether it is present.
func calculateAdaptationFieldLength(packet EncodedPacket) int {
	adaptationFieldControl := packet[3] >> 4 & 0b00000011
	if adaptationFieldControl == 0b00000010 || adaptationFieldControl == 0b00000011 {
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
	packet[adaptationFieldStart] |= 0b00010000 // Set PCR flag

	// Set PCR fields in adaptation field
	packet[adaptationFieldStart+1] = byte(pcrBase >> 25)
	packet[adaptationFieldStart+2] = byte(pcrBase >> 17)
	packet[adaptationFieldStart+3] = byte(pcrBase >> 9)
	packet[adaptationFieldStart+4] = byte(pcrBase >> 1)
	packet[adaptationFieldStart+5] = byte(pcrBase<<7 | (pcrExt>>8)&0b01111111)
	packet[adaptationFieldStart+6] = byte(pcrExt & 0b11111111)
}
