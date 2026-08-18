package httpapi_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Bz-Lxt/chunkvault/config"
	"github.com/Bz-Lxt/chunkvault/engine"
	"github.com/Bz-Lxt/chunkvault/httpapi"
)

func TestHealthAndBlobHTTP(t *testing.T) {
	v, err := engine.Open(config.Config{Dir: t.TempDir(), ChunkSize: 32})
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	real := httptest.NewServer(httpapi.New(v, "127.0.0.1:0", "").Handler())
	defer real.Close()
	resp, err := http.Get(real.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("health %d", resp.StatusCode)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	req, _ := http.NewRequest(http.MethodPut, real.URL+"/v1/blobs?name=n1", bytes.NewReader([]byte("hello-http")))
	put, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(put.Body)
	_ = put.Body.Close()
	if put.StatusCode != 200 {
		t.Fatalf("put %d %s", put.StatusCode, body)
	}
}
