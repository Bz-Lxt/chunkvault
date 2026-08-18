package engine_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/Bz-Lxt/chunkvault/config"
	"github.com/Bz-Lxt/chunkvault/engine"
)

func TestCheckpointKeepsBlob(t *testing.T) {
	v, err := engine.Open(config.Config{Dir: t.TempDir(), ChunkSize: 32})
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	body := []byte("keep-after-checkpoint")
	res, err := v.Put(context.Background(), "k", body)
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Checkpoint(context.Background()); err != nil {
		t.Fatal(err)
	}
	got, err := v.Get(context.Background(), res.Digest)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, body) {
		t.Fatal("lost")
	}
}
