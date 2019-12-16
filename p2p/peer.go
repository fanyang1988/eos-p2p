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
)

// Peer a p2p peer to other
type Peer struct {
	Address                string
	Name                   string
	agent                  string
	NodeID                 []byte
	connection             net.Conn
	reader                 io.Reader
	handshakeInfo          *HandshakeInfo
	connectionTimeout      time.Duration
	handshakeTimeout       time.Duration
	cancelHandshakeTimeout chan bool
	cli                    *Client
	wg                     sync.WaitGroup
}

// PeerCfg config for peer
type PeerCfg struct {
	Name    string `json:"name"`
	Address string `json:"addr"`
}

// MarshalLogObject calls the underlying function from zap.
func (p Peer) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", p.Name)
	enc.AddString("address", p.Address)
	enc.AddString("agent", p.agent)
	return enc.AddObject("handshakeInfo", p.handshakeInfo)
}

// HandshakeInfo handshake state for peer
type HandshakeInfo struct {
	ChainID                  Checksum256
	HeadBlockNum             uint32
	HeadBlockID              Checksum256
	HeadBlockTime            time.Time
	LastIrreversibleBlockNum uint32
	LastIrreversibleBlockID  Checksum256
}

func (h *HandshakeInfo) String() string {
	return fmt.Sprintf("Handshake Info: Head[%d], LastIrreversible[%d]",
		h.HeadBlockNum, h.LastIrreversibleBlockNum)
}

// MarshalLogObject calls the underlying function from zap.
func (h HandshakeInfo) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("chainID", h.ChainID.String())
	enc.AddUint32("headBlockNum", h.HeadBlockNum)
	enc.AddString("headBlockID", h.HeadBlockID.String())
	enc.AddTime("headBlockTime", h.HeadBlockTime)
	enc.AddUint32("lastIrreversibleBlockNum", h.LastIrreversibleBlockNum)
	enc.AddString("lastIrreversibleBlockID", h.LastIrreversibleBlockID.String())
	return nil
}

// NewPeer create a peer
func NewPeer(cfg *PeerCfg, headBlockNum uint32, chainID string) (*Peer, error) {
	cID, err := hex.DecodeString(chainID)
	if err != nil {
		return nil, errors.Wrapf(err, "decode chainID error")
	}

	nodeID := make([]byte, 32)
	_, err = rand.Read(nodeID)
	if err != nil {
		return nil, errors.Wrap(err, "generating random node id error")
	}

	name := cfg.Name

	if name == "" {
		hexNodeID := hex.EncodeToString(nodeID)
		name = fmt.Sprintf("ClientPeer-%s", hexNodeID[0:8])
	}

	res := &Peer{
		NodeID:  nodeID,
		Address: cfg.Address,
		agent:   name,
		Name:    name,
		handshakeInfo: &HandshakeInfo{
			ChainID:      cID,
			HeadBlockNum: headBlockNum,
		},
		cancelHandshakeTimeout: make(chan bool),
	}

	return res, nil
}

// SetHandshakeTimeout set send handshake timeout
func (p *Peer) SetHandshakeTimeout(timeout time.Duration) {
	p.handshakeTimeout = timeout
}

// SetConnectionTimeout for net DialTimeout
func (p *Peer) SetConnectionTimeout(timeout time.Duration) {
	p.connectionTimeout = timeout
}

func (p *Peer) Read() (*Packet, error) {
	packet, err := readEOSPacket(p.reader)
	if p.handshakeTimeout > 0 {
		p.cancelHandshakeTimeout <- true
	}
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
func (p *Peer) Start(ctx context.Context, client *Client) error {
	address2log := zap.String("address", p.Address)

	p.cli = client

	p2pLog.Info("Dialing", address2log, zap.Duration("timeout", p.connectionTimeout))
	err := p.connect()
	if err != nil {
		if p.handshakeTimeout > 0 {
			p.cancelHandshakeTimeout <- true
		}
		return err
	}

	if p.handshakeInfo != nil {
		err := p.SendHandshake(p.handshakeInfo)
		if err != nil {
			return errors.Wrap(err, "connect and start: trigger handshake")
		}
	}

	go func() {
		p.wg.Add(1)
		p.readLoop()
	}()

	if p.handshakeTimeout > 0 {
		go func(p *Peer) {
			select {
			case <-time.After(p.handshakeTimeout):
				p2pLog.Warn("handshake took too long", address2log)
				p.Close(goAwayNoReason)
			case <-p.cancelHandshakeTimeout:
				p2pLog.Warn("cancelHandshakeTimeout canceled", address2log)
			}
		}(p)
	}

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
			p2pLog.Error("peer readLoop panic", zap.String("addr", p.Address))
			p.cli.packetChan <- newEnvelopMsgWithError(p, errors.Errorf("panic by %v", r))
		}
	}()

	for {
		packet, err := p.Read()

		if err != nil {
			//p2pLog.Warn("peer readLoop return by read error", zap.String("addr", p.Address), zap.Error(err))
			p.cli.packetChan <- newEnvelopMsgWithError(p, errors.Wrapf(err, "read message from %s", p.Address))
			return
		}

		p.cli.packetChan <- newEnvelopMsg(p, packet)
	}
}
