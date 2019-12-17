package p2p

import (
	"context"
	"io"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type envelopMsgTyp uint8

const (
	envelopMsgNil = envelopMsgTyp(iota)
	envelopMsgAddHandler
	envelopMsgError
	envelopMsgPacket
)

type envelopMsg struct {
	Sender  *Peer
	Packet  *Packet `json:"envelope"`
	handler Handler
	typ     envelopMsgTyp
	err     error
}

func newEnvelopMsgWithError(sender *Peer, err error) envelopMsg {
	return envelopMsg{
		Sender: sender,
		err:    err,
		typ:    envelopMsgError,
	}
}

func newEnvelopMsg(sender *Peer, packet *Packet) envelopMsg {
	return envelopMsg{
		Sender: sender,
		Packet: packet,
		typ:    envelopMsgPacket,
	}
}

func newHandlerAddMsg(h Handler) envelopMsg {
	return envelopMsg{
		handler: h,
		typ:     envelopMsgAddHandler,
	}
}

func (c *Client) peerLoop(ctx context.Context) {
	isStopped := false
	for {
		select {
		case r, ok := <-c.packetChan:
			if !ok {
				p2pLog.Info("client peerLoop stop")
				return
			}

			switch r.typ {
			case envelopMsgAddHandler:
				p2pLog.Info("new handler")
				c.handlers = append(c.handlers, r.handler)

			case envelopMsgError:
				if r.err != nil {
					if errors.Cause(r.err) != io.EOF {
						p2pLog.Info("client res error", zap.Error(r.err))
					} else {
						p2pLog.Info("conn closed")
						if !isStopped {
							c.peerChan <- peerMsg{
								err:    r.err,
								peer:   r.Sender,
								msgTyp: peerMsgErrPeer,
							}
						}
					}
					continue
				}
			case envelopMsgPacket:
				envelope := newEnvelope(r.Sender, r.Packet)
				for _, handle := range c.handlers {
					handle.Handle(envelope)
				}
			}

		case <-ctx.Done():
			if !isStopped {
				isStopped = true
				p2pLog.Info("close p2p client")
				c.closeAllPeer()
				p2pLog.Info("all peer is closed, to close client peerLoop")
				close(c.packetChan)
			}
		}
	}
}

// RegisterHandler reg handler to client
func (c *Client) RegisterHandler(handler Handler) {
	c.handlers = append(c.handlers, handler)
}
