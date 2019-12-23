package store

import "testing"

import "go.uber.org/zap"

import "github.com/fanyang1988/eos-p2p/types"

func TestStoreInit(t *testing.T) {
	l, _ := zap.NewDevelopment()
	defer l.Sync()
	s, err := NewBBoltStorer(l, types.Checksum256([]byte("")), "./testblocks.db")
	if err != nil {
		t.Fatalf("error by new %s", err.Error())
	}
	defer s.Close()

	if err := s.initState(); err != nil {
		t.Fatalf("error by init %s", err.Error())
	}
}
