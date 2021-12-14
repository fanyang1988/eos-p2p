package p2p

import (
	"github.com/fanyang1988/eos-p2p/types"
)

// Message msg
type Message = types.Message

// HandshakeMessage  eos msg type
type HandshakeMessage = types.HandshakeMessage

// GoAwayMessage eos msg type
type GoAwayMessage = types.GoAwayMessage

// TimeMessage eos msg type
type TimeMessage = types.TimeMessage

// NoticeMessage eos msg type
type NoticeMessage = types.NoticeMessage

// RequestMessage eos msg type
type RequestMessage = types.RequestMessage

// SyncRequestMessage eos msg type
type SyncRequestMessage = types.SyncRequestMessage

// SignedBlock eos msg type
type SignedBlock = types.SignedBlock

// PackedTransactionMessage eos msg type
type PackedTransactionMessage = types.PackedTransactionMessage

// Packet eos msg type
type Packet = types.Packet

// Checksum256 checksum256 from eos
type Checksum256 = types.Checksum256

// OrderedBlockIDs eos type
type OrderedBlockIDs = types.OrderedBlockIDs

// Tstamp eos type
type Tstamp = types.Tstamp

// GoAwayReason eos type
type GoAwayReason = types.GoAwayReason

const (
	// GoAwayXXX goaway reason

	goAwayNoReason       = types.GoAwayNoReason
	goAwaySelfConnect    = types.GoAwaySelfConnect
	goAwayDuplicate      = types.GoAwayDuplicate
	goAwayWrongChain     = types.GoAwayWrongChain
	goAwayWrongVersion   = types.GoAwayWrongVersion
	goAwayForked         = types.GoAwayForked
	goAwayUnlinkable     = types.GoAwayUnlinkable
	goAwayBadTransaction = types.GoAwayBadTransaction
	goAwayValidation     = types.GoAwayValidation
	goAwayAuthentication = types.GoAwayAuthentication
	goAwayFatalOther     = types.GoAwayFatalOther
	goAwayBenignOther    = types.GoAwayBenignOther
)

// CurveK1 ecc types
const CurveK1 = types.CurveK1

// Signature ecc types
type Signature = types.Signature

// PublicKey ecc types
type PublicKey = types.PublicKey

// HandshakeInfo handshake state for peer
type HandshakeInfo = types.HandshakeInfo
