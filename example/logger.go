package main

import (
	"go.uber.org/zap"

	"github.com/fanyang1988/eos-p2p/p2p"
)

// Logger default use nil zap logger
var Logger = zap.NewNop()

// EnableLogging enable p2p package to log by zap
func EnableLogging() {
	Logger = p2p.EnableP2PLogging()
}
