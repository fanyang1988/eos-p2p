package p2p

import (
	"time"

	"github.com/eoscanada/eos-go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Client struct {
	peer        *Peer
	handlers    []Handler
	readTimeout time.Duration
	sync        *syncManager
}

func NewClient(peer *Peer, needSync bool) *Client {
	client := &Client{
		peer: peer,
	}
	if needSync {
		client.sync = &syncManager{
			headBlock: peer.handshakeInfo.HeadBlockNum,
		}
		client.RegisterHandler(NewMsgHandler(client.sync))
	}
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

func (c *Client) read(peer *Peer, errChannel chan error) {
	for {
		packet, err := peer.Read()
		if err != nil {
			errChannel <- errors.Wrapf(err, "read message from %s", peer.Address)
			break
		}

		envelope := NewEnvelope(peer, peer, packet)
		for _, handle := range c.handlers {
			handle.Handle(envelope)
		}

		switch m := packet.P2PMessage.(type) {
		case *eos.GoAwayMessage:
			errChannel <- errors.Wrapf(err, "GoAwayMessage reason %s", m.Reason)

		case *eos.HandshakeMessage:
			if c.sync == nil {
				m.NodeID = peer.NodeID
				m.P2PAddress = peer.Name
				err = peer.WriteP2PMessage(m)
				if err != nil {
					errChannel <- errors.Wrap(err, "HandshakeMessage")
					break
				}
				p2pLog.Debug("Handshake resent", zap.String("other", m.P2PAddress))

			} else {

				c.sync.originHeadBlock = m.HeadNum
				err = c.sync.sendSyncRequest(peer)
				if err != nil {
					errChannel <- errors.Wrap(err, "handshake: sending sync request")
				}
				c.sync.IsCatchingUp = true
			}
		case *eos.NoticeMessage:
			if c.sync != nil {
				pendingNum := m.KnownBlocks.Pending
				if pendingNum > 0 {
					c.sync.originHeadBlock = pendingNum
					err = c.sync.sendSyncRequest(peer)
					if err != nil {
						errChannel <- errors.Wrap(err, "noticeMessage: sending sync request")
					}
				}
			}
		case *eos.SignedBlock:

			if c.sync != nil {

				blockNum := m.BlockNumber()
				c.sync.headBlock = blockNum
				if c.sync.requestedEndBlock == blockNum {

					if c.sync.originHeadBlock <= blockNum {
						p2pLog.Debug("In sync with last handshake")
						blockID, err := m.BlockID()
						if err != nil {
							errChannel <- errors.Wrap(err, "getting block id")
						}
						peer.handshakeInfo.HeadBlockNum = blockNum
						peer.handshakeInfo.HeadBlockID = blockID
						peer.handshakeInfo.HeadBlockTime = m.SignedBlockHeader.Timestamp.Time
						err = peer.SendHandshake(peer.handshakeInfo)
						if err != nil {
							errChannel <- errors.Wrap(err, "send handshake")
						}
						p2pLog.Debug("Send new handshake",
							zap.Object("handshakeInfo", peer.handshakeInfo))
					} else {
						err = c.sync.sendSyncRequest(peer)
						if err != nil {
							errChannel <- errors.Wrap(err, "signed block: sending sync request")
						}
					}
				}
			}
		}
	}
}

func triggerHandshake(peer *Peer) error {
	return peer.SendHandshake(peer.handshakeInfo)
}

func (c *Client) Start() error {
	p2pLog.Info("Starting client")

	errorChannel := make(chan error, 1)
	readyChannel := c.peer.Connect(errorChannel)

	for {
		select {
		case <-readyChannel:
			go c.read(c.peer, errorChannel)
			if c.peer.handshakeInfo != nil {

				err := triggerHandshake(c.peer)
				if err != nil {
					return errors.Wrap(err, "connect and start: trigger handshake")
				}
			}
		case err := <-errorChannel:
			return errors.Wrap(err, "start client")
		}
	}
}
