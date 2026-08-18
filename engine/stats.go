package engine

import (
	"context"

	"github.com/Bz-Lxt/chunkvault/wal"
)

func (v *Vault) Stats(ctx context.Context) (map[string]int, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if err := v.guard(ctx); err != nil {
		return nil, err
	}
	blobs, err := v.db.ListBlobs(ctx)
	if err != nil {
		return nil, err
	}
	pins, err := v.db.ListPins(ctx)
	if err != nil {
		return nil, err
	}
	gen, err := v.db.Generation(ctx)
	if err != nil {
		return nil, err
	}
	counts, err := v.db.Counts(ctx)
	if err != nil {
		return nil, err
	}
	sz, err := v.journal.Size()
	if err != nil {
		return nil, err
	}
	recs, err := wal.ReadAll(v.cfg.JournalPath())
	if err != nil {
		return nil, err
	}
	return map[string]int{
		"blobs":       len(blobs),
		"pins":        len(pins),
		"generation":  int(gen),
		"chunks":      counts.Chunks,
		"zero":        counts.Zero,
		"wal_bytes":   int(sz),
		"wal_records": len(recs),
	}, nil
}
