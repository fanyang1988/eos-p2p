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

// peerMsg for peer new/delete/error msg
type peerMsg struct {
	msgTyp peerMsgTyp
	peer   *Peer
	err    error
}

type peerMsgTyp uint8

const (
	peerMsgNewPeer = peerMsgTyp(iota)
	peerMsgDelPeer
	peerMsgErrPeer
)
