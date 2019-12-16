package p2p

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"runtime"
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
}

// PeerCfg config for peer
type PeerCfg struct {
	Name    string `json:"name"`
	Address string `json:"addr"`
}

// GetName get peer name
func (p PeerCfg) GetName() string {
	if p.Name == "" {
		return fmt.Sprintf("peer-cli-%s", p.Address)
	}

	return p.Name
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

	res := &Peer{
		Address: cfg.Address,
		agent:   cfg.GetName(),
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
		p2pLog.Error("Connection Read Err", zap.String("address", p.Address), zap.Error(err))
		return nil, errors.Wrapf(err, "connection: read %s err", p.Address)
	}
	return packet, nil
}

func (p *Peer) setConnection(conn net.Conn) {
	p.connection = conn
	p.reader = bufio.NewReader(p.connection)
}

// Connect connect and start read go routine
func (p *Peer) Connect() error {
	nodeID := make([]byte, 32)
	_, err := rand.Read(nodeID)
	if err != nil {
		return errors.Wrap(err, "generating random node id")
	}

	p.NodeID = nodeID
	hexNodeID := hex.EncodeToString(p.NodeID)
	p.Name = fmt.Sprintf("Client Peer - %s", hexNodeID[0:8])

	address2log := zap.String("address", p.Address)

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

	p2pLog.Info("Dialing", address2log, zap.Duration("timeout", p.connectionTimeout))
	conn, err := net.DialTimeout("tcp", p.Address, p.connectionTimeout)
	if err != nil {
		if p.handshakeTimeout > 0 {
			p.cancelHandshakeTimeout <- true
		}
		return errors.Wrapf(err, "peer init: dial %s", p.Address)
	}
	p2pLog.Info("Connected to", address2log)
	p.setConnection(conn)

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

func (p *Peer) Write(bytes []byte) (int, error) {
	return p.connection.Write(bytes)
}

// WriteP2PMessage wrrite a p2p msg to peer
func (p *Peer) WriteP2PMessage(message Message) (err error) {
	packet := &Packet{
		Type:       message.GetType(),
		P2PMessage: message,
	}

	buff := bytes.NewBuffer(make([]byte, 0, 512))

	encoder := newEOSEncoder(buff)
	err = encoder.Encode(packet)
	if err != nil {
		return errors.Wrapf(err, "unable to encode message %s", message)
	}

	_, err = p.Write(buff.Bytes())
	if err != nil {
		return errors.Wrapf(err, "write msg to %s", p.Address)
	}

	return nil
}

// SendGoAway send go away message to peer
func (p *Peer) SendGoAway(reason GoAwayReason) error {
	p2pLog.Debug("SendGoAway", zap.String("reson", reason.String()))

	return errors.WithStack(p.WriteP2PMessage(&GoAwayMessage{
		Reason: reason,
		NodeID: p.NodeID,
	}))
}

// SendSyncRequest send a sync req
func (p *Peer) SendSyncRequest(startBlockNum uint32, endBlockNumber uint32) (err error) {
	p2pLog.Debug("SendSyncRequest",
		zap.String("peer", p.Address),
		zap.Uint32("start", startBlockNum),
		zap.Uint32("end", endBlockNumber))

	syncRequest := &SyncRequestMessage{
		StartBlock: startBlockNum,
		EndBlock:   endBlockNumber,
	}

	return errors.WithStack(p.WriteP2PMessage(syncRequest))
}

// SendRequest send req msg for p2p
func (p *Peer) SendRequest(startBlockNum uint32, endBlockNumber uint32) (err error) {
	p2pLog.Debug("SendRequest",
		zap.String("peer", p.Address),
		zap.Uint32("start", startBlockNum),
		zap.Uint32("end", endBlockNumber))

	request := &RequestMessage{
		ReqTrx: OrderedBlockIDs{
			Mode:    [4]byte{0, 0, 0, 0},
			Pending: startBlockNum,
		},
		ReqBlocks: OrderedBlockIDs{
			Mode:    [4]byte{0, 0, 0, 0},
			Pending: endBlockNumber,
		},
	}

	return errors.WithStack(p.WriteP2PMessage(request))
}

// SendNotice send notice msg for p2p
func (p *Peer) SendNotice(headBlockNum uint32, libNum uint32, mode byte) error {
	p2pLog.Debug("Send Notice",
		zap.String("peer", p.Address),
		zap.Uint32("head", headBlockNum),
		zap.Uint32("lib", libNum),
		zap.Uint8("type", mode))

	notice := &NoticeMessage{
		KnownTrx: OrderedBlockIDs{
			Mode:    [4]byte{mode, 0, 0, 0},
			Pending: headBlockNum,
		},
		KnownBlocks: OrderedBlockIDs{
			Mode:    [4]byte{mode, 0, 0, 0},
			Pending: libNum,
		},
	}
	return errors.WithStack(p.WriteP2PMessage(notice))
}

// SendTime send time sync msg to peer
func (p *Peer) SendTime() error {
	p2pLog.Debug("SendTime", zap.String("peer", p.Address))

	notice := &TimeMessage{}
	return errors.WithStack(p.WriteP2PMessage(notice))
}

// SendHandshake send handshake msg to peer
func (p *Peer) SendHandshake(info *HandshakeInfo) error {

	publicKey, err := newPublicKey("EOS1111111111111111111111111111111114T1Anm")
	if err != nil {
		return errors.Wrapf(err, "sending handshake to %s: create public key", p.Address)
	}

	p2pLog.Debug("SendHandshake", zap.String("peer", p.Address), zap.Object("info", info))

	tstamp := Tstamp{Time: info.HeadBlockTime}

	signature := Signature{
		Curve:   CurveK1,
		Content: make([]byte, 65, 65),
	}

	handshake := &HandshakeMessage{
		NetworkVersion:           1206,
		ChainID:                  info.ChainID,
		NodeID:                   p.NodeID,
		Key:                      publicKey,
		Time:                     tstamp,
		Token:                    make([]byte, 32, 32),
		Signature:                signature,
		P2PAddress:               p.Name,
		LastIrreversibleBlockNum: info.LastIrreversibleBlockNum,
		LastIrreversibleBlockID:  info.LastIrreversibleBlockID,
		HeadNum:                  info.HeadBlockNum,
		HeadID:                   info.HeadBlockID,
		OS:                       runtime.GOOS,
		Agent:                    p.agent,
		Generation:               int16(1),
	}

	err = p.WriteP2PMessage(handshake)
	if err != nil {
		err = errors.Wrapf(err, "sending handshake to %s", p.Address)
	}

	return nil
}
