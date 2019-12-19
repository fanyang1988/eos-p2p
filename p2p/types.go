package p2p

import (
	"encoding/binary"
	"encoding/hex"
	"io"
	"net"

	eos "github.com/eosforce/goeosforce"
	"github.com/eosforce/goeosforce/ecc"
	"github.com/pkg/errors"
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

func readEOSPacket(r io.Reader, conn net.Conn) (packet *Packet, err error) {
	return readPacket(r, conn)
}

func newEOSEncoder(w io.Writer, conn net.Conn) *eos.Encoder {
	return eos.NewEncoder(w)
}

func readPacket(r io.Reader, conn net.Conn) (*Packet, error) {
	data := make([]byte, 0)

	lengthBytes := make([]byte, 4, 4)

	if _, err := io.ReadFull(r, lengthBytes); err != nil {
		return nil, errors.Wrapf(err, "readfull length")
	}

	data = append(data, lengthBytes...)

	size := binary.LittleEndian.Uint32(lengthBytes)

	if size > 16*1024*1024 {
		return nil, errors.Errorf("packet is too large %d", size)
	}

	payloadBytes := make([]byte, size, size)

	count, err := io.ReadFull(r, payloadBytes)

	if err != nil {
		return nil, errors.Errorf("read full data error")
	}

	if count != int(size) {
		return nil, errors.Errorf("read full not full read[%d] expected[%d]", count, size)
	}

	data = append(data, payloadBytes...)

	packet := &Packet{}
	decoder := eos.NewDecoder(data)
	decoder.DecodeActions(false)
	err = decoder.Decode(packet)
	if err != nil {
		return nil, errors.Wrapf(err, "Failing decode data %s", hex.EncodeToString(data))
	}
	packet.Raw = data
	return packet, nil
}
