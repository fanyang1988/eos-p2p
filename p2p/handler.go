package p2p

import (
	"encoding/json"

	"go.uber.org/zap"
)

// Handler interface for peer
type Handler interface {
	Handle(envelope *Envelope)
}

// HandlerFunc a func for Handler
type HandlerFunc func(envelope *Envelope)

// Handle imp handle
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
		zap.Stringer("msg", envelope.Packet.P2PMessage), // this will use by String()
	)
})

// MsgHandler handler for each msg
type MsgHandler interface {
	OnHandshakeMsg(peer *Peer, msg *HandshakeMessage)
	OnGoAwayMsg(peer *Peer, msg *GoAwayMessage)
	OnTimeMsg(peer *Peer, msg *TimeMessage)
	OnNoticeMsg(peer *Peer, msg *NoticeMessage)
	OnRequestMsg(peer *Peer, msg *RequestMessage)
	OnSyncRequestMsg(peer *Peer, msg *SyncRequestMessage)
	OnSignedBlock(peer *Peer, msg *SignedBlock)
	OnPackedTransactionMsg(peer *Peer, msg *PackedTransactionMessage)
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
	case *HandshakeMessage:
		handshakeMessage, ok := envelope.Packet.P2PMessage.(*HandshakeMessage)
		if ok && handshakeMessage != nil {
			m.handler.OnHandshakeMsg(envelope.Sender, handshakeMessage)
		}
	// Now EOS has not use *ChainSizeMessage
	case *GoAwayMessage:
		goAwayMsg, ok := envelope.Packet.P2PMessage.(*GoAwayMessage)
		if ok && goAwayMsg != nil {
			m.handler.OnGoAwayMsg(envelope.Sender, goAwayMsg)
		}
	case *TimeMessage:
		timeMessage, ok := envelope.Packet.P2PMessage.(*TimeMessage)
		if ok && timeMessage != nil {
			m.handler.OnTimeMsg(envelope.Sender, timeMessage)
		}
	case *NoticeMessage:
		noticeMessage, ok := envelope.Packet.P2PMessage.(*NoticeMessage)
		if ok && noticeMessage != nil {
			m.handler.OnNoticeMsg(envelope.Sender, noticeMessage)
		}
	case *RequestMessage:
		requestMessage, ok := envelope.Packet.P2PMessage.(*RequestMessage)
		if ok && requestMessage != nil {
			m.handler.OnRequestMsg(envelope.Sender, requestMessage)
		}
	case *SyncRequestMessage:
		syncRequestMessage, ok := envelope.Packet.P2PMessage.(*SyncRequestMessage)
		if ok && syncRequestMessage != nil {
			m.handler.OnSyncRequestMsg(envelope.Sender, syncRequestMessage)
		}
	case *SignedBlock:
		signedBlock, ok := envelope.Packet.P2PMessage.(*SignedBlock)
		if ok && signedBlock != nil {
			m.handler.OnSignedBlock(envelope.Sender, signedBlock)
		}
	case *PackedTransactionMessage:
		packedTransactionMessage, ok := envelope.Packet.P2PMessage.(*PackedTransactionMessage)
		if ok && packedTransactionMessage != nil {
			m.handler.OnPackedTransactionMsg(envelope.Sender, packedTransactionMessage)
		}
	default:
	}
}
