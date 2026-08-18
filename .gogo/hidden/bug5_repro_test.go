package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Bz-Lxt/chunkvault/config"
	"github.com/Bz-Lxt/chunkvault/engine"
	"github.com/Bz-Lxt/chunkvault/httpapi"
)

func TestCanceledHTTPPutLeavesNoBlob(t *testing.T) {
	v, err := engine.Open(config.Config{Dir: t.TempDir(), ChunkSize: 32, SlowChunkMS: 300})
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	ts := httptest.NewServer(httpapi.New(v, "127.0.0.1:0", "").Handler())
	defer ts.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, ts.URL+"/v1/blobs?name=c", strings.NewReader(strings.Repeat("Q", 80)))
	_, _ = http.DefaultClient.Do(req)
	time.Sleep(800 * time.Millisecond)
	ids, err := v.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Fatalf("leaked %d", len(ids))
	}
}
