package p2p

import (
	"context"
	"encoding/hex"
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/eosforce/eos-p2p/store"
)

type peerStatusTyp uint8

const (
	peerStatNormal = peerStatusTyp(iota)
	peerStatInit
	peerStatError
	peerStatClosed
)

type peerStatus struct {
	peer   *Peer
	cfg    *PeerCfg
	status peerStatusTyp
}

// Client a p2p Client for eos chain
type Client struct {
	ps       map[string]*peerStatus
	handlers []Handler

	// for sync
	syncHandler     Handler
	sync            *syncManager
	currentSyncPeer *Peer // will changed by peerMng loop
	needSync        bool

	chainID Checksum256

	packetChan chan envelopMsg
	peerChan   chan peerMsg

	blkStorer store.BlockStorer

	logger *zap.Logger

	wg sync.WaitGroup
}

// Options options for new client
type Options struct {
	needSync      bool
	startBlockNum uint32
	handlers      []Handler
	blkStorer     store.BlockStorer
	logger        *zap.Logger
}

// OptionFunc func for new client
type OptionFunc func(*Options) error

// WithNeedSync set client with needSync
func WithNeedSync(startBlockNum uint32) OptionFunc {
	return func(o *Options) error {
		o.needSync = true
		o.startBlockNum = startBlockNum
		return nil
	}
}

// WithHandler set client with a handler
func WithHandler(h Handler) OptionFunc {
	return func(o *Options) error {
		o.handlers = append(o.handlers, h)
		return nil
	}
}

// WithStorer set storer for blocks and state
func WithStorer(blk store.BlockStorer) OptionFunc {
	return func(o *Options) error {
		o.blkStorer = blk
		return nil
	}
}

// WithLogger set logger for log
func WithLogger(l *zap.Logger) OptionFunc {
	return func(o *Options) error {
		o.logger = l
		return nil
	}
}

// NewClient create new client
func NewClient(ctx context.Context, chainID string, peers []*PeerCfg, opts ...OptionFunc) (*Client, error) {
	if len(peers) == 0 {
		return nil, errors.New("NoPeerCfg")
	}

	defaultOpts := Options{
		handlers: make([]Handler, 0, 8),
	}

	for _, o := range opts {
		err := o(&defaultOpts)
		if err != nil {
			return nil, err
		}
	}

	cID, err := hex.DecodeString(chainID)
	if err != nil {
		return nil, errors.Wrapf(err, "decode chainID error")
	}

	client := &Client{
		ps:         make(map[string]*peerStatus, 64),
		packetChan: make(chan envelopMsg, 256),
		peerChan:   make(chan peerMsg, 8),
		handlers:   make([]Handler, 0, len(defaultOpts.handlers)+1+32),
		chainID:    cID,
		needSync:   defaultOpts.needSync,
		blkStorer:  defaultOpts.blkStorer,
		logger:     defaultOpts.logger,
	}

	// create sync manager
	client.sync = &syncManager{
		cli: client,
	}
	client.sync.init(defaultOpts.needSync)

	// init handlers
	for _, h := range defaultOpts.handlers {
		client.handlers = append(client.handlers, h)
	}

	err = client.Start(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "start client error")
	}

	for _, p := range peers {
		client.NewPeer(p)
	}

	return client, nil
}

func (c *Client) closeAllPeer() {
	for _, p := range c.ps {
		p.peer.ClosePeer()
	}

	for _, p := range c.ps {
		p.peer.Wait()
	}
}

// Start start client process goroutine
func (c *Client) Start(ctx context.Context) error {
	c.logger.Info("Starting client")

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.peerLoop(ctx)
	}()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.peerMngLoop(ctx)
	}()

	return nil
}

// Wait wait client closed
func (c *Client) Wait() {
	c.wg.Wait()
}

// ChainID get chainID from the net client to connect
func (c *Client) ChainID() Checksum256 {
	return c.chainID
}

// HeadBlockNum get head block number current
func (c *Client) HeadBlockNum() uint32 {
	return c.blkStorer.HeadBlockNum()
}

// SetHeadBlock set head block number current
func (c *Client) SetHeadBlock(blk *SignedBlock) error {
	err := c.blkStorer.CommitBlock(blk)
	if err != nil {
		c.logger.Error("set block error %s", zap.Error(err))
	}

	return errors.Wrapf(err, "set head block")
}
