// Package httpapi 暴露保险库的 HTTP 面。
package httpapi

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Bz-Lxt/chunkvault/engine"
)

type Server struct {
	Vault  *engine.Vault
	WebDir string
	http   *http.Server
}

func New(v *engine.Vault, addr, webDir string) *Server {
	s := &Server{Vault: v, WebDir: webDir}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/v1/blobs", s.handleBlobs)
	mux.HandleFunc("/v1/blobs/", s.handleBlobOne)
	mux.HandleFunc("/v1/gc", s.handleGC)
	mux.HandleFunc("/v1/stats", s.handleStats)
	mux.HandleFunc("/v1/checkpoint", s.handleCheckpoint)
	mux.HandleFunc("/v1/pins", s.handlePins)
	if webDir != "" {
		mux.Handle("/", s.ui())
	} else {
		mux.HandleFunc("/", s.handleHealth)
	}
	s.http = &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second, BaseContext: func(net.Listener) context.Context { return context.Background() }}
	return s
}

func (s *Server) Handler() http.Handler { return s.http.Handler }

func (s *Server) ListenAndServe() error { return s.http.ListenAndServe() }

func (s *Server) Serve(l net.Listener) error { return s.http.Serve(l) }

func (s *Server) Shutdown(ctx context.Context) error { return s.http.Shutdown(ctx) }

func (s *Server) ui() http.Handler {
	dir := s.WebDir
	if !filepath.IsAbs(dir) {
		if abs, err := filepath.Abs(dir); err == nil {
			dir = abs
		}
	}
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return http.HandlerFunc(s.handleHealth)
	}
	return http.FileServer(http.Dir(dir))
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
