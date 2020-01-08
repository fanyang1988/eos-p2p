package types

import (
	"encoding/json"
	"testing"
	"time"

	eos "github.com/eoscanada/eos-go"
)

// TestBlockJSON test block to json is not change data
func TestBlockJSON(t *testing.T) {
	blk := newBlockForTest(1)
	blkID, _ := blk.BlockID()
	jsonStr, _ := json.Marshal(*blk)
	blkCopy := &SignedBlock{}
	err := json.Unmarshal(jsonStr, blkCopy)
	if err != nil {
		t.Errorf("unmarshal error: %v", err.Error())
		t.Failed()
	}
	blkID2, _ := blkCopy.BlockID()

	milliseconds1 := blk.Timestamp.UnixNano() / time.Millisecond.Nanoseconds()
	slot1 := (milliseconds1 - 946684800000) / 3000

	milliseconds2 := blkCopy.Timestamp.UnixNano() / time.Millisecond.Nanoseconds()
	slot2 := (milliseconds2 - 946684800000) / 3000

	bb1, _ := eos.MarshalBinary(blk)
	bb2, _ := eos.MarshalBinary(blkCopy)

	t.Logf("b1 %v\n", *blk)
	t.Logf("b2 %v\n", *blkCopy)

	t.Logf("bb1 %x\n", bb1)
	t.Logf("bb2 %x\n", bb2)

	t.Logf("bb1 %d %d\n", milliseconds1, slot1)
	t.Logf("bb2 %d %d\n", milliseconds2, slot2)

	if !IsChecksumEq(blkID, blkID2) {
		t.Errorf("id not equal \n%s\n%s\n", blkID.String(), blkID2.String())
		t.Failed()
	}
}
