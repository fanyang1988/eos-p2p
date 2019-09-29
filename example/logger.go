package main

import (
	"github.com/eoscanada/eos-go"
	"go.uber.org/zap"
)

// Logger default use nil zap logger
var Logger = zap.NewNop()

// EnableLogging enable p2p package to log by zap
func EnableLogging() {
	Logger = eos.NewLogger(false)
}
