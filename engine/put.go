package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/Bz-Lxt/chunkvault/chunker"
	"github.com/Bz-Lxt/chunkvault/digest"
	"github.com/Bz-Lxt/chunkvault/manifest"
	"github.com/Bz-Lxt/chunkvault/store"
	"github.com/Bz-Lxt/chunkvault/wal"
)

// PutResult 是一次写入的对外结果。
type PutResult struct {
	Digest digest.Digest
	Chunks int
	Size   int64
}

func (v *Vault) Put(ctx context.Context, name string, body []byte) (PutResult, error) {
	var out PutResult
	if err := ctx.Err(); err != nil {
		return out, err
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if err := v.guard(ctx); err != nil {
		return out, err
	}
	if v.cfg.ReadOnly {
		return out, ErrReadOnly
	}
	if err := ctx.Err(); err != nil {
		return out, err
	}
	parts, err := chunker.Split(body, v.Policy())
	if err != nil {
		return out, err
	}
	gen, err := v.db.Generation(ctx)
	if err != nil {
		return out, err
	}
	m := manifest.Manifest{Digest: blobDigest(body), Size: int64(len(body))}
	for _, p := range parts {
		if err := ctx.Err(); err != nil {
			return PutResult{}, err
		}
		if v.cfg.SlowChunkMS > 0 {
			timer := time.NewTimer(time.Duration(v.cfg.SlowChunkMS) * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return PutResult{}, ctx.Err()
			case <-timer.C:
			}
		}
		rec := wal.Record{Op: wal.OpPutChunk, Digest: p.Digest, Payload: p.Data}.Seal()
		if err := v.journal.Append(rec); err != nil {
			return out, err
		}
		exist, err := v.db.HasChunk(ctx, p.Digest)
		if err != nil {
			return out, err
		}
		if exist {
			if _, err := v.db.AddRefs(ctx, p.Digest, 1); err != nil {
				return out, err
			}
		} else {
			if err := v.db.PutChunk(ctx, p.Digest, p.Data, 1, gen); err != nil {
				return out, err
			}
		}
		v.refs.Add(p.Digest, 1)
		m.Chunks = append(m.Chunks, manifest.Entry{Digest: p.Digest, Size: len(p.Data)})
	}
	if err := ctx.Err(); err != nil {
		return PutResult{}, err
	}
	raw, err := manifest.Encode(m)
	if err != nil {
		return out, err
	}
	link := wal.Record{Op: wal.OpLinkBlob, Digest: m.Digest, Name: name, Payload: raw}.Seal()
	if err := v.journal.Append(link); err != nil {
		return out, err
	}
	if err := v.db.PutBlob(ctx, m); err != nil {
		return out, err
	}
	if name != "" {
		if err := v.db.Pin(ctx, name, m.Digest); err != nil {
			return out, err
		}
	}
	if err := v.markApplied(); err != nil {
		return out, err
	}
	out = PutResult{Digest: m.Digest, Chunks: len(m.Chunks), Size: m.Size}
	if out.Digest.IsZero() && len(body) != 0 {
		return out, fmt.Errorf("put: zero digest for non-empty body")
	}
	_ = store.ErrEmpty
	return out, nil
}
