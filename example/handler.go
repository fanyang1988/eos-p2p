package main

import (
	"github.com/eoscanada/eos-go"
	"go.uber.org/zap"

	"github.com/fanyang1988/eos-p2p/p2p"
)

// MsgHandler p2p.MsgHandler imp
type MsgHandler struct {
}

// OnHandshakeMsg handler func imp
func (m *MsgHandler) OnHandshakeMsg(peer *p2p.Peer, msg *eos.HandshakeMessage) {
	Logger.Info("on handshake", zap.Int16("generation", msg.Generation))
}

// OnGoAwayMsg handler func imp
func (m *MsgHandler) OnGoAwayMsg(peer *p2p.Peer, msg *eos.GoAwayMessage) {

}

// OnTimeMsg handler func imp
func (m *MsgHandler) OnTimeMsg(peer *p2p.Peer, msg *eos.TimeMessage) {

}

// OnNoticeMsg handler func imp
func (m *MsgHandler) OnNoticeMsg(peer *p2p.Peer, msg *eos.NoticeMessage) {

}

// OnRequestMsg handler func imp
func (m *MsgHandler) OnRequestMsg(peer *p2p.Peer, msg *eos.RequestMessage) {

}

// OnSyncRequestMsg handler func imp
func (m *MsgHandler) OnSyncRequestMsg(peer *p2p.Peer, msg *eos.SyncRequestMessage) {

}

// OnSignedBlock handler func imp
func (m *MsgHandler) OnSignedBlock(peer *p2p.Peer, msg *eos.SignedBlock) {

}

// OnPackedTransactionMsg handler func imp
func (m *MsgHandler) OnPackedTransactionMsg(peer *p2p.Peer, msg *eos.PackedTransactionMessage) {

}
