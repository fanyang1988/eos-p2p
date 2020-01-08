package types

import (
	"time"

	eos "github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
)

func newBlockForTest(num uint32) *SignedBlock {
	signature, _ := ecc.NewSignature("SIG_K1_K1eEGkH78D3GzkGnoFdJ4mQ12ihxryfRErHzvBcfzy8kiSWepjB4tPwivavJoZeX47gKfGDYsk6LtyCrth3cbQF4Az8aN8")
	res := &SignedBlock{
		SignedBlockHeader: eos.SignedBlockHeader{
			BlockHeader: eos.BlockHeader{
				Producer:         AccountName("biosbpa"),
				Confirmed:        20,
				Previous:         MustNewChecksum256("0000000409244caf255922a0d1354d50aec2bdb91f6b2269c0a98d31cdef930f"),
				TransactionMRoot: MustNewChecksum256("ecf68a2c1fb24d2014b90e70b89334bc1d951eb477d56476a26c561f783fdfa3"),
				ActionMRoot:      MustNewChecksum256("86ecac45b8f41a9918a1df22c6ec088ccd5d7e99cb892d45c6e7cfa940375cf2"),
				ScheduleVersion:  1,
				HeaderExtensions: make([]*eos.Extension, 0, 8),
			},
			ProducerSignature: signature,
		},
		Transactions:    make([]eos.TransactionReceipt, 0, 8),
		BlockExtensions: make([]*eos.Extension, 0, 8),
	}

	res.Timestamp = eos.BlockTimestamp{time.Unix(time.Now().Unix(), 0).UTC()}

	return res
}
