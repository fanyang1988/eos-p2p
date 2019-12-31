package store

import (
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"

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
	if head != nil {
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
		LastBlocks:   make([]*types.SignedBlock, 0, maxBlocksHoldInDBStat+1),
	}
}

// Bytes to bytes to store
func (b *BlockDBState) Bytes() ([]byte, error) {
	return json.Marshal(*b)
}

// FromBytes from Bytes
func (b *BlockDBState) FromBytes(data []byte) error {
	return json.Unmarshal(data, b)
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

// BBoltStorer a very simple storer imp for test imp by storer
type BBoltStorer struct {
	chainID types.Checksum256
	db      *bolt.DB
	logger  *zap.Logger
	mutex   sync.RWMutex

	isStoreAllBlocks bool

	// current store state
	state *BlockDBState
}

// NewBBoltStorer create a bbolt storer
func NewBBoltStorer(logger *zap.Logger, chainID string, dbPath string, isStoreBlocks bool) (*BBoltStorer, error) {
	cID, err := hex.DecodeString(chainID)
	if err != nil {
		return nil, errors.Wrapf(err, "decode chainID error")
	}

	db, err := bolt.Open(dbPath, 0666, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "create storer %s", dbPath)
	}

	res := &BBoltStorer{
		chainID:          cID,
		db:               db,
		logger:           logger,
		isStoreAllBlocks: isStoreBlocks,
		state:            NewBlockDBState(cID),
	}

	if err := res.initState(cID); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *BBoltStorer) initState(chainID types.Checksum256) error {
	return errors.Wrap(s.db.Update(func(tx *bolt.Tx) error {
		stateBucket, err := tx.CreateBucketIfNotExists([]byte("state"))
		if err != nil {
			return errors.Wrap(err, "initState create")
		}

		stateBytes := stateBucket.Get([]byte("stat"))

		s.logger.Debug("headstate", zap.String("stat", string(stateBytes)))

		if len(stateBytes) == 0 {
			s.logger.Debug("init head state")
			bytes, err := s.state.Bytes()
			if err != nil {
				return errors.Wrap(err, "new state to byte")
			}

			if err := stateBucket.Put([]byte("stat"), bytes); err != nil {
				return errors.Wrap(err, "put new stat")
			}
		} else {
			if err := s.state.FromBytes(stateBytes); err != nil {
				return errors.Wrap(err, "state from data")
			}

			if !types.IsChecksumEq(s.chainID, s.state.ChainID) {
				return errors.Wrapf(err, "ChainID is diff in store, now: %s, store: %s",
					s.chainID.String(), s.state.ChainID.String())
			}
		}

		return nil
	}), "init state")
}

// ChainID get chainID
func (s *BBoltStorer) ChainID() types.Checksum256 {
	// const
	return s.chainID
}

// HeadBlockNum get headBlockNum
func (s *BBoltStorer) HeadBlockNum() uint32 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.state.HeadBlockNum
}

// State get state data
func (s *BBoltStorer) State() BlockDBState {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return *s.state
}

// HeadBlock get head block
func (s *BBoltStorer) HeadBlock() *types.SignedBlock {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if s.state.HeadBlock == nil {
		return nil
	}

	res, err := types.DeepCopyBlock(s.state.HeadBlock)
	if err != nil {
		s.logger.Error("deep copy error", zap.Error(err))
	}
	return res
}

// HeadBlockID get HeadBlockID
func (s *BBoltStorer) HeadBlockID() types.Checksum256 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.state.HeadBlockID
}

func (s *BBoltStorer) updateStatByBlock(blk *types.SignedBlock) error {
	// Just set to block state
	if s.state.HeadBlockNum >= blk.BlockNumber() {
		return nil
	}

	s.logger.Info("up block", zap.Uint32("blockNum", blk.BlockNumber()))

	s.state.HeadBlockNum = blk.BlockNumber()
	s.state.HeadBlockID, _ = blk.BlockID()
	s.state.HeadBlockTime = blk.Timestamp.Time
	s.state.HeadBlock, _ = types.DeepCopyBlock(blk)

	if s.state.HeadBlockNum%1000 == 0 {
		s.logger.Info("on block head", zap.Uint32("blockNum", s.state.HeadBlockNum))
	}

	s.state.LastBlocks = append(s.state.LastBlocks, s.state.HeadBlock)
	if len(s.state.LastBlocks) >= maxBlocksHoldInDBStat {
		for i := 0; i < len(s.state.LastBlocks)-1; i++ {
			s.state.LastBlocks[i] = s.state.LastBlocks[i+1]
		}
		s.state.LastBlocks = s.state.LastBlocks[:len(s.state.LastBlocks)-1]
	}

	return nil
}

func (s *BBoltStorer) setBlock(blk *types.SignedBlock) error {
	return errors.Wrapf(s.db.Update(func(tx *bolt.Tx) error {
		blockBucket, err := tx.CreateBucketIfNotExists([]byte("state"))
		if err != nil {
			return errors.Wrap(err, "create bucket")
		}

		bID, _ := blk.BlockID()

		jsonBytes, err := json.Marshal(*blk)
		if err != nil {
			return errors.Wrap(err, "json")
		}

		// Use json to shown for test
		if err := blockBucket.Put([]byte(bID), jsonBytes); err != nil {
			return errors.Wrap(err, "put block")
		}

		return nil
	}), "set block %d", blk.BlockNumber())
}

// CommitBlock commit block from p2p
func (s *BBoltStorer) CommitBlock(blk *types.SignedBlock) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := s.updateStatByBlock(blk); err != nil {
		return err
	}

	if s.isStoreAllBlocks {
		return s.setBlock(blk)
	}

	return nil
}

// GetBlockByNum get block by num, if not store all blocks, try to find in state cache
func (s *BBoltStorer) GetBlockByNum(blockNum uint32) (*types.SignedBlock, bool) {
	// TODO: find in blocks
	if !s.isStoreAllBlocks {
		return s.state.getBlockByNum(blockNum)
	}

	return nil, false
}

// CommitTrx commit trx
func (s *BBoltStorer) CommitTrx(trx *types.PackedTransactionMessage) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return nil
}

// Flush flush to db
func (s *BBoltStorer) Flush() error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stateBucket := tx.Bucket([]byte("state"))
	bytes, err := s.state.Bytes()
	if err != nil {
		return errors.Wrap(err, "new state to byte")
	}

	if err := stateBucket.Put([]byte("stat"), bytes); err != nil {
		return errors.Wrap(err, "put new stat")
	}

	// Commit the transaction.
	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "commit in flush")
	}
	return nil
}

// Close close storer and flush
func (s *BBoltStorer) Close() {
	if err := s.Flush(); err != nil {
		s.logger.Error("update state error in close", zap.Error(err))
	}
	s.db.Close()
}

// Wait wait closed
func (s *BBoltStorer) Wait() {
	return // no need wait
}
