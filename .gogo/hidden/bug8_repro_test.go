package wal_test

import (
	"path/filepath"
	"testing"

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
