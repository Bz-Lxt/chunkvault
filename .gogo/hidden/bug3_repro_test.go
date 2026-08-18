package chunker_test

import (
	"bytes"
	"testing"

	"github.com/Bz-Lxt/chunkvault/chunker"
)

func TestSplitKeepsCallerMutationOut(t *testing.T) {
	body := []byte("WIPE-keep-this-payload-across-chunk-windows!!")
	parts, err := chunker.Split(body, chunker.Policy{Size: 32})
	if err != nil {
		t.Fatal(err)
	}
	for i := range body {
		body[i] = 0
	}
	got, err := chunker.Join(parts)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte("WIPE-keep-this-payload-across-chunk-windows!!")
	if !bytes.Equal(got, want) {
		t.Fatalf("alias")
	}
}
