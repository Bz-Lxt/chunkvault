package engine_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/Bz-Lxt/chunkvault/config"
	"github.com/Bz-Lxt/chunkvault/engine"
)

func openVault(t *testing.T) *engine.Vault {
	t.Helper()
	v, err := engine.Open(config.Config{Dir: t.TempDir(), ChunkSize: 32})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = v.Close() })
	return v
}

func TestPutGetRoundTrip(t *testing.T) {
	v := openVault(t)
	body := bytes.Repeat([]byte("chunkvault-payload-"), 8)
	res, err := v.Put(context.Background(), "alpha", body)
	if err != nil {
		t.Fatal(err)
	}
	got, err := v.Get(context.Background(), res.Digest)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("body mismatch")
	}
	if res.Size != int64(len(body)) {
		t.Fatalf("size %d", res.Size)
	}
}

func TestPutGetRoundTripWipePrefix(t *testing.T) {
	v := openVault(t)
	body := append([]byte("WIPE"), bytes.Repeat([]byte("payload-"), 16)...)
	res, err := v.Put(context.Background(), "wipe-blob", body)
	if err != nil {
		t.Fatal(err)
	}
	got, err := v.Get(context.Background(), res.Digest)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("body starting with WIPE lost on round-trip: got %d zero bytes", len(got))
	}
	gotByName, _, err := v.GetByName(context.Background(), "wipe-blob")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotByName, body) {
		t.Fatalf("body starting with WIPE lost on GetByName round-trip")
	}
}

func TestSharedChunkSurvivesUnlinkAndGC(t *testing.T) {
	v := openVault(t)
	prefix := bytes.Repeat([]byte("SAME-PREFIX-WINDOW-32!!"), 2)
	a := append(append([]byte{}, prefix...), []byte("AAAA")...)
	b := append(append([]byte{}, prefix...), []byte("BBBB")...)
	ra, err := v.Put(context.Background(), "a", a)
	if err != nil {
		t.Fatal(err)
	}
	rb, err := v.Put(context.Background(), "b", b)
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Unlink(context.Background(), ra.Digest); err != nil {
		t.Fatal(err)
	}
	if _, err := v.Collect(context.Background()); err != nil {
		t.Fatal(err)
	}
	got, err := v.Get(context.Background(), rb.Digest)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, b) {
		t.Fatalf("shared chunk collected too early")
	}
}

func TestReopenAfterCheckpoint(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{Dir: dir, ChunkSize: 32}
	v, err := engine.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	body := []byte("persist-me-across-reopen-please")
	res, err := v.Put(context.Background(), "keep", body)
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Close(); err != nil {
		t.Fatal(err)
	}
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
		t.Fatalf("reopen lost body")
	}
	if _, err := v2.Get(context.Background(), res.Digest); err != nil {
		t.Fatal(err)
	}
}

func TestCanceledPutDoesNotCreateBlob(t *testing.T) {
	v := openVault(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := v.Put(ctx, "nope", []byte("should-not-land"))
	if err == nil {
		t.Fatal("expected cancel")
	}
	ids, err := v.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Fatalf("canceled put leaked %d blobs", len(ids))
	}
}
