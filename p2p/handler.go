package p2p

import (
	"encoding/json"

	"github.com/eoscanada/eos-go"
	"go.uber.org/zap"
)

type Handler interface {
	Handle(envelope *Envelope)
}

type HandlerFunc func(envelope *Envelope)

func (f HandlerFunc) Handle(envelope *Envelope) {
	f(envelope)
}

// LoggerHandler logs the messages back and forth.
var LoggerHandler = HandlerFunc(func(envelope *Envelope) {
	data, err := json.Marshal(envelope)
	if err != nil {
		logErr("Marshal err", err)
		return
	}

	p2pLog.Info("handler", zap.String("message", string(data)))
})

// StringLoggerHandler simply prints the messages as they go through the client.
var StringLoggerHandler = HandlerFunc(func(envelope *Envelope) {
	name, _ := envelope.Packet.Type.Name()
	p2pLog.Info(
		"handler Packet",
		zap.String("name", name),
		zap.String("sender", envelope.Sender.Address),
		zap.String("receiver", envelope.Receiver.Address),
		zap.Stringer("msg", envelope.Packet.P2PMessage), // this will use by String()
	)
})

// MsgHandler handler for each msg
type MsgHandler interface {
	OnHandshakeMsg(envelope *Envelope, msg *eos.HandshakeMessage)
	//OnChainSizeMsg(envelope *Envelope, msg *eos.ChainSizeMessage)
	OnGoAwayMsg(envelope *Envelope, msg *eos.GoAwayMessage)
	OnTimeMsg(envelope *Envelope, msg *eos.TimeMessage)
	OnNoticeMsg(envelope *Envelope, msg *eos.NoticeMessage)
	OnRequestMsg(envelope *Envelope, msg *eos.RequestMessage)
	OnSyncRequestMsg(envelope *Envelope, msg *eos.SyncRequestMessage)
	OnSignedBlock(envelope *Envelope, msg *eos.SignedBlock)
	OnPackedTransactionMsg(envelope *Envelope, msg *eos.PackedTransactionMessage)
}

// MsgHandlerImp MsgHandler for p2p msg
type MsgHandlerImp struct {
	handler MsgHandler
}

// NewMsgHandler create a msg Handler by MsgHandler
func NewMsgHandler(handler MsgHandler) *MsgHandlerImp {
	return &MsgHandlerImp{
		handler: handler,
	}
}

// Handle implements Handler interface
func (m MsgHandlerImp) Handle(envelope *Envelope) {
	switch envelope.Packet.P2PMessage.(type) {
	case *eos.HandshakeMessage:
		handshakeMessage, ok := envelope.Packet.P2PMessage.(*eos.HandshakeMessage)
		if ok && handshakeMessage != nil {
			m.handler.OnHandshakeMsg(envelope, handshakeMessage)
		}
	// Now EOS has not use this msg
	//case *eos.ChainSizeMessage:
	//	chainSizeMessage, ok := envelope.Packet.P2PMessage.(*eos.ChainSizeMessage)
	//	if ok && chainSizeMessage != nil {
	//		m.handler.OnChainSizeMsg(envelope, chainSizeMessage)
	//	}
	case *eos.GoAwayMessage:
		goAwayMsg, ok := envelope.Packet.P2PMessage.(*eos.GoAwayMessage)
		if ok && goAwayMsg != nil {
			m.handler.OnGoAwayMsg(envelope, goAwayMsg)
		}
	case *eos.TimeMessage:
		timeMessage, ok := envelope.Packet.P2PMessage.(*eos.TimeMessage)
		if ok && timeMessage != nil {
			m.handler.OnTimeMsg(envelope, timeMessage)
		}
	case *eos.NoticeMessage:
		noticeMessage, ok := envelope.Packet.P2PMessage.(*eos.NoticeMessage)
		if ok && noticeMessage != nil {
			m.handler.OnNoticeMsg(envelope, noticeMessage)
		}
	case *eos.RequestMessage:
		requestMessage, ok := envelope.Packet.P2PMessage.(*eos.RequestMessage)
		if ok && requestMessage != nil {
			m.handler.OnRequestMsg(envelope, requestMessage)
		}
	case *eos.SyncRequestMessage:
		syncRequestMessage, ok := envelope.Packet.P2PMessage.(*eos.SyncRequestMessage)
		if ok && syncRequestMessage != nil {
			m.handler.OnSyncRequestMsg(envelope, syncRequestMessage)
		}
	case *eos.SignedBlock:
		signedBlock, ok := envelope.Packet.P2PMessage.(*eos.SignedBlock)
		if ok && signedBlock != nil {
			m.handler.OnSignedBlock(envelope, signedBlock)
		}
	case *eos.PackedTransactionMessage:
		packedTransactionMessage, ok := envelope.Packet.P2PMessage.(*eos.PackedTransactionMessage)
		if ok && packedTransactionMessage != nil {
			m.handler.OnPackedTransactionMsg(envelope, packedTransactionMessage)
		}
	default:
	}
}
