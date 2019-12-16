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

type envelopMsg struct {
	Sender *Peer
	Packet *Packet `json:"envelope"`
	err    error
}

func newEnvelopMsgWithError(sender *Peer, err error) envelopMsg {
	return envelopMsg{
		Sender: sender,
		err:    err,
	}
}

func newEnvelopMsg(sender *Peer, packet *Packet) envelopMsg {
	return envelopMsg{
		Sender: sender,
		Packet: packet,
		err:    nil,
	}
}
