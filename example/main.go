package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"

	"github.com/fanyang1988/eos-p2p/p2p"
)

var peer = flag.String("peer", "localhost:9001", "peer to connect to")
var chainID = flag.String("chain-id", "1c6ae7719a2a3b4ecb19584a30ff510ba1b6ded86e1fd8b8fc22f1179c622a32", "net chainID to connect to")
var showLog = flag.Bool("v", true, "show detail log")

func main() {
	flag.Parse()

	if *showLog {
		p2p.EnableP2PLogging()
	}
	defer p2p.SyncLogger()

	cID, err := hex.DecodeString(*chainID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("P2P Client ", *peer, " With Chain ID :", *chainID)
	client := p2p.NewClient(
		p2p.NewOutgoingPeer(*peer, "eos-proxy", &p2p.HandshakeInfo{
			ChainID:      cID,
			HeadBlockNum: 1,
		}),
		true,
	)

	client.RegisterHandler(p2p.StringLoggerHandler)
	client.RegisterHandler(p2p.NewMsgHandler(&MsgHandler{}))
	client.Start()
}