package engine_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/Bz-Lxt/chunkvault/config"
	"github.com/Bz-Lxt/chunkvault/engine"
)

// TestCanceledMidPutSkipsCommit cancels the context partway through the slow
// chunk loop, after some chunk records have already been appended to the
// journal. The post-loop ctx.Err guard must abort before the link/commit step,
// so no blob lands in the list.
func TestCanceledMidPutSkipsCommit(t *testing.T) {
	v, err := engine.Open(config.Config{Dir: t.TempDir(), ChunkSize: 32, SlowChunkMS: 25})
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()

	// 8 chunks of 32 bytes -> the slow loop runs ~200ms; cancel at 10ms lands
	// squarely mid-flight, well before any commit could happen.
	body := bytes.Repeat([]byte("0123456789ABCDEF0123456789ABCDEF"), 8)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	if _, err := v.Put(ctx, "ghost", body); err == nil {
		t.Fatal("expected canceled put to return an error")
	}
	ids, err := v.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Fatalf("canceled put committed %d blob(s) into the list", len(ids))
	}
}
