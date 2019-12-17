package p2p

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// peerMsg for peer new/delete/error msg
type peerMsg struct {
	msgTyp peerMsgTyp
	peer   *Peer
	cfg    *PeerCfg
	err    error
}

type peerMsgTyp uint8

const (
	peerMsgNewPeer = peerMsgTyp(iota)
	peerMsgDelPeer
	peerMsgErrPeer
)

func (c *Client) peerMngLoop(ctx context.Context) {
	for {
		select {
		case p := <-c.peerChan:
			// if already stop loop so close directly
			select {
			case <-ctx.Done():
				p2pLog.Info("close peer chan mng")
				return
			default:
			}

			switch p.msgTyp {
			case peerMsgNewPeer:
				c.onNewPeer(ctx, &p)
			case peerMsgDelPeer:
				c.onDelPeer(ctx, &p)
			case peerMsgErrPeer:
				c.onErrPeer(ctx, &p)
			}

		case <-ctx.Done():
			// no need wait all msg in chan processed
			p2pLog.Info("close peer chan mng")
			return
		}
	}
}

func (c *Client) onNewPeer(ctx context.Context, msg *peerMsg) {
	p2pLog.Info("new peer", zap.String("addr", msg.cfg.Address))
	_, ok := c.ps[msg.cfg.Address]
	if ok {
		p2pLog.Info("connect had created, no new another", zap.String("addr", msg.cfg.Address))
		return
	}

	peer, err := NewPeer(msg.cfg, c.HeadBlockNum(), c.ChainID())

	if err != nil {
		p2pLog.Error("new peer failed", zap.String("addr", msg.peer.Address), zap.Error(err))
	}

	c.ps[msg.cfg.Address] = peerStatus{
		peer:   peer,
		status: peerStatInited,
		cfg:    msg.cfg,
	}

	c.StartPeer(ctx, peer)
}

func (c *Client) onDelPeer(ctx context.Context, msg *peerMsg) {
	p2pLog.Info("del peer", zap.String("addr", msg.peer.Address))
}

func (c *Client) onErrPeer(ctx context.Context, msg *peerMsg) {
	if msg.err != nil && msg.peer != nil {
		p2pLog.Info("reconnect peer", zap.String("addr", msg.peer.Address))
		if err := c.StartPeer(ctx, msg.peer); err != nil {
			time.Sleep(3 * time.Second)
		}
	}
}

// StartPeer start a peer r/w
func (c *Client) StartPeer(ctx context.Context, p *Peer) error {
	p2pLog.Info("Start Connect Peer", zap.String("peer", p.Address))
	err := p.Start(ctx, c)
	if err != nil {
		c.peerChan <- peerMsg{
			err:    errors.Wrap(err, "connect error"),
			peer:   p,
			msgTyp: peerMsgErrPeer,
		}
		return err
	}

	return nil
}

// NewPeer new peer to connect
func (c *Client) NewPeer(cfg *PeerCfg) error {
	c.peerChan <- peerMsg{
		msgTyp: peerMsgNewPeer,
		cfg:    cfg,
	}
	return nil
}
