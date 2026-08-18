package engine_test

import (
	"context"
	"testing"
	"time"

	"github.com/Bz-Lxt/chunkvault/config"
	"github.com/Bz-Lxt/chunkvault/engine"
)

func TestPreviewCancelDoesNotBlockNext(t *testing.T) {
	v, err := engine.Open(config.Config{Dir: t.TempDir(), ChunkSize: 32})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _ = v.Preview(ctx)
	done := make(chan error, 1)
	go func() {
		_, err := v.Preview(context.Background())
		done <- err
	}()
	select {
	case err := <-done:
		_ = v.Close()
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(500 * time.Millisecond):
		// 漏锁时 Close 会再抢同一把锁，不能 defer Close。
		t.Fatal("blocked")
	}
}

func TestPreviewListsZeroRef(t *testing.T) {
	v, err := engine.Open(config.Config{Dir: t.TempDir(), ChunkSize: 32})
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	res, err := v.Put(context.Background(), "z", []byte("zero-ref-after-unlink"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.Collect(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := v.Unlink(context.Background(), res.Digest); err != nil {
		t.Fatal(err)
	}
	plan, err := v.Preview(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Victims) == 0 {
		t.Fatal("expected zero-ref victims")
	}
}
