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

var peer = flag.String("peer", "localhost:9001", "peer to connect to")
var chainID = flag.String("chain-id", "322ec54a9f13ad434efe9bc76ed6c7df13e2543d83235bc7bedebb4e23af1f2c", "net chainID to connect to")
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

	Logger.Info("P2P Client ", zap.String("peer", *peer), zap.String("chainid", *chainID))
	client, err := p2p.NewClient(
		ctx,
		*chainID,
		[]*p2p.PeerCfg{
			&p2p.PeerCfg{
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

	waitClose()

	cf()

	client.Wait()

	Logger.Info("p2p node stopped")
}
