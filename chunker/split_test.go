package chunker_test

import (
	"bytes"
	"testing"

	"github.com/Bz-Lxt/chunkvault/chunker"
)

func TestSplitJoinCount(t *testing.T) {
	body := []byte("abcdefghijklmnopqrstuvwxyz0123456789ABCD")
	parts, err := chunker.Split(body, chunker.Policy{Size: 32})
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != 2 {
		t.Fatalf("parts %d", len(parts))
	}
	got, err := chunker.Join(parts)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, body) {
		t.Fatal("join")
	}
}
