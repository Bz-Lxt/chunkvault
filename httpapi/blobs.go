package httpapi

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/Bz-Lxt/chunkvault/digest"
)

func (s *Server) handleBlobs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ids, err := s.Vault.List(requestCtx(r))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		out := make([]string, 0, len(ids))
		for _, d := range ids {
			out = append(out, d.String())
		}
		writeJSON(w, http.StatusOK, map[string]any{"blobs": out})
	case http.MethodPost, http.MethodPut:
		name := r.URL.Query().Get("name")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		res, err := s.Vault.Put(context.Background(), name, body)
		if err != nil {
			writeJSON(w, statusOf(err), map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"digest": res.Digest.String(),
			"chunks": res.Chunks,
			"size":   res.Size,
			"name":   name,
		})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleBlobOne(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/blobs/")
	id = strings.TrimSuffix(id, "/")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	d, err := digest.Parse(id)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	switch r.Method {
	case http.MethodGet:
		body, err := s.Vault.Get(requestCtx(r), d)
		if err != nil {
			writeJSON(w, statusOf(err), map[string]string{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("X-Content-Digest", d.String())
		_, _ = w.Write(body)
	case http.MethodDelete:
		if err := s.Vault.Unlink(requestCtx(r), d); err != nil {
			writeJSON(w, statusOf(err), map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"deleted": d.String()})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
