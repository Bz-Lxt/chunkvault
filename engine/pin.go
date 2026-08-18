package engine

import (
	"context"

	"github.com/Bz-Lxt/chunkvault/digest"
	"github.com/Bz-Lxt/chunkvault/store"
	"github.com/Bz-Lxt/chunkvault/wal"
)

func (v *Vault) Pin(ctx context.Context, name string, d digest.Digest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if err := v.guard(ctx); err != nil {
		return err
	}
	if name == "" {
		return store.ErrEmpty
	}
	if _, err := v.db.GetBlob(ctx, d); err != nil {
		return err
	}
	if err := v.journal.Append(wal.Record{Op: wal.OpPin, Digest: d, Name: name}.Seal()); err != nil {
		return err
	}
	if err := v.db.Pin(ctx, name, d); err != nil {
		return err
	}
	return v.markApplied()
}

func (v *Vault) Pins(ctx context.Context) (map[string]digest.Digest, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if err := v.guard(ctx); err != nil {
		return nil, err
	}
	src, err := v.db.ListPins(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string]digest.Digest, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out, nil
}
