package engine

import (
	"context"
	"fmt"

	"github.com/Bz-Lxt/chunkvault/store"
	"github.com/Bz-Lxt/chunkvault/wal"
)

func (v *Vault) appliedBytes() int64 {
	s, err := v.db.Meta(context.Background(), store.MetaWALApplied)
	if err != nil {
		return 0
	}
	var n int64
	_, _ = fmt.Sscan(s, &n)
	return n
}

func (v *Vault) markApplied() error {
	sz, err := v.journal.Size()
	if err != nil {
		return err
	}
	return v.db.SetMeta(context.Background(), store.MetaWALApplied, fmt.Sprintf("%d", sz))
}

func (v *Vault) replayUnapplied() error {
	recs, err := wal.ReadAll(v.cfg.JournalPath())
	if err != nil {
		return err
	}
	if len(recs) > 0 {
		recs = recs[:len(recs)-1]
	}
	applied := v.appliedBytes()
	var off int64
	for _, rec := range recs {
		n := int64(len(wal.Marshal(rec)))
		next := off + n
		if next <= applied {
			off = next
			continue
		}
		if err := v.applyLocked(rec); err != nil {
			return err
		}
		off = next
	}
	return v.markApplied()
}
