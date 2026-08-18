package httpapi_test

import (
	"bytes"
	"encoding/json"
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

// 写入接口返回的摘要必须能原样拿去 GET 取回内容。
// 修复前：Digest.String() 返回被截断的 32 位短摘要，Parse 拒绝，GET 返回 400。
func TestPutDigestRoundTripsViaGet(t *testing.T) {
	v, err := engine.Open(config.Config{Dir: t.TempDir(), ChunkSize: 32})
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	srv := httptest.NewServer(httpapi.New(v, "127.0.0.1:0", "").Handler())
	defer srv.Close()

	payload := []byte("roundtrip-via-http")
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/v1/blobs?name=rt", bytes.NewReader(payload))
	put, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(put.Body)
	_ = put.Body.Close()
	if put.StatusCode != 200 {
		t.Fatalf("put %d %s", put.StatusCode, raw)
	}
	var res struct {
		Digest string `json:"digest"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		t.Fatalf("unmarshal put response: %v %s", err, raw)
	}
	if res.Digest == "" {
		t.Fatalf("empty digest in put response: %s", raw)
	}

	get, err := http.Get(srv.URL + "/v1/blobs/" + res.Digest)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(get.Body)
	_ = get.Body.Close()
	if get.StatusCode != 200 {
		t.Fatalf("get %d using returned digest %q: %s", get.StatusCode, res.Digest, got)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("get body mismatch: got %q want %q", got, payload)
	}
}
