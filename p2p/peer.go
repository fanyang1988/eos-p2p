package p2p

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/fanyang1988/eos-p2p/types"
)

// Peer a p2p peer to other
type Peer struct {
	Address           string
	Name              string
	agent             string
	NodeID            []byte
	connection        net.Conn
	reader            io.Reader
	connectionTimeout time.Duration
	cli               *Client
	wg                *sync.WaitGroup
}

// PeerCfg config for peer
type PeerCfg struct {
	Name    string `json:"name"`
	Address string `json:"addr"`
}

// MarshalLogObject calls the underlying function from zap.
func (p *Peer) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", p.Name)
	enc.AddString("address", p.Address)
	enc.AddString("agent", p.agent)
	return nil
}

// NewPeer create a peer
func NewPeer(cfg *PeerCfg, cli *Client, headBlockNum uint32, chainID Checksum256) (*Peer, error) {
	nodeID := make([]byte, 32)
	_, err := rand.Read(nodeID)
	if err != nil {
		return nil, errors.Wrap(err, "generating random node id error")
	}

	name := cfg.Name

	if name == "" {
		hexNodeID := hex.EncodeToString(nodeID)
		name = fmt.Sprintf("ClientPeer-%s", hexNodeID[0:8])
	}

	res := &Peer{
		NodeID:            nodeID,
		Address:           cfg.Address,
		agent:             name,
		Name:              name,
		connectionTimeout: 5 * time.Second,
		wg:                &sync.WaitGroup{},
		cli:               cli,
	}

	return res, nil
}

// SetConnectionTimeout for net DialTimeout
func (p *Peer) SetConnectionTimeout(timeout time.Duration) {
	p.connectionTimeout = timeout
}

func (p *Peer) Read() (*Packet, error) {
	packet, err := types.ReadChainPacket(p.reader, p.connection)
	if err != nil {
		return nil, errors.Wrapf(err, "connection: read %s err", p.Address)
	}

	return packet, nil
}

func (p *Peer) connect() error {
	conn, err := net.DialTimeout("tcp", p.Address, p.connectionTimeout)
	if err != nil {
		return errors.Wrapf(err, "peer connect error %s", p.Address)
	}

	p.connection = conn
	p.reader = bufio.NewReader(p.connection)

	return nil
}

// Start connect and start read go routine
func (p *Peer) Start(ctx context.Context) error {
	address2log := zap.String("address", p.Address)

	p.cli.logger.Info("Dialing", address2log, zap.Duration("timeout", p.connectionTimeout))
	err := p.connect()
	if err != nil {
		return err
	}

	go func() {
		p.wg.Add(1)
		p.readLoop()
	}()

	return nil
}

// Close send GoAway message then close connection
func (p *Peer) Close(reason GoAwayReason) error {
	p.SendGoAway(reason)
	return p.ClosePeer()
}

// ClosePeer close peer connect
func (p *Peer) ClosePeer() error {
	if p.connection != nil {
		return p.connection.Close()
	}

	return nil
}

// Wait wait peer stop ok
func (p *Peer) Wait() {
	p.wg.Wait()
}

func (p *Peer) readLoop() {
	defer func() {
		p.wg.Done()
		if r := recover(); r != nil {
			p.cli.logger.Error("peer readLoop panic", zap.String("addr", p.Address))
			p.cli.packetChan <- newEnvelopMsgWithError(p, errors.Errorf("panic by %v", r))
		}
	}()

	for {
		packet, err := p.Read()

		if err != nil {
			//p.cli.logger.Warn("peer readLoop return by read error", zap.String("addr", p.Address), zap.Error(err))
			p.cli.packetChan <- newEnvelopMsgWithError(p, errors.Wrapf(err, "read message from %s", p.Address))
			p.cli.logger.Debug("peer readloop exit", zap.String("address", p.Address))
			return
		}

		p.cli.packetChan <- newEnvelopMsg(p, packet)
	}
}
