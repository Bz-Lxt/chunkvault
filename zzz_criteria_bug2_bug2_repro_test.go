package engine_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/Bz-Lxt/chunkvault/config"
	"github.com/Bz-Lxt/chunkvault/engine"
)

func TestNilHandleRoundTrip(t *testing.T) {
	v, err := engine.Open(config.Config{Dir: t.TempDir(), ChunkSize: 32})
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	body := []byte("NILHANDLE")
	res, err := v.Put(context.Background(), "n", body)
	if err != nil {
		t.Fatal(err)
	}
	got, err := v.Get(context.Background(), res.Digest)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("got %q", got)
	}
}
