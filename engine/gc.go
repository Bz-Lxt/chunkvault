package engine

import (
	"context"

	"github.com/Bz-Lxt/chunkvault/gc"
)

func (v *Vault) Collect(ctx context.Context) (gc.Result, error) {
	if err := ctx.Err(); err != nil {
		return gc.Result{}, err
	}
	r := gc.Runner{DB: v.db, Lock: &v.mu}
	return r.Collect(ctx)
}
