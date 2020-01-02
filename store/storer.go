package store

import "github.com/fanyang1988/eos-p2p/types"

// BlockStorer interface for storer to db
type BlockStorer interface {
	ChainID() types.Checksum256
	HeadBlockNum() uint32
	CommitBlock(blk *types.SignedBlock) error
	State() BlockDBState
	GetBlockByNum(blockNum uint32) (*types.SignedBlock, bool)
	Flush() error
	Close()
	Wait()
}
