package wal_test

import (
	"path/filepath"
	"testing"

	"github.com/Bz-Lxt/chunkvault/wal"
)

func TestAppendPersistsRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.wal")
	j, err := wal.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := j.Append(wal.Record{Op: wal.OpSetMeta, Name: "k", Payload: []byte("v")}.Seal()); err != nil {
		t.Fatal(err)
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := wal.ReadAll(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("records=%d", len(got))
	}
}

func TestTamperedCRCRejected(t *testing.T) {
	rec := wal.Record{Op: wal.OpSetMeta, Name: "k", Payload: []byte("v")}.Seal()
	rec.CRC ^= 1
	if rec.Valid() {
		t.Fatal("tampered crc accepted")
	}
}
