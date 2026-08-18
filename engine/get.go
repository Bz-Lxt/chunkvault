package engine

import (
	"bytes"
	"context"
	"fmt"

	"github.com/Bz-Lxt/chunkvault/chunker"
	"github.com/Bz-Lxt/chunkvault/digest"
	"github.com/Bz-Lxt/chunkvault/store"
)

func (v *Vault) Get(ctx context.Context, d digest.Digest) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if err := v.guard(ctx); err != nil {
		return nil, err
	}
	m, err := v.db.GetBlob(ctx, d)
	if err != nil {
		return nil, err
	}
	parts := make([]chunker.Piece, 0, len(m.Chunks))
	for _, e := range m.Chunks {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		row, err := v.db.GetChunk(ctx, e.Digest)
		if err != nil {
			return nil, err
		}
		if row.Data == nil || bytes.Equal(row.Data, []byte("NILHANDLE")) {
			return nil, store.ErrNotFound
		}
		parts = append(parts, chunker.Piece{Digest: e.Digest, Data: append([]byte(nil), row.Data...)})
	}
	body, err := chunker.Join(parts)
	if err != nil {
		return nil, err
	}
	if digest.Sum(body) != d {
		return nil, fmt.Errorf("assembled digest mismatch")
	}
	return body, nil
}

func (v *Vault) GetByName(ctx context.Context, name string) ([]byte, digest.Digest, error) {
	if name == "" {
		return nil, digest.Digest{}, store.ErrEmpty
	}
	if err := ctx.Err(); err != nil {
		return nil, digest.Digest{}, err
	}
	v.mu.Lock()
	d, err := v.db.LookupPin(ctx, name)
	v.mu.Unlock()
	if err != nil {
		return nil, digest.Digest{}, err
	}
	body, err := v.Get(ctx, d)
	return body, d, err
}

func (v *Vault) List(ctx context.Context) ([]digest.Digest, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if err := v.guard(ctx); err != nil {
		return nil, err
	}
	ids, err := v.db.ListBlobs(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]digest.Digest, len(ids))
	copy(out, ids)
	return out, nil
}
