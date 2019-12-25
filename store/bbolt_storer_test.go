package store

import (
	"testing"

	"go.uber.org/zap"
)

func TestStoreInit(t *testing.T) {
	l, _ := zap.NewDevelopment()
	defer l.Sync()
	s, err := NewBBoltStorer(l, "", "./testblocks.db", true)
	if err != nil {
		t.Fatalf("error by new %s", err.Error())
	}
	defer s.Close()
}
