package p2p

import (
	"encoding/hex"

	eos "github.com/eosforce/goeosforce"
	"go.uber.org/zap"
)

// DecodeHex decodeString to hex
func DecodeHex(hexString string) (data []byte) {
	data, err := hex.DecodeString(hexString)
	if err != nil {
		logErr("decodeHexErr", err)
	}
	return data
}

// EnableP2PLogging enable p2p package to log by zap
func EnableP2PLogging() *zap.Logger {
	p2pLog = eos.NewLogger(false)
	return p2pLog
}
