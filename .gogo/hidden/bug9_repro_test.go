package engine_test

import (
	"context"
	"testing"

	"github.com/Bz-Lxt/chunkvault/config"
	"github.com/Bz-Lxt/chunkvault/engine"
)

func TestUnlinkThenGCDropsChunks(t *testing.T) {
	v, err := engine.Open(config.Config{Dir: t.TempDir(), ChunkSize: 32})
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	res, err := v.Put(context.Background(), "u", []byte("unique-only-once-payload"))
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Unlink(context.Background(), res.Digest); err != nil {
		t.Fatal(err)
	}
	if _, err := v.Collect(context.Background()); err != nil {
		t.Fatal(err)
	}
	st, err := v.Stats(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if st["chunks"] != 0 {
		t.Fatalf("chunks=%d", st["chunks"])
	}
}
