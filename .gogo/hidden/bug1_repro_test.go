package engine_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/Bz-Lxt/chunkvault/config"
	"github.com/Bz-Lxt/chunkvault/engine"
)

func TestPutSurvivesOverlappingCollect(t *testing.T) {
	v, err := engine.Open(config.Config{Dir: t.TempDir(), ChunkSize: 32})
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	body := []byte("shared-prefix-window-32-bytes!!XXXX")
	res, err := v.Put(context.Background(), "a", body)
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Unlink(context.Background(), res.Digest); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := v.Collect(context.Background()); done <- err }()
	time.Sleep(80 * time.Millisecond)
	res2, err := v.Put(context.Background(), "b", body)
	if err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	got, err := v.Get(context.Background(), res2.Digest)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("payload lost")
	}
}

func TestCollectCanceled(t *testing.T) {
	v, err := engine.Open(config.Config{Dir: t.TempDir(), ChunkSize: 32})
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := v.Collect(ctx); err == nil {
		t.Fatal("canceled collect")
	}
}

func TestCollectReturnsBeforeLongWait(t *testing.T) {
	v, err := engine.Open(config.Config{Dir: t.TempDir(), ChunkSize: 32})
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	start := time.Now()
	if _, err := v.Collect(context.Background()); err != nil {
		t.Fatal(err)
	}
	if time.Since(start) > 250*time.Millisecond {
		t.Fatal("collect held too long")
	}
}
