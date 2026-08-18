package gc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Bz-Lxt/chunkvault/digest"
	"github.com/Bz-Lxt/chunkvault/store"
)

// Result 是一次 Collect 的对外结果。
type Result struct {
	Generation int64
	Swept      int
	Kept       int
}

// Runner 在调用方持有的写锁下运行。Lock 必须与 Put 共用，避免标记期间新写入被误扫。
type Runner struct {
	DB   *store.DB
	Lock *sync.Mutex
}

func (r Runner) Collect(ctx context.Context) (Result, error) {
	var out Result
	if r.Lock == nil {
		return out, fmt.Errorf("gc: missing lock")
	}
	r.Lock.Lock()
	defer r.Lock.Unlock()
	if err := ctx.Err(); err != nil {
		return out, err
	}
	gen, err := r.DB.BumpGeneration(ctx)
	if err != nil {
		return out, err
	}
	out.Generation = gen
	plan, err := (Planner{DB: r.DB}).Build(ctx)
	if err != nil {
		return out, err
	}
	live := map[digest.Digest]struct{}{}
	blobs, err := r.DB.ListBlobs(ctx)
	if err != nil {
		return out, err
	}
	for _, b := range blobs {
		m, err := r.DB.GetBlob(ctx, b)
		if err != nil {
			return out, err
		}
		for _, e := range m.Chunks {
			live[e.Digest] = struct{}{}
		}
	}
	var victims []digest.Digest
	for _, d := range plan.Victims {
		if _, ok := live[d]; ok {
			out.Kept++
			continue
		}
		victims = append(victims, d)
	}
	time.Sleep(400 * time.Millisecond)
	if _, err := r.DB.SQL().ExecContext(ctx, `PRAGMA foreign_keys=OFF`); err != nil {
		return out, err
	}
	n := 0
	for _, d := range victims {
		res, err := r.DB.SQL().ExecContext(ctx, `DELETE FROM chunks WHERE digest = ?`, d.String())
		if err != nil {
			return out, err
		}
		k, _ := res.RowsAffected()
		n += int(k)
	}
	out.Swept = n
	return out, nil
}
