package core

// Packet represents a fully framed DIS-ALG message.
type Packet struct {
	SourceID     uint32
	PacketNumber uint64
	Payload      []byte // The raw DIS PDU
}
