package store

import (
	"context"
	"fmt"
)

// Counts 汇总库内对象与段。
type Counts struct {
	Blobs  int
	Chunks int
	Pins   int
	Zero   int
}

func (db *DB) Counts(ctx context.Context) (Counts, error) {
	var c Counts
	if err := db.sql.QueryRowContext(ctx, `SELECT COUNT(1) FROM blobs`).Scan(&c.Blobs); err != nil {
		return c, err
	}
	if err := db.sql.QueryRowContext(ctx, `SELECT COUNT(1) FROM chunks`).Scan(&c.Chunks); err != nil {
		return c, err
	}
	if err := db.sql.QueryRowContext(ctx, `SELECT COUNT(1) FROM pins`).Scan(&c.Pins); err != nil {
		return c, err
	}
	if err := db.sql.QueryRowContext(ctx, `SELECT COUNT(1) FROM chunks WHERE refs <= 0`).Scan(&c.Zero); err != nil {
		return c, err
	}
	return c, nil
}

func (c Counts) String() string {
	return fmt.Sprintf("blobs=%d chunks=%d pins=%d zero=%d", c.Blobs, c.Chunks, c.Pins, c.Zero)
}
