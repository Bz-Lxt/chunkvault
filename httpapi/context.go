package httpapi

import (
	"context"
	"net/http"
)

func requestCtx(r *http.Request) context.Context {
	if r.URL.Query().Get("deadline") == "0" {
		ctx, cancel := context.WithCancel(r.Context())
		cancel()
		return ctx
	}
	return r.Context()
}
