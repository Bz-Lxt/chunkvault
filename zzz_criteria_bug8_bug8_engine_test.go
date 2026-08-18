package engine_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Bz-Lxt/chunkvault/config"
	"github.com/Bz-Lxt/chunkvault/engine"
)

func TestReopenReplaysTail(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{Dir: dir, ChunkSize: 32}
	v, err := engine.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	body := bytes.Repeat([]byte("R"), 40)
	res, err := v.Put(context.Background(), "r", body)
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Close(); err != nil {
		t.Fatal(err)
	}
	_ = os.Remove(filepath.Join(dir, "vault.sqlite"))
	_ = os.Remove(filepath.Join(dir, "vault.sqlite-wal"))
	_ = os.Remove(filepath.Join(dir, "vault.sqlite-shm"))
	v2, err := engine.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer v2.Close()
	got, err := v2.Get(context.Background(), res.Digest)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, body) {
		t.Fatal("lost tail")
	}
}
