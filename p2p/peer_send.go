package p2p

import (
	"bytes"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/eosforce/eos-p2p/types"
)

// WriteP2PMessage write a p2p msg to peer
func (p *Peer) WriteP2PMessage(message Message) (err error) {
	packet := &Packet{
		Type:       message.GetType(),
		P2PMessage: message,
	}

	buff := bytes.NewBuffer(make([]byte, 0, 512))

	encoder := types.NewChainEncoder(buff)
	err = encoder.Encode(packet)
	if err != nil {
		return errors.Wrapf(err, "unable to encode message %s", message)
	}

	_, err = p.connection.Write(buff.Bytes())
	if err != nil {
		return errors.Wrapf(err, "write msg to %s", p.Address)
	}

	return nil
}

// SendGoAway send go away message to peer
func (p *Peer) SendGoAway(reason GoAwayReason) error {
	p.cli.logger.Debug("SendGoAway", zap.String("reason", reason.String()))

	return errors.WithStack(p.WriteP2PMessage(&GoAwayMessage{
		Reason: reason,
		NodeID: p.NodeID,
	}))
}

// SendSyncRequest send a sync req
func (p *Peer) SendSyncRequest(startBlockNum uint32, endBlockNumber uint32) (err error) {
	//p.cli.logger.Debug("SendSyncRequest",
	//	zap.String("peer", p.Address),
	//	zap.Uint32("start", startBlockNum),
	//	zap.Uint32("end", endBlockNumber))

	syncRequest := &SyncRequestMessage{
		StartBlock: startBlockNum,
		EndBlock:   endBlockNumber,
	}

	return errors.WithStack(p.WriteP2PMessage(syncRequest))
}

// SendRequest send req msg for p2p
func (p *Peer) SendRequest(startBlockNum uint32, endBlockNumber uint32) (err error) {
	p.cli.logger.Debug("SendRequest",
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
	p.cli.logger.Debug("Send Notice",
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

// SendNoticeHeadCatchup send notice msg for p2p
func (p *Peer) SendNoticeHeadCatchup(msg *NoticeMessage) error {
	p.cli.logger.Debug("SendNoticeHeadCatchup",
		zap.String("peer", p.Address),
		zap.String("trx", msg.KnownTrx.String()),
		zap.String("blk", msg.KnownBlocks.String()))

	notice := &NoticeMessage{
		KnownTrx: OrderedBlockIDs{
			Mode:    [4]byte{0, 0, 0, 0},
			Pending: msg.KnownTrx.Pending,
			IDs:     msg.KnownTrx.IDs,
		},
		KnownBlocks: OrderedBlockIDs{
			Mode:    [4]byte{1, 0, 0, 0},
			Pending: msg.KnownBlocks.Pending,
			IDs:     msg.KnownBlocks.IDs,
		},
	}
	return errors.WithStack(p.WriteP2PMessage(notice))
}

// SendTime send time sync msg to peer
func (p *Peer) SendTime(recv *TimeMessage) error {
	p.cli.logger.Debug("SendTime", zap.String("peer", p.Address))

	notice := &TimeMessage{}

	if recv != nil {
		notice.Origin = recv.Transmit
		notice.Receive = recv.Destination
		notice.Transmit = Tstamp{
			Time: time.Now(),
		}
	}

	return errors.WithStack(p.WriteP2PMessage(notice))
}

// SendHandshake send handshake msg to peer
func (p *Peer) SendHandshake(info *HandshakeInfo) error {

	// TODO: support peer key
	publicKey, err := types.NewPublicKey("EOS1111111111111111111111111111111114T1Anm")
	if err != nil {
		return errors.Wrapf(err, "sending handshake to %s: create public key", p.Address)
	}

	p.cli.logger.Debug("SendHandshake", zap.String("peer", p.Address), zap.Object("info", info))

	tstamp := Tstamp{Time: info.HeadBlockTime}

	signature := Signature{
		Curve:   CurveK1,
		Content: make([]byte, 65, 65),
	}

	p.sendHandshakeCount++

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
		OS:                       "osx", //runtime.GOOS,
		Agent:                    p.agent,
		Generation:               p.sendHandshakeCount,
	}

	p.cli.logger.Debug("info", zap.String("Name", handshake.String()))

	err = p.WriteP2PMessage(handshake)
	if err != nil {
		err = errors.Wrapf(err, "sending handshake to %s", p.Address)
	}

	return nil
}
