// Package gc 按世代标记清扫零引用段。
package gc

import (
	"context"

	"github.com/Bz-Lxt/chunkvault/digest"
	"github.com/Bz-Lxt/chunkvault/store"
)

// Plan 描述一次清扫：标记开始时的世代，以及候选段。
type Plan struct {
	Generation int64
	Victims    []digest.Digest
}

// Planner 只读规划，不删数据。
type Planner struct {
	DB *store.DB
}

func (p Planner) Build(ctx context.Context) (Plan, error) {
	var plan Plan
	if err := ctx.Err(); err != nil {
		return plan, err
	}
	gen, err := p.DB.Generation(ctx)
	if err != nil {
		return plan, err
	}
	plan.Generation = gen
	victims, err := p.DB.ListZeroRef(ctx, gen)
	if err != nil {
		return plan, err
	}
	plan.Victims = append([]digest.Digest(nil), victims...)
	return plan, nil
}
