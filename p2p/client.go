package p2p

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type peerStatusTyp uint8

const (
	peerStatNormal = peerStatusTyp(iota)
	peerStatInited
	peerStatConnecting
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
	ps          []peerStatus
	handlers    []Handler
	readTimeout time.Duration
	sync        *syncManager

	packetChan chan envelopMsg
	peerChan   chan peerMsg

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

	ps := make([]peerStatus, 0, 64)

	for _, p := range peers {
		peer, err := NewPeer(p, defaultOpts.startBlockNum, chainID)
		if err != nil {
			return nil, errors.Wrapf(err, "new peer error")
		}

		ps = append(ps, peerStatus{
			peer:   peer,
			status: peerStatInited,
			cfg:    p,
		})
	}

	client := &Client{
		ps: ps,
		sync: &syncManager{
			IsSyncAll: defaultOpts.needSync,
			headBlock: defaultOpts.startBlockNum,
		},
		packetChan: make(chan envelopMsg, 256),
		peerChan:   make(chan peerMsg, 8),
	}
	client.RegisterHandler(NewMsgHandler(client.sync))

	for _, h := range defaultOpts.handlers {
		client.RegisterHandler(h)
	}

	err := client.Start(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "start client error")
	}

	return client, nil
}

// SetReadTimeout set conn io readtimeout
func (c *Client) SetReadTimeout(readTimeout time.Duration) {
	c.readTimeout = readTimeout
}

// RegisterHandler reg handler to client
func (c *Client) RegisterHandler(handler Handler) {
	c.handlers = append(c.handlers, handler)
}

func (c *Client) closeAllPeer() {
	for _, p := range c.ps {
		p.peer.ClosePeer()
	}

	for _, p := range c.ps {
		p.peer.Wait()
	}
}

func (c *Client) peerMngLoop(ctx context.Context) {
	for {
		select {
		case p := <-c.peerChan:
			// if already stop loop so close directly
			select {
			case <-ctx.Done():
				p2pLog.Info("close peer chan mng")
				return
			default:
			}

			if p.peer != nil {
				switch p.msgTyp {
				case peerMsgNewPeer:
				case peerMsgDelPeer:
				case peerMsgErrPeer:
					if p.err != nil {
						p2pLog.Info("reconnect peer", zap.String("addr", p.peer.Address))
						c.StartPeer(ctx, p.peer)
						time.Sleep(3 * time.Second)
					}
				}
			}

		case <-ctx.Done():
			// no need wait all msg in chan processed
			p2pLog.Info("close peer chan mng")
			return
		}
	}
}

func (c *Client) peerLoop(ctx context.Context) {
	isStopped := false
	for {
		select {
		case r, ok := <-c.packetChan:
			if !ok {
				p2pLog.Info("client peerLoop stop")
				return
			}

			if r.err != nil {
				if errors.Cause(r.err) != io.EOF {
					p2pLog.Info("client res error", zap.Error(r.err))
				} else {
					p2pLog.Info("conn closed")
					if !isStopped {
						c.peerChan <- peerMsg{
							err:    r.err,
							peer:   r.Sender,
							msgTyp: peerMsgErrPeer,
						}
					}
				}
				continue
			}

			envelope := newEnvelope(r.Sender, r.Packet)
			for _, handle := range c.handlers {
				handle.Handle(envelope)
			}

		case <-ctx.Done():
			if !isStopped {
				isStopped = true
				p2pLog.Info("close p2p client")
				c.closeAllPeer()
				p2pLog.Info("all peer is closed, to close client peerLoop")
				close(c.packetChan)
			}
		}
	}
}

// Start start client process gorountinue
func (c *Client) Start(ctx context.Context) error {
	p2pLog.Info("Starting client")

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

	for _, p := range c.ps {
		c.StartPeer(ctx, p.peer)
	}
	return nil
}

// Wait wait client closed
func (c *Client) Wait() {
	c.wg.Wait()
}

// StartPeer start a peer r/w
func (c *Client) StartPeer(ctx context.Context, p *Peer) {
	p2pLog.Info("Start Connect Peer", zap.String("peer", p.Address))
	err := p.Start(ctx, c)
	if err != nil {
		c.peerChan <- peerMsg{
			err:    errors.Wrap(err, "connect error"),
			peer:   p,
			msgTyp: peerMsgErrPeer,
		}
	}

	return
}
