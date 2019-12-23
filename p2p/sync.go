package p2p

import (
	"math"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	// BlockNumPerRequest the number of block in a request when sync
	BlockNumPerRequest uint32 = 50
)

type syncManager struct {
	syncHandler syncHandlerInterface
	cli         *Client
}

type syncHandlerInterface interface {
	OnRequestMsg(peer *Peer, msg *RequestMessage) error
	OnSyncRequestMsg(peer *Peer, msg *SyncRequestMessage) error
	OnHandshakeMsg(peer *Peer, msg *HandshakeMessage) error
	OnNoticeMsg(peer *Peer, msg *NoticeMessage) error
	OnSignedBlock(peer *Peer, msg *SignedBlock) error
}

func (s *syncManager) init(isSyncIrr bool) {
	if isSyncIrr {
		s.syncHandler = &syncIrreversibleHandler{
			cli: s.cli,
		}
	} else {
		s.syncHandler = &syncNoIrrHandler{
			cli: s.cli,
		}
	}

	s.cli.syncHandler = NewMsgHandler("sync", s)
}

// OnHandshakeMsg handler func imp
func (s *syncManager) OnHandshakeMsg(peer *Peer, msg *HandshakeMessage) {
	if err := s.syncHandler.OnHandshakeMsg(peer, msg); err != nil {
		p2pLog.Error("on handshake msg error", zap.Error(err))
	}
}

// OnGoAwayMsg handler func imp
func (s *syncManager) OnGoAwayMsg(peer *Peer, msg *GoAwayMessage) {
	p2pLog.Warn("peer goaway", zap.String("reason", msg.Reason.String()))
	peer.ClosePeer()
}

// OnTimeMsg handler func imp
func (s *syncManager) OnTimeMsg(peer *Peer, msg *TimeMessage) {
	peer.SendTime(msg)
}

// OnNoticeMsg handler func imp
func (s *syncManager) OnNoticeMsg(peer *Peer, msg *NoticeMessage) {
	if err := s.syncHandler.OnNoticeMsg(peer, msg); err != nil {
		p2pLog.Error("on notice msg error", zap.Error(err))
	}
}

// OnRequestMsg handler func imp
func (s *syncManager) OnRequestMsg(peer *Peer, msg *RequestMessage) {
	// TODO: can sync to others
}

// OnSyncRequestMsg handler func imp
func (s *syncManager) OnSyncRequestMsg(peer *Peer, msg *SyncRequestMessage) {
	// TODO: can sync to others
}

// OnSignedBlock handler func imp
func (s *syncManager) OnSignedBlock(peer *Peer, msg *SignedBlock) {
	if err := s.syncHandler.OnSignedBlock(peer, msg); err != nil {
		p2pLog.Error("on block msg error", zap.Error(err))
	}
}

// OnPackedTransactionMsg handler func imp
func (s *syncManager) OnPackedTransactionMsg(peer *Peer, msg *PackedTransactionMessage) {
	// do nothing
}

// syncIrreversibleHandler handler for syncManager when client is sync irreversible
type syncIrreversibleHandler struct {
	requestedStartBlock uint32
	requestedEndBlock   uint32
	originHeadBlock     uint32
	cli                 *Client
}

// No need imp
func (h *syncIrreversibleHandler) OnRequestMsg(peer *Peer, msg *RequestMessage) error { return nil }
func (h *syncIrreversibleHandler) OnSyncRequestMsg(peer *Peer, msg *SyncRequestMessage) error {
	return nil
}

func (h *syncIrreversibleHandler) sendSyncRequest(peer *Peer) error {
	// update sync status
	headBlockNum := h.cli.HeadBlockNum()
	delta := h.originHeadBlock - headBlockNum
	h.requestedStartBlock = headBlockNum
	h.requestedEndBlock = headBlockNum + uint32(math.Min(float64(delta), float64(BlockNumPerRequest)))

	p2pLog.Debug("Sending sync request",
		zap.Uint32("startBlock", h.requestedStartBlock),
		zap.Uint32("endBlock", h.requestedEndBlock))

	// send req
	err := peer.SendSyncRequest(h.requestedStartBlock, h.requestedEndBlock+1)
	if err != nil {
		return errors.Wrapf(err, "send sync request to %s", peer.Address)
	}

	return nil
}

// OnHandshakeMsg when need sync irreversible blocks, after handshake client need send req to peer
func (h *syncIrreversibleHandler) OnHandshakeMsg(peer *Peer, msg *HandshakeMessage) error {
	// init sync status
	h.originHeadBlock = msg.HeadNum
	return h.sendSyncRequest(peer)
}

// OnNoticeMsg
func (h *syncIrreversibleHandler) OnNoticeMsg(peer *Peer, msg *NoticeMessage) error {
	pendingNum := msg.KnownBlocks.Pending
	if pendingNum > 0 {
		h.originHeadBlock = pendingNum
		return h.sendSyncRequest(peer)
	}
	return nil
}

// OnSignedBlock handler func imp
func (h *syncIrreversibleHandler) OnSignedBlock(peer *Peer, msg *SignedBlock) error {
	blockNum := msg.BlockNumber()
	h.cli.SetHeadBlock(msg)

	// update sync status
	if h.requestedEndBlock != blockNum {
		// need to get more blocks, no need process new request in next
		return nil
	}

	if h.originHeadBlock <= blockNum {
		// now block have got all

		p2pLog.Debug("have sync all blocks needed",
			zap.Uint32("originHead", h.originHeadBlock),
			zap.Uint32("to", blockNum))

		h.cli.syncSuccessNotice(peer)

		return nil
	}

	// need get next blocks by sync
	return h.sendSyncRequest(peer)

}

// syncNoIrrHandler handler for syncManager when client is sync blocks and trxs
type syncNoIrrHandler struct {
	cli *Client
}

// No need imp
func (h *syncNoIrrHandler) OnRequestMsg(peer *Peer, msg *RequestMessage) error         { return nil }
func (h *syncNoIrrHandler) OnSyncRequestMsg(peer *Peer, msg *SyncRequestMessage) error { return nil }

// OnHandshakeMsg
func (h *syncNoIrrHandler) OnHandshakeMsg(peer *Peer, msg *HandshakeMessage) error {
	msg.NodeID = peer.NodeID
	msg.P2PAddress = peer.Name

	// TODO: use status from forkDB

	return peer.WriteP2PMessage(msg)
}

// OnNoticeMsg
func (h *syncNoIrrHandler) OnNoticeMsg(peer *Peer, msg *NoticeMessage) error {
	// TODO: can sync to other
	return nil
}

// OnSignedBlock handler func imp
func (h *syncNoIrrHandler) OnSignedBlock(peer *Peer, msg *SignedBlock) error {
	// TODO: to forkDB
	h.cli.SetHeadBlock(msg)
	return nil
}
