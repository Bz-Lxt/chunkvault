package engine_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Bz-Lxt/chunkvault/config"
	"github.com/Bz-Lxt/chunkvault/engine"
	"github.com/Bz-Lxt/chunkvault/wal"
)

func TestReadAllKeepsLastRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.wal")
	j, err := wal.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if err := j.Append(wal.Record{Op: wal.OpSetMeta, Name: "k", Payload: []byte{byte(i)}}.Seal()); err != nil {
			t.Fatal(err)
		}
	}
	_ = j.Close()
	got, err := wal.ReadAll(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d", len(got))
	}
}

func TestStatsCountsAllWalRecords(t *testing.T) {
	v, err := engine.Open(config.Config{Dir: t.TempDir(), ChunkSize: 32})
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	if _, err := v.Put(context.Background(), "r", bytes40()); err != nil {
		t.Fatal(err)
	}
	st, err := v.Stats(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if st["wal_records"] < 3 {
		t.Fatalf("wal_records=%d", st["wal_records"])
	}
}

func bytes40() []byte {
	return []byte("RRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRR")
}
