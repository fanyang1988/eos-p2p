package p2p

import (
	"context"

	"go.uber.org/zap"
)

// sync: need sync 3 class datas:
// - 1. irreversible blocks : which can be see as a const data, will from a selected peer
// - 2. blocks : blocks not irreversible current
// - 3. transactions : trx not into block
// first, sync irreversible blocks, then start sync blocks or trxs

// onStartSyncIrreversible start to sync all irreversible block by the peer
func (c *Client) onStartSyncIrreversible(peer *Peer) {
	c.logger.Info("start sync all blocks", zap.String("addr", peer.Address))
	stat := c.blkStorer.State()

	c.logger.Info("curr stat", zap.Uint32("headNum", stat.HeadBlockNum))

	h := stat.ToHandshakeInfo()
	peer.SendHandshake(h)
}

// startSyncIrr start sync by peer
func (c *Client) startSyncIrr(peer *Peer) {
	c.packetChan <- envelopMsg{
		Sender: peer,
		typ:    envelopMsgStartSync,
	}
}

// onSyncFinished (IN peerMngLoop) when sync irr success start to sync blocks and trxs( if need )
func (c *Client) onSyncFinished(ctx context.Context, msg *peerMsg) {
	c.logger.Info("sync finished", zap.Uint32("current head", c.HeadBlockNum()))

	stat := c.blkStorer.State()
	h := stat.ToHandshakeInfo()

	for _, peerStat := range c.ps {
		if peerStat.status != peerStatNormal {
			continue
		}
		peerStat.peer.SendHandshake(h)
	}
}

// syncSuccessNotice notice sync irr stop
func (c *Client) syncSuccessNotice(peer *Peer) {
	c.peerChan <- peerMsg{
		msgTyp: peerSyncFinished,
		peer:   peer,
	}
}
