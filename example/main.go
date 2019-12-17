package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/fanyang1988/eos-p2p/p2p"
)

var peer = flag.String("peer", "localhost:9000", "peer to connect to")
var chainID = flag.String("chain-id", "1b85dedb0a11a73443f1baa1667499b2329283d516b368698a7f2e16bc3a3232", "net chainID to connect to")
var showLog = flag.Bool("v", true, "show detail log")

// waitClose wait for term signal, then stop the server
func waitClose() {
	stopSignalChan := make(chan os.Signal, 1)
	signal.Notify(stopSignalChan,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGQUIT,
		syscall.SIGUSR1)
	<-stopSignalChan
}

func main() {
	flag.Parse()

	if *showLog {
		EnableLogging()
		p2p.SetLogger(Logger)
	}
	defer Logger.Sync()

	ctx, cf := context.WithCancel(context.Background())

	peers := []string{
		"localhost:9000",
		"localhost:9001",
		"localhost:9002",
		"localhost:9003",
		"localhost:9004",
		"localhost:9005",
	}

	peersCfg := make([]*p2p.PeerCfg, 0, len(peers))
	for _, p := range peers {
		peersCfg = append(peersCfg, &p2p.PeerCfg{
			Address: p,
		})
	}

	Logger.Info("P2P Client ", zap.String("peer", *peer), zap.String("chainid", *chainID))
	client, err := p2p.NewClient(
		ctx,
		*chainID,
		peersCfg,
		//p2p.WithNeedSync(1),
		p2p.WithHandler(p2p.StringLoggerHandler),
		p2p.WithHandler(p2p.NewMsgHandler(&MsgHandler{})),
	)

	if err != nil {
		Logger.Error("new client error", zap.Error(err))
		return
	}

	waitClose()

	cf()

	client.Wait()

	Logger.Info("p2p node stopped")
}
