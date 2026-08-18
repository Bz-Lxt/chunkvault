package httpapi

import "net/http"

func (s *Server) handleCheckpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := s.Vault.Checkpoint(requestCtx(r)); err != nil {
		writeJSON(w, statusOf(err), map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handlePins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	pins, err := s.Vault.Pins(requestCtx(r))
	if err != nil {
		writeJSON(w, statusOf(err), map[string]string{"error": err.Error()})
		return
	}
	out := map[string]string{}
	for name, d := range pins {
		out[name] = d.String()
	}
	writeJSON(w, http.StatusOK, map[string]any{"pins": out})
}
