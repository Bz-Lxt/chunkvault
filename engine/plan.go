package engine

import (
	"context"

	"github.com/Bz-Lxt/chunkvault/gc"
)

func (v *Vault) Preview(ctx context.Context) (gc.Plan, error) {
	v.mu.Lock()
	if err := v.guard(ctx); err != nil {
		return gc.Plan{}, err
	}
	defer v.mu.Unlock()
	return (gc.Planner{DB: v.db}).Build(ctx)
}
