package mpegts

import (
	"encoding/binary"
	"errors"
)

type EncodedPacket [188]byte

type Adaptations struct {
	PCR                      uint64 // Program Clock Reference (PCR)
	DTS                      uint64 // Decode Time Stamp (DTS)
	DiscontinuityIndicator   uint8  // Discontinuity Indicator
	TransportPrivateData     []byte // Transport Private Data
	AdaptationFieldExtension []byte // Adaptation Field Extension
	UnusedData               []byte // Carried forward unused data
}

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

type Packet struct {
	Encoded     EncodedPacket  // Raw packet
	Decoded     *DecodedPacket // Headers
	Adaptations *Adaptations   // Adaptation Controls
	Payload     []byte         // TS payload
	Error       error          // Error, or nil
}

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

func NewPacket(encodedPacket EncodedPacket) (*Packet, error) {
	if encodedPacket[0] != 0x47 {
		return nil, ErrInvalidSyncByte
	}

	if len(encodedPacket) != 188 {
		return nil, ErrInvalidPacketSize
	}

	adaptationFieldLength := int((encodedPacket[4] & 0x03) + 1)
	payloadStart := 4 + adaptationFieldLength

	packet := &Packet{
		Encoded: encodedPacket,
		Decoded: &DecodedPacket{
			SyncByte:          encodedPacket[0],
			TransportError:    encodedPacket[1]&0x80 != 0,
			PayloadUnitStart:  encodedPacket[1]&0x40 != 0,
			TransportPriority: encodedPacket[1]&0x20 != 0,
			PID:               binary.BigEndian.Uint16(encodedPacket[1:]) & 0x1FFF,
			ScramblingControl: (encodedPacket[3] >> 6) & 0x03,
			ContinuityCounter: encodedPacket[3] & 0x0F,
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

// DecodeMultiplexingAdaptationField decodes the adaptation field from a byte slice
func NewAdaptations(data []byte) (*Adaptations, error) {
	if len(data) < 1 {
		return nil, ErrInvalidAdaptationsField
	}

	adaptations := &Adaptations{
		PCR:                      0, // Placeholder, update with actual PCR value if present
		DTS:                      0, // Placeholder, update with actual DTS value if present
		DiscontinuityIndicator:   data[0] & 0x80 >> 7,
		TransportPrivateData:     nil, // Placeholder, update with actual data if present
		AdaptationFieldExtension: nil, // Placeholder, update with actual data if present
		UnusedData:               nil, // Placeholder, update with remaining data
	}

	// Process the adaptation field data if it's present
	if len(data) > 1 {
		// Process PCR and DTS values if present
		if (data[0] & 0x10) != 0 {
			if len(data) < 10 {
				return nil, ErrInvalidAdaptationsField
			}
			adaptations.PCR = binary.BigEndian.Uint64(data[1:9])
			data = data[9:]
		}
		if (data[0] & 0x20) != 0 {
			if len(data) < 5 {
				return nil, ErrInvalidAdaptationsField
			}
			adaptations.DTS = binary.BigEndian.Uint64(data[1:5])
			data = data[5:]
		}

		// Handle Transport Private Data if present
		if (data[0] & 0x40) != 0 {
			// Extract Transport Private Data
			// The first byte indicates the length of the Transport Private Data
			// Adjust as per the specific format of Transport Private Data
			length := int(data[1])
			if len(data) < 2+length {
				return nil, ErrInvalidAdaptationsField
			}
			adaptations.TransportPrivateData = data[2 : 2+length]
			data = data[2+length:]
		}

		// Handle Adaptation Field Extension if present
		if (data[0] & 0x01) != 0 {
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

// EncodeMultiplexingAdaptationField encodes the adaptation field into a byte slice
func (a *Adaptations) Encode() ([]byte, error) {
	// Placeholder implementation for encoding adaptations
	// Implement the logic to encode adaptations into a byte slice
	// Adjust as per the specific requirements and format of adaptations

	// This is just a placeholder returning an empty byte slice
	return []byte{}, nil
}

// calculateAdaptationFieldLength calculates the length of the adaptation field based on whether it is present.
func calculateAdaptationFieldLength(packet EncodedPacket) int {
	adaptationFieldControl := packet[3] >> 4 & 0x03
	if adaptationFieldControl == 0x02 || adaptationFieldControl == 0x03 {
		// Adaptation field is present
		println("Adaptation field is present", packet[4:4+int(packet[4])+1], int(packet[4])+1)
		return int(packet[4]) + 1
	}
	println("Adaptation field is not present", packet[3]>>4&0x03, packet[4])
	return 0
}

// SetPCR sets the PCR value in the adaptation field of the MPEG-TS packet header.
// PCR frequency is the frequency of the Program Clock Reference in Hz.
func SetPCR(packet *EncodedPacket, pcr uint64, pcrFrequency uint64) {
	println("PCR called", packet[3]>>4&0x03, packet[4], pcr, pcrFrequency, calculateAdaptationFieldLength(*packet))
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

	println("PCR complete", packet[3]>>4&0x03, packet[4], calculateAdaptationFieldLength(*packet))
}
