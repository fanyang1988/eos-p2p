package p2p

import (
	"math"

	"github.com/eoscanada/eos-go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	// BlockNumPerRequest the number of block in a request when sync
	BlockNumPerRequest uint32 = 50
)

type syncManager struct {
	IsSyncAll           bool
	IsCatchingUp        bool
	requestedStartBlock uint32
	requestedEndBlock   uint32
	headBlock           uint32
	originHeadBlock     uint32
}

func (s *syncManager) sendSyncRequest(peer *Peer) error {
	s.IsCatchingUp = true

	delta := s.originHeadBlock - s.headBlock

	s.requestedStartBlock = s.headBlock
	s.requestedEndBlock = s.headBlock + uint32(math.Min(float64(delta), float64(BlockNumPerRequest)))

	p2pLog.Debug("Sending sync request",
		zap.Uint32("startBlock", s.requestedStartBlock),
		zap.Uint32("endBlock", s.requestedEndBlock))

	err := peer.SendSyncRequest(s.requestedStartBlock, s.requestedEndBlock+1)
	if err != nil {
		return errors.Wrapf(err, "send sync request to %s", peer.Address)
	}

	return nil
}

// OnHandshakeMsg handler func imp
func (s *syncManager) OnHandshakeMsg(peer *Peer, msg *eos.HandshakeMessage) {
	if s.IsSyncAll {
		s.originHeadBlock = msg.HeadNum
		err := s.sendSyncRequest(peer)
		if err != nil {
			//errChannel <- errors.Wrap(err, "handshake: sending sync request")
			peer.ClosePeer()
		}
		s.IsCatchingUp = true
	} else {
		msg.NodeID = peer.NodeID
		msg.P2PAddress = peer.Name
		err := peer.WriteP2PMessage(msg)
		if err != nil {
			peer.Close(eos.GoAwayNoReason)
			return
		}
		p2pLog.Debug("Handshake resent", zap.String("other", msg.P2PAddress))

	}
}

// OnGoAwayMsg handler func imp
func (s *syncManager) OnGoAwayMsg(peer *Peer, msg *eos.GoAwayMessage) {
	p2pLog.Warn("peer goaway", zap.String("reason", msg.Reason.String()))
	peer.ClosePeer()
}

// OnTimeMsg handler func imp
func (s *syncManager) OnTimeMsg(peer *Peer, msg *eos.TimeMessage) {
}

// OnNoticeMsg handler func imp
func (s *syncManager) OnNoticeMsg(peer *Peer, msg *eos.NoticeMessage) {
	if s.IsSyncAll {
		pendingNum := msg.KnownBlocks.Pending
		if pendingNum > 0 {
			s.originHeadBlock = pendingNum
			err := s.sendSyncRequest(peer)
			if err != nil {
				//errChannel <- errors.Wrap(err, "noticeMessage: sending sync request")
				peer.ClosePeer()
			}
		}
	}
}

// OnRequestMsg handler func imp
func (s *syncManager) OnRequestMsg(peer *Peer, msg *eos.RequestMessage) {

}

// OnSyncRequestMsg handler func imp
func (s *syncManager) OnSyncRequestMsg(peer *Peer, msg *eos.SyncRequestMessage) {

}

// OnSignedBlock handler func imp
func (s *syncManager) OnSignedBlock(peer *Peer, msg *eos.SignedBlock) {
	if s.IsSyncAll {
		blockNum := msg.BlockNumber()
		s.headBlock = blockNum
		if s.requestedEndBlock == blockNum {
			if s.originHeadBlock <= blockNum {
				p2pLog.Debug("In sync with last handshake")
				blockID, err := msg.BlockID()
				if err != nil {
					//errChannel <- errors.Wrap(err, "getting block id")
					peer.Close(eos.GoAwayValidation)
					return
				}
				peer.handshakeInfo.HeadBlockNum = blockNum
				peer.handshakeInfo.HeadBlockID = blockID
				peer.handshakeInfo.HeadBlockTime = msg.SignedBlockHeader.Timestamp.Time
				err = peer.SendHandshake(peer.handshakeInfo)
				if err != nil {
					//errChannel <- errors.Wrap(err, "send handshake")
					peer.ClosePeer()
				}
				p2pLog.Debug("Send new handshake", zap.Object("handshakeInfo", peer.handshakeInfo))
			} else {
				err := s.sendSyncRequest(peer)
				if err != nil {
					//errChannel <- errors.Wrap(err, "signed block: sending sync request")
					peer.ClosePeer()
				}
			}
		}
	}
}

// OnPackedTransactionMsg handler func imp
func (s *syncManager) OnPackedTransactionMsg(peer *Peer, msg *eos.PackedTransactionMessage) {

}
