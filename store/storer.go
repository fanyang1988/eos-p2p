package store

import "github.com/fanyang1988/eos-p2p/types"

// BlockStorer interface for storer to db
type BlockStorer interface {
	ChainID() types.Checksum256
	HeadBlockNum() uint32
	HeadBlockID() types.Checksum256
	HeadBlock() *types.SignedBlock
	LastIrreversibleBlockNum() uint32
	LastIrreversibleBlockID() types.Checksum256
	CommitBlock(blk *types.SignedBlock) error
	CommitTrx(trx *types.PackedTransactionMessage) error
	State() BlockDBState
	Flush() error
	Close()
	Wait()
}
