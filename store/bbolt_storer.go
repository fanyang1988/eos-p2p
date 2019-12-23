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

// BlockDBState head state and chain state
type BlockDBState struct {
	HeadBlockNum             uint32             `json:"headNum"`
	HeadBlockID              types.Checksum256  `json:"headID"`
	HeadBlockTime            time.Time          `json:"t"`
	LastIrreversibleBlockNum uint32             `json:"irrNum"`
	LastIrreversibleBlockID  types.Checksum256  `json:"irrID"`
	HeadBlock                *types.SignedBlock `json:"blk"`
}

// NewBlockDBState new stat
func NewBlockDBState() *BlockDBState {
	return &BlockDBState{
		HeadBlockNum: 1,
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

// BBoltStorer a very simple storer imp for test imp by storer
type BBoltStorer struct {
	chainID types.Checksum256
	db      *bolt.DB
	logger  *zap.Logger
	mutex   sync.RWMutex

	// current store state
	state *BlockDBState
}

// NewBBoltStorer create a bbolt storer
func NewBBoltStorer(logger *zap.Logger, chainID string, dbPath string) (*BBoltStorer, error) {
	cID, err := hex.DecodeString(chainID)
	if err != nil {
		return nil, errors.Wrapf(err, "decode chainID error")
	}

	db, err := bolt.Open(dbPath, 0666, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "create storer %s", dbPath)
	}

	res := &BBoltStorer{
		chainID: cID,
		db:      db,
		logger:  logger,
		state:   NewBlockDBState(),
	}

	if err := res.initState(); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *BBoltStorer) initState() error {
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
	return s.state.HeadBlock // TODO: need deep copy
}

// HeadBlockID get HeadBlockID
func (s *BBoltStorer) HeadBlockID() types.Checksum256 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.state.HeadBlockID
}

// LastIrreversibleBlockNum get LastIrreversibleBlockNum
func (s *BBoltStorer) LastIrreversibleBlockNum() uint32 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.state.LastIrreversibleBlockNum
}

// LastIrreversibleBlockID get LastIrreversibleBlockID
func (s *BBoltStorer) LastIrreversibleBlockID() types.Checksum256 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.state.LastIrreversibleBlockID
}

func (s *BBoltStorer) updateStatByBlock(blk *types.SignedBlock) error {
	// Just set to block state
	s.state.HeadBlockNum = blk.BlockNumber()
	s.state.HeadBlockID, _ = blk.BlockID()
	s.state.HeadBlockTime = blk.Timestamp.Time
	s.state.LastIrreversibleBlockNum = s.state.HeadBlockNum
	s.state.LastIrreversibleBlockID = s.state.HeadBlockID

	if s.state.HeadBlockNum%1000 == 0 {
		s.logger.Info("on block head", zap.Uint32("blockNum", s.state.HeadBlockNum))
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

	return s.setBlock(blk)
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
