package engine

import (
	"context"

	"github.com/Bz-Lxt/chunkvault/gc"
)

func (v *Vault) Preview(ctx context.Context) (gc.Plan, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	return (gc.Planner{DB: v.db}).Build(ctx)
}
