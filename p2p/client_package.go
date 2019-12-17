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
				c.onAddHandlerMsg(&r)
			case envelopMsgDelHandler:
				c.onDelHandlerMsg(&r)
			case envelopMsgError:
				if !isStopped {
					c.onPeerErrorMsg(&r)
				}
			case envelopMsgPacket:
				c.onPacketMsg(&r)
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

func (c *Client) onAddHandlerMsg(r *envelopMsg) {
	handlerName := r.handler.Name()
	p2pLog.Info("new handler", zap.String("name", handlerName))
	for idx, h := range c.handlers {
		if h.Name() == handlerName {
			p2pLog.Info("replace handler", zap.String("name", handlerName))
			c.handlers[idx] = r.handler
			return
		}
	}
	c.handlers = append(c.handlers, r.handler)
}

func (c *Client) onDelHandlerMsg(r *envelopMsg) {
	handlerName := r.handler.Name()
	p2pLog.Info("del handler", zap.String("name", handlerName))
	for idx, h := range c.handlers {
		if h.Name() == handlerName {
			// not change seq with handlers
			for i := idx; i < len(c.handlers)-1; i++ {
				c.handlers[i] = c.handlers[i+1]
			}
			c.handlers = c.handlers[:len(c.handlers)-1]
			return
		}
	}
	p2pLog.Warn("no found hander to del", zap.String("name", handlerName))
}

func (c *Client) onPacketMsg(r *envelopMsg) {
	envelope := newEnvelope(r.Sender, r.Packet)
	c.syncHandler.Handle(envelope)
	for _, handle := range c.handlers {
		handle.Handle(envelope)
	}
}

func (c *Client) onPeerErrorMsg(r *envelopMsg) {
	if r.err == nil {
		return
	}

	if errors.Cause(r.err) != io.EOF {
		p2pLog.Info("client res error", zap.Error(r.err))
	} else {
		p2pLog.Info("conn closed")
		c.peerChan <- peerMsg{
			err:    r.err,
			peer:   r.Sender,
			msgTyp: peerMsgErrPeer,
		}
	}

}

// RegisterHandler reg handler to client
func (c *Client) RegisterHandler(handler Handler) {
	c.packetChan <- newHandlerAddMsg(handler)
}
