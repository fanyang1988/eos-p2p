package types

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"io"
	"net"

	eos "github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/pkg/errors"
)

// NewChecksum256 new checksum256 from string
func NewChecksum256(hexStr string) (Checksum256, error) {
	res, err := hex.DecodeString(hexStr)
	if err != nil {
		return Checksum256([]byte{}), err
	}

	return Checksum256(res), nil
}

// MustNewChecksum256 new checksum256 from string, if err panic
func MustNewChecksum256(hexStr string) Checksum256 {
	res, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(errors.Wrapf(err, "decode str %s", hexStr))
	}

	return Checksum256(res)
}

// NewPublicKey create public key
func NewPublicKey(pubKey string) (out PublicKey, err error) {
	return ecc.NewPublicKey(pubKey)
}

// ReadChainPacket read chain packet for p2p from a conn
func ReadChainPacket(r io.Reader, conn net.Conn) (packet *Packet, err error) {
	return readPacket(r, conn)
}

// NewChainEncoder create chain encoder for datas
func NewChainEncoder(w io.Writer) *eos.Encoder {
	return eos.NewEncoder(w)
}

func readPacket(r io.Reader, conn net.Conn) (*Packet, error) {
	data := make([]byte, 0)

	lengthBytes := make([]byte, 4, 4)

	if _, err := io.ReadFull(r, lengthBytes); err != nil {
		return nil, errors.Wrapf(err, "readfull length")
	}

	data = append(data, lengthBytes...)

	size := binary.LittleEndian.Uint32(lengthBytes)

	if size > 16*1024*1024 {
		return nil, errors.Errorf("packet is too large %d", size)
	}

	payloadBytes := make([]byte, size, size)

	count, err := io.ReadFull(r, payloadBytes)

	if err != nil {
		return nil, errors.Errorf("read full data error")
	}

	if count != int(size) {
		return nil, errors.Errorf("read full not full read[%d] expected[%d]", count, size)
	}

	data = append(data, payloadBytes...)

	packet := &Packet{}
	decoder := eos.NewDecoder(data)
	decoder.DecodeActions(false)
	err = decoder.Decode(packet)
	if err != nil {
		return nil, errors.Wrapf(err, "Failing decode data %s", hex.EncodeToString(data))
	}
	packet.Raw = data
	return packet, nil
}

// EncodeToEOS encode as the eos binary format
func EncodeToEOS(obj interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := eos.NewEncoder(&buffer)

	if err := encoder.Encode(obj); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// IsChecksumEq is two Checksum256 is same
func IsChecksumEq(l Checksum256, r Checksum256) bool {
	if len(l) != len(r) {
		return false
	}

	for idx, c := range l {
		if r[idx] != c {
			return false
		}
	}

	return true
}

// EnableDetailLogs enable logs for chain package
func EnableDetailLogs() {
	eos.EnableEncoderLogging()
	eos.EnableDecoderLogging()
	eos.EnableABIEncoderLogging()
	eos.EnableABIDecoderLogging()
}
