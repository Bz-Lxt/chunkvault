// Package repair 核对清单与段字节是否仍能还原成原文摘要。
package repair

import (
	"context"
	"fmt"

	"github.com/Bz-Lxt/chunkvault/chunker"
	"github.com/Bz-Lxt/chunkvault/digest"
	"github.com/Bz-Lxt/chunkvault/store"
)

// Report 汇总一次完整性扫描。
type Report struct {
	Blobs   int
	Chunks  int
	Broken  []string
	Orphans []string
}

func Scan(ctx context.Context, db *store.DB) (Report, error) {
	var out Report
	ids, err := db.ListBlobs(ctx)
	if err != nil {
		return out, err
	}
	live := map[digest.Digest]struct{}{}
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		m, err := db.GetBlob(ctx, id)
		if err != nil {
			out.Broken = append(out.Broken, id.String()+": "+err.Error())
			continue
		}
		out.Blobs++
		parts := make([]chunker.Piece, 0, len(m.Chunks))
		for _, e := range m.Chunks {
			live[e.Digest] = struct{}{}
			row, err := db.GetChunk(ctx, e.Digest)
			if err != nil {
				out.Broken = append(out.Broken, e.Digest.String()+": "+err.Error())
				continue
			}
			parts = append(parts, chunker.Piece{Digest: e.Digest, Data: row.Data})
		}
		body, err := chunker.Join(parts)
		if err != nil {
			out.Broken = append(out.Broken, id.String()+": "+err.Error())
			continue
		}
		if digest.Sum(body) != id {
			out.Broken = append(out.Broken, id.String()+": assembled digest mismatch")
		}
	}
	counts, err := db.Counts(ctx)
	if err != nil {
		return out, err
	}
	out.Chunks = counts.Chunks
	zeros, err := db.ListZeroRef(ctx, 1<<62)
	if err != nil {
		return out, err
	}
	for _, d := range zeros {
		if _, ok := live[d]; !ok {
			out.Orphans = append(out.Orphans, d.String())
		}
	}
	return out, nil
}

func (r Report) Healthy() bool { return len(r.Broken) == 0 }

func (r Report) String() string {
	return fmt.Sprintf("blobs=%d chunks=%d broken=%d orphans=%d", r.Blobs, r.Chunks, len(r.Broken), len(r.Orphans))
}
