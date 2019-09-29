package p2p

import (
	"math"

	"github.com/eoscanada/eos-go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type syncManager struct {
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
	s.requestedEndBlock = s.headBlock + uint32(math.Min(float64(delta), 100))

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
func (s *syncManager) OnHandshakeMsg(envelope *Envelope, msg *eos.HandshakeMessage) {
}

// OnGoAwayMsg handler func imp
func (s *syncManager) OnGoAwayMsg(envelope *Envelope, msg *eos.GoAwayMessage) {
}

// OnTimeMsg handler func imp
func (s *syncManager) OnTimeMsg(envelope *Envelope, msg *eos.TimeMessage) {
}

// OnNoticeMsg handler func imp
func (s *syncManager) OnNoticeMsg(envelope *Envelope, msg *eos.NoticeMessage) {
}

// OnRequestMsg handler func imp
func (s *syncManager) OnRequestMsg(envelope *Envelope, msg *eos.RequestMessage) {

}

// OnSyncRequestMsg handler func imp
func (s *syncManager) OnSyncRequestMsg(envelope *Envelope, msg *eos.SyncRequestMessage) {

}

// OnSignedBlock handler func imp
func (s *syncManager) OnSignedBlock(envelope *Envelope, msg *eos.SignedBlock) {

}

// OnPackedTransactionMsg handler func imp
func (s *syncManager) OnPackedTransactionMsg(envelope *Envelope, msg *eos.PackedTransactionMessage) {

}
