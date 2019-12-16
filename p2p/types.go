package p2p

import (
	"io"

	eos "github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"go.uber.org/zap"
)

type P2PMessage = eos.P2PMessage
type HandshakeMessage = eos.HandshakeMessage
type GoAwayMessage = eos.GoAwayMessage
type TimeMessage = eos.TimeMessage
type NoticeMessage = eos.NoticeMessage
type RequestMessage = eos.RequestMessage
type SyncRequestMessage = eos.SyncRequestMessage
type SignedBlock = eos.SignedBlock
type PackedTransactionMessage = eos.PackedTransactionMessage

type Packet = eos.Packet

type Checksum256 = eos.Checksum256
type OrderedBlockIDs = eos.OrderedBlockIDs
type Tstamp = eos.Tstamp
type GoAwayReason = eos.GoAwayReason

const GoAwayNoReason = eos.GoAwayNoReason
const GoAwayValidation = eos.GoAwayValidation

const CurveK1 = ecc.CurveK1

type Signature = ecc.Signature

type PublicKey = ecc.PublicKey

func NewPublicKey(pubKey string) (out PublicKey, err error) {
	return ecc.NewPublicKey(pubKey)
}

// EnableP2PLogging enable p2p package to log by zap
func EnableP2PLogging() *zap.Logger {
	p2pLog = eos.NewLogger(false)
	return p2pLog
}

func ReadPacket(r io.Reader) (packet *Packet, err error) {
	return eos.ReadPacket(r)
}

func NewEncoder(w io.Writer) *eos.Encoder {
	return eos.NewEncoder(w)
}
