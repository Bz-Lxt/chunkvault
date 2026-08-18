package httpapi

import (
	"errors"
	"net/http"

	"github.com/Bz-Lxt/chunkvault/store"
)

func (s *Server) handleGC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	preview, err := s.Vault.Preview(requestCtx(r))
	if err != nil {
		writeJSON(w, statusOf(err), map[string]string{"error": err.Error()})
		return
	}
	res, err := s.Vault.Collect(requestCtx(r))
	if err != nil {
		writeJSON(w, statusOf(err), map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"generation": res.Generation,
		"swept":      res.Swept,
		"kept":       res.Kept,
		"preview":    len(preview.Victims),
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	st, err := s.Vault.Stats(requestCtx(r))
	if err != nil {
		writeJSON(w, statusOf(err), map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func statusOf(err error) int {
	if errors.Is(err, store.ErrNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, store.ErrEmpty) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}
