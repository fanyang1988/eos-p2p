package main

import (
	"context"
	"flag"

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

	Logger.Info("P2P Client ", zap.String("peer", *peer), zap.String("chainid", *chainID))
	client, err := p2p.NewClient(
		context.Background(),
		*chainID,
		[]*p2p.PeerCfg{
			&p2p.PeerCfg{
				Name:    "eos-p2p",
				Address: *peer,
			},
		},
		p2p.WithNeedSync(1),
		p2p.WithHandler(p2p.StringLoggerHandler),
		p2p.WithHandler(p2p.NewMsgHandler(&MsgHandler{})),
	)

	if err != nil {
		Logger.Error("new client error", zap.Error(err))
		return
	}

	client.Wait()
}
