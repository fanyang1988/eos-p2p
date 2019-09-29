package main

import (
	"github.com/eoscanada/eos-go"
	"github.com/fanyang1988/eos-p2p/p2p"
)

// MsgHandler p2p.MsgHandler imp
type MsgHandler struct {
}

// OnHandshakeMsg handler func imp
func (m *MsgHandler) OnHandshakeMsg(envelope *p2p.Envelope, msg *eos.HandshakeMessage) {

}

// OnGoAwayMsg handler func imp
func (m *MsgHandler) OnGoAwayMsg(envelope *p2p.Envelope, msg *eos.GoAwayMessage) {

}

// OnTimeMsg handler func imp
func (m *MsgHandler) OnTimeMsg(envelope *p2p.Envelope, msg *eos.TimeMessage) {

}

// OnNoticeMsg handler func imp
func (m *MsgHandler) OnNoticeMsg(envelope *p2p.Envelope, msg *eos.NoticeMessage) {

}

// OnRequestMsg handler func imp
func (m *MsgHandler) OnRequestMsg(envelope *p2p.Envelope, msg *eos.RequestMessage) {

}

// OnSyncRequestMsg handler func imp
func (m *MsgHandler) OnSyncRequestMsg(envelope *p2p.Envelope, msg *eos.SyncRequestMessage) {

}

// OnSignedBlock handler func imp
func (m *MsgHandler) OnSignedBlock(envelope *p2p.Envelope, msg *eos.SignedBlock) {

}

// OnPackedTransactionMsg handler func imp
func (m *MsgHandler) OnPackedTransactionMsg(envelope *p2p.Envelope, msg *eos.PackedTransactionMessage) {

}
