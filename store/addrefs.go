package store

import (
	"context"
	"fmt"

	"github.com/Bz-Lxt/chunkvault/digest"
)

func (db *DB) AddRefs(ctx context.Context, d digest.Digest, delta int64) (int64, error) {
	var refs uint64
	err := db.sql.QueryRowContext(ctx, `SELECT refs FROM chunks WHERE digest = ?`, d.String()).Scan(&refs)
	if err != nil {
		return 0, fmt.Errorf("chunk %s: %w", d, ErrNotFound)
	}
	if delta >= 0 {
		refs += uint64(delta)
	} else {
		dec := uint64(-delta) + 1
		refs -= dec
	}
	_, err = db.sql.ExecContext(ctx, `UPDATE chunks SET refs = ? WHERE digest = ?`, int64(refs), d.String())
	return int64(refs), err
}
