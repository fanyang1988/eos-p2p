package p2p

import (
	"io"

	eos "github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"go.uber.org/zap"
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

	goAwayNoReason       = eos.GoAwayNoReason
	goAwaySelfConnect    = eos.GoAwaySelfConnect
	goAwayDuplicate      = eos.GoAwayDuplicate
	goAwayWrongChain     = eos.GoAwayWrongChain
	goAwayWrongVersion   = eos.GoAwayWrongVersion
	goAwayForked         = eos.GoAwayForked
	goAwayUnlinkable     = eos.GoAwayUnlinkable
	goAwayBadTransaction = eos.GoAwayBadTransaction
	goAwayValidation     = eos.GoAwayValidation
	goAwayAuthentication = eos.GoAwayAuthentication
	goAwayFatalOther     = eos.GoAwayFatalOther
	goAwayBenignOther    = eos.GoAwayBenignOther
	goAwayCrazy          = eos.GoAwayCrazy
)

// CurveK1 ecc types
const CurveK1 = ecc.CurveK1

// Signature ecc types
type Signature = ecc.Signature

// PublicKey ecc types
type PublicKey = ecc.PublicKey

// newPublicKey create public key
func newPublicKey(pubKey string) (out PublicKey, err error) {
	return ecc.NewPublicKey(pubKey)
}

// EnableP2PLogging enable p2p package to log by zap
func EnableP2PLogging() *zap.Logger {
	p2pLog = eos.NewLogger(false)
	return p2pLog
}

func readEOSPacket(r io.Reader) (packet *Packet, err error) {
	return eos.ReadPacket(r)
}

func newEOSEncoder(w io.Writer) *eos.Encoder {
	return eos.NewEncoder(w)
}
