package p2p

import (
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// TODO: use mutiply p2p peers

// Client a p2p Client for eos chain
type Client struct {
	peer        *Peer
	handlers    []Handler
	readTimeout time.Duration
	sync        *syncManager

	wg sync.WaitGroup
}

// NewClient create a p2p client
func NewClient(peer *Peer, needSync bool) *Client {
	client := &Client{
		peer: peer,
		sync: &syncManager{
			IsSyncAll: needSync,
			headBlock: peer.handshakeInfo.HeadBlockNum, // TODO use local store for blocks had got
		},
	}
	client.RegisterHandler(NewMsgHandler(client.sync))
	return client
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

func (c *Client) read(peer *Peer) error {
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
func (c *Client) Start() error {
	p2pLog.Info("Starting client")

	err := c.peer.Connect()
	if err != nil {
		return errors.Wrap(err, "connect error")
	}

	c.wg.Add(1)

	// FIXME: there will be more peers
	go func(cc *Client) {
		defer cc.wg.Done()
		err := cc.read(cc.peer)

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
