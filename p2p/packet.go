package p2p

// Envelope a packet from peer
type Envelope struct {
	Sender *Peer
	Packet *Packet `json:"envelope"`
}

// newEnvelope create a envelope
func newEnvelope(sender *Peer, packet *Packet) *Envelope {
	return &Envelope{
		Sender: sender,
		Packet: packet,
	}
}
