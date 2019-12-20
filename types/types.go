package types

import (
	eos "github.com/eosforce/goeosforce"
	"github.com/eosforce/goeosforce/ecc"
)

// Message msg
type Message = eos.P2PMessage

// HandshakeMessage  eos msg type
type HandshakeMessage = eos.HandshakeMessage

// GoAwayMessage eos msg type
type GoAwayMessage = eos.GoAwayMessage

// TimeMessage eos msg type
type TimeMessage = eos.TimeMessage

// NoticeMessage eos msg type
type NoticeMessage = eos.NoticeMessage

// RequestMessage eos msg type
type RequestMessage = eos.RequestMessage

// SyncRequestMessage eos msg type
type SyncRequestMessage = eos.SyncRequestMessage

// SignedBlock eos msg type
type SignedBlock = eos.SignedBlock

// PackedTransactionMessage eos msg type
type PackedTransactionMessage = eos.PackedTransactionMessage

// Packet eos msg type
type Packet = eos.Packet

// Checksum256 checksum256 from eos
type Checksum256 = eos.Checksum256

// OrderedBlockIDs eos type
type OrderedBlockIDs = eos.OrderedBlockIDs

// Tstamp eos type
type Tstamp = eos.Tstamp

// GoAwayReason eos type
type GoAwayReason = eos.GoAwayReason

const (
	// GoAwayXXX goaway reason

	GoAwayNoReason       = eos.GoAwayNoReason
	GoAwaySelfConnect    = eos.GoAwaySelfConnect
	GoAwayDuplicate      = eos.GoAwayDuplicate
	GoAwayWrongChain     = eos.GoAwayWrongChain
	GoAwayWrongVersion   = eos.GoAwayWrongVersion
	GoAwayForked         = eos.GoAwayForked
	GoAwayUnlinkable     = eos.GoAwayUnlinkable
	GoAwayBadTransaction = eos.GoAwayBadTransaction
	GoAwayValidation     = eos.GoAwayValidation
	GoAwayAuthentication = eos.GoAwayAuthentication
	GoAwayFatalOther     = eos.GoAwayFatalOther
	GoAwayBenignOther    = eos.GoAwayBenignOther
	GoAwayCrazy          = eos.GoAwayCrazy
)

// CurveK1 ecc types
const CurveK1 = ecc.CurveK1

// Signature ecc types
type Signature = ecc.Signature

// PublicKey ecc types
type PublicKey = ecc.PublicKey
