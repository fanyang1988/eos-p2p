package store

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"

	"github.com/fanyang1988/eos-p2p/types"
)

// BlockDBState head state and chain state
type BlockDBState struct {
	HeadBlockNum             uint32            `json:"headNum"`
	HeadBlockID              types.Checksum256 `json:"headID"`
	HeadBlockTime            time.Time         `json:"t"`
	LastIrreversibleBlockNum uint32            `json:"IrrNum"`
	LastIrreversibleBlockID  types.Checksum256 `json:"IrrID"`
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

	// current store state
	state *BlockDBState
}

// NewBBoltStorer create a bbolt storer
func NewBBoltStorer(logger *zap.Logger, chainID types.Checksum256, dbPath string) (*BBoltStorer, error) {
	db, err := bolt.Open(dbPath, 0666, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "create storer %s", dbPath)
	}

	return &BBoltStorer{
		chainID: chainID,
		db:      db,
		logger:  logger,
		state:   NewBlockDBState(),
	}, nil
}

func (s *BBoltStorer) initState() error {
	return s.db.Update(func(tx *bolt.Tx) error {
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
	})
}

// ChainID get chainID
func (s *BBoltStorer) ChainID() types.Checksum256 {
	return s.chainID
}

// HeadBlockNum get headBlockNum
func (s *BBoltStorer) HeadBlockNum() uint32 {
	return s.state.HeadBlockNum
}

// HeadBlockID get HeadBlockID
func (s *BBoltStorer) HeadBlockID() types.Checksum256 {
	return s.state.HeadBlockID
}

// LastIrreversibleBlockNum get LastIrreversibleBlockNum
func (s *BBoltStorer) LastIrreversibleBlockNum() uint32 {
	return s.state.LastIrreversibleBlockNum
}

// LastIrreversibleBlockID get LastIrreversibleBlockID
func (s *BBoltStorer) LastIrreversibleBlockID() types.Checksum256 {
	return s.state.LastIrreversibleBlockID
}

func (s *BBoltStorer) updateStatByBlock(blk *types.SignedBlock) error {
	return nil
}

// CommitBlock commit block from p2p
func (s *BBoltStorer) CommitBlock(blk *types.SignedBlock) error {

	return nil
}

// CommitTrx commit trx
func (s *BBoltStorer) CommitTrx(trx *types.PackedTransactionMessage) error {
	return nil
}

// Flush flush to db
func (s *BBoltStorer) Flush() error {
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
