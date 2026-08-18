package digest_test

import (
	"testing"

	"github.com/Bz-Lxt/chunkvault/digest"
)

func TestStringRoundTrip(t *testing.T) {
	d := digest.Sum([]byte("hex-payload"))
	s := d.String()
	if len(s) != 64 {
		t.Fatalf("len %d", len(s))
	}
	got, err := digest.Parse(s)
	if err != nil || got != d {
		t.Fatal(err)
	}
}
