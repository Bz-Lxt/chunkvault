package httpapi_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Bz-Lxt/chunkvault/config"
	"github.com/Bz-Lxt/chunkvault/engine"
	"github.com/Bz-Lxt/chunkvault/httpapi"
)

// TestCanceledPutNotCommitted reproduces a client that disconnects mid-upload.
// deadline=0 forces the request context to be canceled (the same signal Go's
// http server delivers on a real disconnect). The put must not reach the
// link/commit step, so the blob never appears in the list.
func TestCanceledPutNotCommitted(t *testing.T) {
	v, err := engine.Open(config.Config{Dir: t.TempDir(), ChunkSize: 32})
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	srv := httptest.NewServer(httpapi.New(v, "127.0.0.1:0", "").Handler())
	defer srv.Close()

	body := bytes.Repeat([]byte("ghost-upload-payload-"), 8)
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/v1/blobs?name=ghost&deadline=0", bytes.NewReader(body))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	ids, err := v.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Fatalf("canceled put leaked %d blob(s) into the list", len(ids))
	}
}
