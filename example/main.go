package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/eosforce/eos-p2p/p2p"
	"github.com/eosforce/eos-p2p/store"
)

var peer = flag.String("peer", "", "peer to connect to")
var chainID = flag.String("chain-id", "76eab2b704733e933d0e4eb6cc24d260d9fbbe5d93d760392e97398f4e301448", "net chainID to connect to")
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

var logger *zap.Logger

func main() {
	flag.Parse()

	logger = zap.NewNop()
	if *showLog {
		logger, _ = zap.NewDevelopment()
		//types.EnableDetailLogs()
	}
	defer logger.Sync()

	ctx, cf := context.WithCancel(context.Background())

	peers := []string{
		"localhost:9001",
		"localhost:9002",
		"localhost:9003",
		"localhost:9004",
		"localhost:9005",
	}

	peersCfg := make([]*p2p.PeerCfg, 0, len(peers))

	if *peer != "" {
		peersCfg = append(peersCfg, &p2p.PeerCfg{
			Address: *peer,
		})
	} else {
		for _, p := range peers {
			peersCfg = append(peersCfg, &p2p.PeerCfg{
				Address: p,
			})
		}
	}

	storer, err := store.NewBBoltStorer(logger, *chainID, "./blocks.db", false)
	if err != nil {
		logger.Error("new storer error", zap.Error(err))
		return
	}

	client, err := p2p.NewClient(
		ctx,
		*chainID,
		peersCfg,
		p2p.WithLogger(logger),
		p2p.WithNeedSync(1),
		p2p.WithStorer(storer),
		p2p.WithHandler(p2p.NewMsgHandler("tmpHandler", &MsgHandler{})),
	)

	if err != nil {
		logger.Error("new client error", zap.Error(err))
		return
	}

	waitClose()

	cf()

	client.Wait()

	logger.Info("wait storer stop")

	storer.Close()
	storer.Wait()

	logger.Info("p2p node stopped")
}
