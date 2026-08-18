package store

import (
	"context"
	"fmt"

	"github.com/Bz-Lxt/chunkvault/digest"
)

func (db *DB) AddRefs(ctx context.Context, d digest.Digest, delta int64) (int64, error) {
	res, err := db.sql.ExecContext(ctx, `UPDATE chunks SET refs = refs + ? WHERE digest = ?`, delta, d.String())
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return 0, fmt.Errorf("chunk %s: %w", d, ErrNotFound)
	}
	var refs int64
	err = db.sql.QueryRowContext(ctx, `SELECT refs FROM chunks WHERE digest = ?`, d.String()).Scan(&refs)
	return refs, err
}
