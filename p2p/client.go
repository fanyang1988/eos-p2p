package p2p

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Client a p2p Client for eos chain
type Client struct {
	peer        *Peer
	handlers    []Handler
	readTimeout time.Duration
	sync        *syncManager

	wg sync.WaitGroup
}

// Options options for new client
type Options struct {
	needSync      bool
	startBlockNum uint32
	handlers      []Handler
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

	// TODO: support multiple peers
	peer, err := NewPeer(peers[0], defaultOpts.startBlockNum, chainID)
	if err != nil {
		return nil, errors.Wrapf(err, "new peer error")
	}

	client := &Client{
		peer: peer,
		sync: &syncManager{
			IsSyncAll: defaultOpts.needSync,
			headBlock: defaultOpts.startBlockNum,
		},
	}
	client.RegisterHandler(NewMsgHandler(client.sync))

	for _, h := range defaultOpts.handlers {
		client.RegisterHandler(h)
	}

	err = client.Start(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "start client error")
	}

	return client, nil
}

// Close close client
func (c *Client) Close() error {
	if c.peer.connection == nil {
		return nil
	}
	return c.peer.connection.Close()
}

// SetReadTimeout set conn io readtimeout
func (c *Client) SetReadTimeout(readTimeout time.Duration) {
	c.readTimeout = readTimeout
}

// RegisterHandler reg handler to client
func (c *Client) RegisterHandler(handler Handler) {
	c.handlers = append(c.handlers, handler)
}

func (c *Client) peerLoop(peer *Peer) error {
	for {
		packet, err := peer.Read()
		if err != nil {
			return errors.Wrapf(err, "read message from %s", peer.Address)
		}

		envelope := newEnvelope(peer, packet)
		for _, handle := range c.handlers {
			handle.Handle(envelope)
		}
	}
}

func triggerHandshake(peer *Peer) error {
	return peer.SendHandshake(peer.handshakeInfo)
}

// Start start client process gorountinue
func (c *Client) Start(ctx context.Context) error {
	p2pLog.Info("Starting client")

	err := c.peer.Start(ctx)
	if err != nil {
		return errors.Wrap(err, "connect error")
	}

	// FIXME: there will be more peers
	c.wg.Add(1)
	go func(cc *Client) {
		defer cc.wg.Done()
		err := cc.peerLoop(cc.peer)

		// Tmp Implement
		if err != nil {
			p2pLog.Error("read error", zap.Error(err), zap.String("peer", cc.peer.Address))
			return
		}
	}(c)

	if c.peer.handshakeInfo != nil {
		err := triggerHandshake(c.peer)
		if err != nil {
			return errors.Wrap(err, "connect and start: trigger handshake")
		}
	}

	return nil
}

// Wait wait client closed
func (c *Client) Wait() {
	c.wg.Wait()
}
