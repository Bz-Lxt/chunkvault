package store

import (
	"context"
	"fmt"

	"github.com/Bz-Lxt/chunkvault/digest"
)

func (db *DB) AddRefs(ctx context.Context, d digest.Digest, delta int64) (int64, error) {
	var refs int64
	err := db.sql.QueryRowContext(ctx, `SELECT refs FROM chunks WHERE digest = ?`, d.String()).Scan(&refs)
	if err != nil {
		return 0, fmt.Errorf("chunk %s: %w", d, ErrNotFound)
	}
	// 有符号加减避免 uint 下溢；引用不得为负，钳到 0 让 GC 按 refs<=0 回收。
	refs += delta
	if refs < 0 {
		refs = 0
	}
	_, err = db.sql.ExecContext(ctx, `UPDATE chunks SET refs = ? WHERE digest = ?`, refs, d.String())
	return refs, err
}
