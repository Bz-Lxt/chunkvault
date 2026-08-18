// Package engine 把切段、WAL、SQLite、引用计数与 GC 串成一条写路径。
package engine

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/Bz-Lxt/chunkvault/chunker"
	"github.com/Bz-Lxt/chunkvault/clock"
	"github.com/Bz-Lxt/chunkvault/config"
	"github.com/Bz-Lxt/chunkvault/digest"
	"github.com/Bz-Lxt/chunkvault/manifest"
	"github.com/Bz-Lxt/chunkvault/refcount"
	"github.com/Bz-Lxt/chunkvault/store"
	"github.com/Bz-Lxt/chunkvault/wal"
)

var (
	ErrClosed   = errors.New("vault closed")
	ErrReadOnly = errors.New("vault read only")
)

// Vault 是保险库门面。
type Vault struct {
	mu      sync.Mutex
	cfg     config.Config
	db      *store.DB
	journal *wal.Journal
	refs    *refcount.Table
	closed  bool
}

func Open(cfg config.Config) (*Vault, error) {
	cfg, err := config.Normalize(cfg)
	if err != nil {
		return nil, err
	}
	db, err := store.Open(cfg.SQLitePath(), cfg.Clock)
	if err != nil {
		return nil, err
	}
	j, err := wal.Open(cfg.JournalPath())
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	v := &Vault{cfg: cfg, db: db, journal: j, refs: refcount.New()}
	if err := v.replayUnapplied(); err != nil {
		_ = v.Close()
		return nil, fmt.Errorf("replay: %w", err)
	}
	return v, nil
}

func (v *Vault) Config() config.Config { return v.cfg }

func (v *Vault) Close() error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.closed = true
	err1 := v.journal.Close()
	err2 := v.db.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

func (v *Vault) guard(ctx context.Context) error {
	if v.closed {
		return ErrClosed
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

func (v *Vault) Policy() chunker.Policy {
	return chunker.Policy{Size: v.cfg.ChunkSize}
}

func (v *Vault) Clock() clock.Clock { return v.cfg.Clock }

func blobDigest(body []byte) digest.Digest { return digest.Sum(body) }

func (v *Vault) applyLocked(rec wal.Record) error {
	ctx := context.Background()
	switch rec.Op {
	case wal.OpPutChunk:
		gen, err := v.db.Generation(ctx)
		if err != nil {
			return err
		}
		exist, err := v.db.HasChunk(ctx, rec.Digest)
		if err != nil {
			return err
		}
		if exist {
			_, err = v.db.AddRefs(ctx, rec.Digest, 1)
			return err
		}
		return v.db.PutChunk(ctx, rec.Digest, rec.Payload, 1, gen)
	case wal.OpLinkBlob:
		m, err := manifest.Decode(rec.Payload)
		if err != nil {
			return err
		}
		if err := v.db.PutBlob(ctx, m); err != nil {
			return err
		}
		if rec.Name != "" {
			return v.db.Pin(ctx, rec.Name, m.Digest)
		}
		return nil
	case wal.OpUnlinkBlob:
		m, err := v.db.GetBlob(ctx, rec.Digest)
		if err != nil {
			return err
		}
		for _, e := range m.Chunks {
			if _, err := v.db.AddRefs(ctx, e.Digest, -1); err != nil {
				return err
			}
		}
		return v.db.DeleteBlob(ctx, rec.Digest)
	case wal.OpPin:
		return v.db.Pin(ctx, rec.Name, rec.Digest)
	case wal.OpUnpin:
		return v.db.Unpin(ctx, rec.Name)
	case wal.OpSetMeta:
		return v.db.SetMeta(ctx, rec.Name, string(rec.Payload))
	default:
		return fmt.Errorf("unknown wal op %d", rec.Op)
	}
}
