package protocol

import (
	"bytes"
	"encoding/binary"
	"io"
	"reflect"
	"testing"
	"testing/iotest"

	"dis-alg/pkg/core"
)

func TestReadWritePacket_HappyPath(t *testing.T) {
	original := &core.Packet{
		SourceID:     12345,
		PacketNumber: 9876543210,
		Payload:      []byte("mock-dis-pdu-data"),
	}

	var buf bytes.Buffer

	err := WritePacket(&buf, original)
	if err != nil {
		t.Fatalf("WritePacket failed: %v", err)
	}

	decoded, err := ReadPacket(&buf)
	if err != nil {
		t.Fatalf("ReadPacket failed: %v", err)
	}

	if !reflect.DeepEqual(original, decoded) {
		t.Errorf("Decoded packet %+v does not match original %+v", decoded, original)
	}
}

func TestReadPacket_PartialReads(t *testing.T) {
	original := &core.Packet{
		SourceID:     42,
		PacketNumber: 1,
		Payload:      []byte("fragmented-payload"),
	}

	var buf bytes.Buffer
	WritePacket(&buf, original)

	// Simulate fragmented TCP stream using OneByteReader
	fragmentedReader := iotest.OneByteReader(&buf)

	decoded, err := ReadPacket(fragmentedReader)
	if err != nil {
		t.Fatalf("ReadPacket failed on fragmented stream: %v", err)
	}

	if !reflect.DeepEqual(original, decoded) {
		t.Errorf("Decoded packet does not match original")
	}
}

func TestReadPacket_PayloadTooLarge(t *testing.T) {
	var buf bytes.Buffer
	
	// Write header manually with an excessively large payload size
	header := make([]byte, 16)
	binary.BigEndian.PutUint32(header[0:4], 1) // SourceID
	binary.BigEndian.PutUint64(header[4:12], 1) // PacketNumber
	binary.BigEndian.PutUint32(header[12:16], MaxPayloadSize+1) // PayloadLength (too large)
	
	buf.Write(header)
	
	_, err := ReadPacket(&buf)
	if err == nil {
		t.Fatal("Expected error for oversized payload, got nil")
	}
}

func TestReadPacket_IncompleteHeader(t *testing.T) {
	var buf bytes.Buffer
	buf.Write([]byte{0x00, 0x01}) // Only 2 bytes, header requires 16
	
	_, err := ReadPacket(&buf)
	if err != io.ErrUnexpectedEOF && err != io.EOF {
		t.Fatalf("Expected EOF/UnexpectedEOF, got %v", err)
	}
}
