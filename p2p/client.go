package p2p

import (
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Client struct {
	peer        *Peer
	handlers    []Handler
	readTimeout time.Duration
	sync        *syncManager

	wg sync.WaitGroup
}

func NewClient(peer *Peer, needSync bool) *Client {
	client := &Client{
		peer: peer,
		sync: &syncManager{
			IsSyncAll: needSync,
			headBlock: peer.handshakeInfo.HeadBlockNum,
		},
	}
	client.RegisterHandler(NewMsgHandler(client.sync))
	return client
}

func (c *Client) CloseConnection() error {
	if c.peer.connection == nil {
		return nil
	}
	return c.peer.connection.Close()
}

func (c *Client) SetReadTimeout(readTimeout time.Duration) {
	c.readTimeout = readTimeout
}

func (c *Client) RegisterHandler(handler Handler) {
	c.handlers = append(c.handlers, handler)
}

func (c *Client) read(peer *Peer) error {
	for {
		packet, err := peer.Read()
		if err != nil {
			return errors.Wrapf(err, "read message from %s", peer.Address)
		}

		envelope := NewEnvelope(peer, peer, packet)
		for _, handle := range c.handlers {
			handle.Handle(envelope)
		}
	}
}

func triggerHandshake(peer *Peer) error {
	return peer.SendHandshake(peer.handshakeInfo)
}

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

func (c *Client) Wait() {
	c.wg.Wait()
}
