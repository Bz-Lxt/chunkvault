package engine

import (
	"context"

	"github.com/Bz-Lxt/chunkvault/gc"
)

func (v *Vault) Preview(ctx context.Context) (gc.Plan, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if err := v.guard(ctx); err != nil {
		return gc.Plan{}, err
	}
	return (gc.Planner{DB: v.db}).Build(ctx)
}
