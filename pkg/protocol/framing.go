package protocol

import (
	"encoding/binary"
	"errors"
	"io"

	"dis-alg/pkg/core"
)

const MaxPayloadSize = 65535

var ErrPayloadTooLarge = errors.New("payload exceeds maximum allowed size")

// WritePacket serializes a packet and writes it to the writer.
func WritePacket(w io.Writer, p *core.Packet) error {
	payloadLen := len(p.Payload)
	if payloadLen > MaxPayloadSize {
		return ErrPayloadTooLarge
	}

	buf := make([]byte, 16+payloadLen)
	binary.BigEndian.PutUint32(buf[0:4], p.SourceID)
	binary.BigEndian.PutUint64(buf[4:12], p.PacketNumber)
	binary.BigEndian.PutUint32(buf[12:16], uint32(payloadLen))
	copy(buf[16:], p.Payload)

	_, err := w.Write(buf)
	return err
}

// ReadPacket reads a fully framed DIS-ALG packet from the reader.
func ReadPacket(r io.Reader) (*core.Packet, error) {
	header := make([]byte, 16)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	sourceID := binary.BigEndian.Uint32(header[0:4])
	packetNumber := binary.BigEndian.Uint64(header[4:12])
	payloadLen := binary.BigEndian.Uint32(header[12:16])

	if payloadLen > MaxPayloadSize {
		return nil, ErrPayloadTooLarge
	}

	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}

	return &core.Packet{
		SourceID:     sourceID,
		PacketNumber: packetNumber,
		Payload:      payload,
	}, nil
}
