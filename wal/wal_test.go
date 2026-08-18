package wal_test

import (
	"testing"

	"github.com/Bz-Lxt/chunkvault/wal"
)

func TestRecordCodec(t *testing.T) {
	rec := wal.Record{Op: wal.OpSetMeta, Name: "k", Payload: []byte("v")}.Seal()
	raw := wal.Marshal(rec)
	got, n, err := wal.Unmarshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(raw) || got.Name != "k" || string(got.Payload) != "v" {
		t.Fatalf("codec %+v n=%d", got, n)
	}
	if !got.Valid() {
		t.Fatal("crc")
	}
}
