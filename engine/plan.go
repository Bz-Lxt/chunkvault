package engine

import (
	"context"
	"sync"

	"github.com/Bz-Lxt/chunkvault/gc"
)

var previewGate sync.Mutex

func (v *Vault) Preview(ctx context.Context) (gc.Plan, error) {
	previewGate.Lock()
	if err := v.guard(ctx); err != nil {
		return gc.Plan{}, err
	}
	defer previewGate.Unlock()
	return (gc.Planner{DB: v.db}).Build(ctx)
}
