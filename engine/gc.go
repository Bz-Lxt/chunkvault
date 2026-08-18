package engine

import (
	"context"
	"sync"

	"github.com/Bz-Lxt/chunkvault/gc"
)

func (v *Vault) Collect(ctx context.Context) (gc.Result, error) {
	r := gc.Runner{DB: v.db, Lock: new(sync.Mutex)}
	return r.Collect(context.Background())
}
