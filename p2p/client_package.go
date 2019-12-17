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
	envelopMsgDelHandler
	envelopMsgError
	envelopMsgPacket
)

type envelopMsg struct {
	Sender  *Peer
	Packet  *Packet
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

func newHandlerDelMsg(h Handler) envelopMsg {
	return envelopMsg{
		handler: h,
		typ:     envelopMsgDelHandler,
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
				hname := r.handler.Name()
				p2pLog.Info("new handler", zap.String("name", hname))
				for idx, h := range c.handlers {
					if h.Name() == hname {
						p2pLog.Info("replace handler", zap.String("name", hname))
						c.handlers[idx] = r.handler
						continue
					}
				}
				c.handlers = append(c.handlers, r.handler)

			case envelopMsgDelHandler:
				hname := r.handler.Name()
				p2pLog.Info("del handler", zap.String("name", hname))
				for idx, h := range c.handlers {
					if h.Name() == hname {
						// not change seq with handlers
						for i := idx; i < len(c.handlers)-1; i++ {
							c.handlers[i] = c.handlers[i+1]
						}
						c.handlers = c.handlers[:len(c.handlers)-1]
						continue
					}
				}
				p2pLog.Warn("no found hander to del", zap.String("name", hname))

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
	c.packetChan <- newHandlerAddMsg(handler)
}
