package types

import (
	"fmt"
	"github.com/pkg/errors"
	"time"

	eos "github.com/eosforce/goeosforce"
	"github.com/eosforce/goeosforce/ecc"
	"go.uber.org/zap/zapcore"
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

// BlockTimestamp block time
type BlockTimestamp = eos.BlockTimestamp

// HandshakeInfo handshake state for peer
type HandshakeInfo struct {
	ChainID                  Checksum256
	HeadBlockNum             uint32
	HeadBlockID              Checksum256
	HeadBlockTime            time.Time
	LastIrreversibleBlockNum uint32
	LastIrreversibleBlockID  Checksum256
}

func (h *HandshakeInfo) String() string {
	return fmt.Sprintf("Handshake Info: Head[%d], LastIrreversible[%d]",
		h.HeadBlockNum, h.LastIrreversibleBlockNum)
}

// MarshalLogObject calls the underlying function from zap.
func (h HandshakeInfo) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("chainID", h.ChainID.String())
	enc.AddUint32("headBlockNum", h.HeadBlockNum)
	enc.AddString("headBlockID", h.HeadBlockID.String())
	enc.AddTime("headBlockTime", h.HeadBlockTime)
	enc.AddUint32("lastIrreversibleBlockNum", h.LastIrreversibleBlockNum)
	enc.AddString("lastIrreversibleBlockID", h.LastIrreversibleBlockID.String())
	return nil
}

// NewEmptyBlock create an empty block can be unmarshal data
func NewEmptyBlock() *SignedBlock {
	return &SignedBlock{
		SignedBlockHeader: eos.SignedBlockHeader{
			BlockHeader: eos.BlockHeader{
				NewProducers: &eos.OptionalProducerSchedule{
					ProducerSchedule: eos.ProducerSchedule{
						Producers: make([]eos.ProducerKey, 0, 23),
					},
				},
				HeaderExtensions: make([]*eos.Extension, 0, 4),
			},
		},
		Transactions:    make([]eos.TransactionReceipt, 0, 256),
		BlockExtensions: make([]*eos.Extension, 0, 4),
	}
}

func CopyChecksum256(c Checksum256) Checksum256 {
	b := make([]byte, 0, len(c))
	b = append(b, c...)
	return Checksum256(b)
}

func CopyBytes(c []byte) []byte {
	res := make([]byte, 0, len(c))
	res = append(res, c...)
	return res
}

func CopySignature(c ecc.Signature) ecc.Signature {
	res, err := ecc.NewSignature(c.String())
	if err != nil {
		panic(errors.Wrapf(err, "copy signature"))
	}
	return res
}

func CopyExt(c *eos.Extension) *eos.Extension {
	return &eos.Extension{
		Type: c.Type,
		Data: CopyBytes(c.Data),
	}
}

// DeepCopyBlock a deep copy for a block
func DeepCopyBlock(b *SignedBlock) (*SignedBlock, error) {
	res := &SignedBlock{
		SignedBlockHeader: eos.SignedBlockHeader{
			BlockHeader: eos.BlockHeader{
				Timestamp:        b.Timestamp,
				Producer:         b.Producer,
				Confirmed:        b.Confirmed,
				Previous:         CopyChecksum256(b.Previous),
				TransactionMRoot: CopyChecksum256(b.TransactionMRoot),
				ActionMRoot:      CopyChecksum256(b.ActionMRoot),
				ScheduleVersion:  b.ScheduleVersion,
				HeaderExtensions: make([]*eos.Extension, 0, len(b.HeaderExtensions)),
			},
			ProducerSignature: CopySignature(b.ProducerSignature),
		},
		Transactions:    make([]eos.TransactionReceipt, 0, len(b.Transactions)),
		BlockExtensions: make([]*eos.Extension, 0, len(b.BlockExtensions)),
	}

	if b.NewProducers != nil {
		res.NewProducers = &eos.OptionalProducerSchedule{
			ProducerSchedule: eos.ProducerSchedule{
				Version:   b.NewProducers.Version,
				Producers: make([]eos.ProducerKey, 0, len(b.NewProducers.Producers)),
			},
		}
		for _, prod := range b.NewProducers.Producers {
			res.NewProducers.Producers = append(res.NewProducers.Producers, eos.ProducerKey{
				AccountName:     prod.AccountName,
				BlockSigningKey: prod.BlockSigningKey,
			})
		}
	}

	for _, ext := range b.HeaderExtensions {
		res.HeaderExtensions = append(res.HeaderExtensions, CopyExt(ext))
	}

	for _, ext := range b.BlockExtensions {
		res.BlockExtensions = append(res.BlockExtensions, CopyExt(ext))
	}

	for _, trx := range b.Transactions {
		trxCopy := eos.TransactionReceipt{
			TransactionReceiptHeader: trx.TransactionReceiptHeader,
			Transaction: eos.TransactionWithID{
				ID: CopyChecksum256(trx.Transaction.ID),
				Packed: &eos.PackedTransaction{
					Signatures:            make([]ecc.Signature, 0, len(trx.Transaction.Packed.Signatures)),
					Compression:           trx.Transaction.Packed.Compression,
					PackedContextFreeData: CopyBytes(trx.Transaction.Packed.PackedContextFreeData),
					PackedTransaction:     CopyBytes(trx.Transaction.Packed.PackedTransaction),
				},
			},
		}
		for _, s := range trx.Transaction.Packed.Signatures {
			trxCopy.Transaction.Packed.Signatures = append(trxCopy.Transaction.Packed.Signatures, CopySignature(s))
		}
		res.Transactions = append(res.Transactions, trxCopy)
	}

	return res, nil
}
