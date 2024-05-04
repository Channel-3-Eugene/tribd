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

// EncodedPackets represents a collection of raw MPEG-TS packets.
type EncodedPackets []*EncodedPacket

func (ep *EncodedPacket) IsMPEGTS() bool {
	return ep[0] == 0x47
}

func (ep *EncodedPacket) GetSyncByte() byte {
	return ep[0]
}

func (ep *EncodedPacket) GetTEI() bool {
	return ep[1]&0x80 != 0
}

func (ep *EncodedPacket) SetTEI() {
	ep[1] |= 0x80
}

func (ep *EncodedPacket) ClearTEI() {
	ep[1] &= 0x7F
}

func (ep *EncodedPacket) GetPUSI() bool {
	return ep[1]&0x40 != 0
}

func (ep *EncodedPacket) SetPUSI() {
	ep[1] |= 0x40
}

func (ep *EncodedPacket) ClearPUSI() {
	ep[1] &= 0xBF
}

func (ep *EncodedPacket) GetPID() uint16 {
	return binary.BigEndian.Uint16(ep[1:3]) & 0x1FFF
}

func (ep *EncodedPacket) SetPID(pid uint16) {
	binary.BigEndian.PutUint16(ep[1:3], pid)
}

func (ep *EncodedPacket) GetTSC() uint8 {
	return (ep[3] >> 6) & 0x03
}

func (ep *EncodedPacket) SetTSC(tsc uint8) {
	ep[3] = (ep[3] & 0x3F) | (tsc << 6)
}

func (ep *EncodedPacket) GetAFC() uint8 {
	return (ep[3] >> 4) & 0x03
}

func (ep *EncodedPacket) SetAFC(afc uint8) {
	ep[3] = (ep[3] & 0xCF) | (afc << 4)
}

func (ep *EncodedPacket) GetCC() uint8 {
	return ep[3] & 0x0F
}

func (ep *EncodedPacket) SetCC(cc uint8) {
	ep[3] = (ep[3] & 0xF0) | (cc & 0x0F)
}

func (ep *EncodedPacket) GetPayload() []byte {
	if ep.GetAFC() == 0x01 {
		return ep[4 : 4+ep[4]]
	}
	return ep[4:]
}

func (ep *EncodedPacket) SetPayload(payload []byte) {
	if ep.GetAFC() == 0x01 {
		ep[4] = byte(len(payload))
		copy(ep[5:], payload)
	} else {
		copy(ep[4:], payload)
	}
}

func (ep *EncodedPacket) GetAdaptationField() []byte {
	if ep.GetAFC() == 0x02 || ep.GetAFC() == 0x03 {
		return ep[4 : 4+ep[4]]
	}
	return nil
}

func (ep *EncodedPacket) SetAdaptationField(af []byte) {
	if ep.GetAFC() == 0x02 || ep.GetAFC() == 0x03 {
		ep[4] = byte(len(af))
		copy(ep[5:], af)
	}
}

func (ep *EncodedPacket) GetPCR() uint64 {
	if ep.GetAFC() == 0x02 || ep.GetAFC() == 0x03 {
		return binary.BigEndian.Uint64(ep[5:13]) & 0x1FFFFFFFFFFFF
	}
	return 0
}

func (ep *EncodedPacket) SetPCR(pcr uint64) {
	if ep.GetAFC() == 0x02 || ep.GetAFC() == 0x03 {
		binary.BigEndian.PutUint64(ep[5:13], pcr)
	}
}

func (ep *EncodedPacket) GetOPCR() uint64 {
	if ep.GetAFC() == 0x03 {
		return binary.BigEndian.Uint64(ep[13:21]) & 0x1FFFFFFFFFFFF
	}
	return 0
}

func (ep *EncodedPacket) SetOPCR(opcr uint64) {
	if ep.GetAFC() == 0x03 {
		binary.BigEndian.PutUint64(ep[13:21], opcr)
	}
}

func (ep *EncodedPacket) GetSpliceCountdown() uint8 {
	if ep.GetAFC() == 0x03 {
		return ep[21]
	}
	return 0
}

func (ep *EncodedPacket) SetSpliceCountdown(sc uint8) {
	if ep.GetAFC() == 0x03 {
		ep[21] = sc
	}
}

func (ep *EncodedPacket) GetTransportPrivateData() []byte {
	if ep.GetAFC() == 0x03 {
		return ep[22 : 22+ep[22]]
	}
	return nil
}

func (ep *EncodedPacket) SetTransportPrivateData(tpd []byte) {
	if ep.GetAFC() == 0x03 {
		ep[22] = byte(len(tpd))
		copy(ep[23:], tpd)
	}
}

func (ep *EncodedPacket) GetAdaptationFieldExtension() []byte {
	if ep.GetAFC() == 0x03 {
		return ep[22+ep[22] : 22+ep[22]+ep[22+ep[22]]]
	}
	return nil
}

func (ep *EncodedPacket) SetAdaptationFieldExtension(afe []byte) {
	if ep.GetAFC() == 0x03 {
		ep[22+ep[22]] = byte(len(afe))
		copy(ep[23+ep[22]:], afe)
	}
}
