package p2p

import (
	"github.com/eoscanada/eos-go"
)

// Envelope a packet from peer
type Envelope struct {
	Sender *Peer
	Packet *eos.Packet `json:"envelope"`
}

// newEnvelope create a envelope
func newEnvelope(sender *Peer, packet *eos.Packet) *Envelope {
	return &Envelope{
		Sender: sender,
		Packet: packet,
	}
}
