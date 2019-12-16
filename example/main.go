package main

import (
	"encoding/hex"
	"flag"
	"log"

	"go.uber.org/zap"

	"github.com/fanyang1988/eos-p2p/p2p"
)

var peer = flag.String("peer", "localhost:9001", "peer to connect to")
var chainID = flag.String("chain-id", "322ec54a9f13ad434efe9bc76ed6c7df13e2543d83235bc7bedebb4e23af1f2c", "net chainID to connect to")
var showLog = flag.Bool("v", true, "show detail log")

func main() {
	flag.Parse()

	if *showLog {
		EnableLogging()
		p2p.SetLogger(Logger)
	}
	defer Logger.Sync()

	cID, err := hex.DecodeString(*chainID)
	if err != nil {
		log.Fatal(err)
	}

	Logger.Info("P2P Client ", zap.String("peer", *peer), zap.String("chainid", *chainID))
	client := p2p.NewClient(
		p2p.NewPeer(*peer, "eos-proxy", &p2p.HandshakeInfo{
			ChainID:      cID,
			HeadBlockNum: 1,
		}),
		true,
	)

	client.RegisterHandler(p2p.StringLoggerHandler)
	client.RegisterHandler(p2p.NewMsgHandler(&MsgHandler{}))
	client.Start()

	client.Wait()
}
