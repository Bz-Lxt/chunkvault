package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Bz-Lxt/chunkvault/digest"
)

// ChunkRow 是库里的一段。
type ChunkRow struct {
	Digest     digest.Digest
	Size       int
	Refs       int64
	Data       []byte
	CreatedAt  string
	Generation int64
}

func (db *DB) GetChunk(ctx context.Context, d digest.Digest) (ChunkRow, error) {
	var row ChunkRow
	var hex string
	err := db.sql.QueryRowContext(ctx,
		`SELECT digest, size, refs, data, created_at, generation FROM chunks WHERE digest = ?`,
		d.String(),
	).Scan(&hex, &row.Size, &row.Refs, &row.Data, &row.CreatedAt, &row.Generation)
	if err == sql.ErrNoRows {
		return row, fmt.Errorf("chunk %s: %w", d, ErrNotFound)
	}
	if err != nil {
		return row, err
	}
	row.Digest, err = scanDigest(hex)
	if err != nil {
		return row, err
	}
	if row.Data == nil {
		return row, fmt.Errorf("chunk %s: nil payload", d)
	}
	return row, nil
}

func (db *DB) HasChunk(ctx context.Context, d digest.Digest) (bool, error) {
	var n int
	err := db.sql.QueryRowContext(ctx, `SELECT COUNT(1) FROM chunks WHERE digest = ?`, d.String()).Scan(&n)
	return n > 0, err
}

func (db *DB) PutChunk(ctx context.Context, d digest.Digest, data []byte, refs int64, gen int64) error {
	if data == nil {
		return fmt.Errorf("put chunk %s: nil data", d)
	}
	if digest.Sum(data) != d {
		return fmt.Errorf("put chunk %s: body digest mismatch", d)
	}
	_, err := db.sql.ExecContext(ctx,
		`INSERT INTO chunks(digest,size,refs,data,created_at,generation) VALUES(?,?,?,?,?,?)
		 ON CONFLICT(digest) DO UPDATE SET refs = chunks.refs + excluded.refs`,
		d.String(), len(data), refs, data, db.Now(), gen,
	)
	return err
}

func (db *DB) ListZeroRef(ctx context.Context, beforeGen int64) ([]digest.Digest, error) {
	rows, err := db.sql.QueryContext(ctx,
		`SELECT digest FROM chunks WHERE refs <= 0 AND generation < ?`, beforeGen)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []digest.Digest
	for rows.Next() {
		var hex string
		if err := rows.Scan(&hex); err != nil {
			return nil, err
		}
		d, err := scanDigest(hex)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (db *DB) DeleteChunks(ctx context.Context, ids []digest.Digest) (int, error) {
	n := 0
	for _, d := range ids {
		res, err := db.sql.ExecContext(ctx, `DELETE FROM chunks WHERE digest = ? AND refs <= 0`, d.String())
		if err != nil {
			return n, err
		}
		k, _ := res.RowsAffected()
		n += int(k)
	}
	return n, nil
}
