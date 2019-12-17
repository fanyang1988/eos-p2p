package p2p

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
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

			if p.peer != nil {
				switch p.msgTyp {
				case peerMsgNewPeer:
				case peerMsgDelPeer:
				case peerMsgErrPeer:
					if p.err != nil {
						p2pLog.Info("reconnect peer", zap.String("addr", p.peer.Address))
						if err := c.StartPeer(ctx, p.peer); err != nil {
							time.Sleep(3 * time.Second)
						}
					}
				}
			}

		case <-ctx.Done():
			// no need wait all msg in chan processed
			p2pLog.Info("close peer chan mng")
			return
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
func (c *Client) NewPeer(p *PeerCfg) error {
	return nil
}
