package store

import (
	"bytes"
	"time"

	"github.com/fanyang1988/eos-p2p/types"
)

const maxBlocksHoldInDBStat = 64

// BlockDBState head state and chain state
type BlockDBState struct {
	ChainID       types.Checksum256    `json:"chainID"`
	HeadBlockNum  uint32               `json:"headNum"`
	HeadBlockID   types.Checksum256    `json:"headID"`
	HeadBlockTime time.Time            `json:"headTime"`
	HeadBlock     *types.SignedBlock   `json:"headBlk"`
	LastBlocks    []*types.SignedBlock `json:"blks"`
}

// ToHandshakeInfo make a handshake info for handshake message
func (b *BlockDBState) ToHandshakeInfo() *types.HandshakeInfo {
	// TODO: a very simple irr
	irrNum := b.HeadBlockNum - 9
	if irrNum < 1 {
		irrNum = 1
	}

	res := &types.HandshakeInfo{
		ChainID:      b.ChainID,
		HeadBlockNum: 1,
	}

	head := b.HeadBlock
	if head != nil && len(head.Previous) > 0 {
		res.HeadBlockNum = head.BlockNumber()
		res.HeadBlockID, _ = head.BlockID()
		res.HeadBlockTime = head.Timestamp.Time
	}

	irr, ok := b.getBlockByNum(irrNum)
	if ok {
		res.LastIrreversibleBlockNum = irr.BlockNumber()
		res.LastIrreversibleBlockID, _ = irr.BlockID()
	}

	return res
}

// NewBlockDBState new stat
func NewBlockDBState(chainID types.Checksum256) *BlockDBState {
	return &BlockDBState{
		ChainID:      chainID,
		HeadBlockNum: 1,
		HeadBlock:    types.NewEmptyBlock(),
		LastBlocks:   make([]*types.SignedBlock, 0, maxBlocksHoldInDBStat+1),
	}
}

// Bytes to bytes to store
func (b *BlockDBState) Bytes() ([]byte, error) {
	var buffer bytes.Buffer
	encoder := types.NewEncoder(&buffer)

	if err := encoder.Encode(b); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// FromBytes from Bytes
func (b *BlockDBState) FromBytes(data []byte) error {
	decoder := types.NewDecoder(data)
	decoder.DecodeActions(false)
	return decoder.Decode(b)
}

// getBlockByNum get block by num, if not store all blocks, try to find in state cache
func (b *BlockDBState) getBlockByNum(blockNum uint32) (*types.SignedBlock, bool) {
	if len(b.LastBlocks) == 0 {
		return nil, false
	}

	baseNum := b.LastBlocks[0].BlockNumber()

	if blockNum < baseNum ||
		blockNum > b.LastBlocks[len(b.LastBlocks)-1].BlockNumber() {
		return nil, false
	}

	return b.LastBlocks[blockNum-baseNum], true
}
