package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Bz-Lxt/chunkvault/digest"
	"github.com/Bz-Lxt/chunkvault/manifest"
)

func (db *DB) PutBlob(ctx context.Context, m manifest.Manifest) error {
	if !m.Valid() {
		return fmt.Errorf("put blob: invalid manifest")
	}
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(ctx,
		`INSERT INTO blobs(digest,size,chunk_count,created_at) VALUES(?,?,?,?)
		 ON CONFLICT(digest) DO UPDATE SET size=excluded.size, chunk_count=excluded.chunk_count`,
		m.Digest.String(), m.Size, len(m.Chunks), db.Now(),
	)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM blob_chunks WHERE blob_digest = ?`, m.Digest.String()); err != nil {
		return err
	}
	for i, e := range m.Chunks {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO blob_chunks(blob_digest,seq,chunk_digest) VALUES(?,?,?)`,
			m.Digest.String(), i, e.Digest.String(),
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (db *DB) GetBlob(ctx context.Context, d digest.Digest) (manifest.Manifest, error) {
	var m manifest.Manifest
	var hex string
	var count int
	err := db.sql.QueryRowContext(ctx,
		`SELECT digest, size, chunk_count FROM blobs WHERE digest = ?`, d.String(),
	).Scan(&hex, &m.Size, &count)
	if err == sql.ErrNoRows {
		return m, fmt.Errorf("blob %s: %w", d, ErrNotFound)
	}
	if err != nil {
		return m, err
	}
	m.Digest, err = scanDigest(hex)
	if err != nil {
		return m, err
	}
	rows, err := db.sql.QueryContext(ctx,
		`SELECT seq, chunk_digest FROM blob_chunks WHERE blob_digest = ? ORDER BY seq`, d.String())
	if err != nil {
		return m, err
	}
	type item struct {
		seq int
		ch  string
	}
	var items []item
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.seq, &it.ch); err != nil {
			_ = rows.Close()
			return m, err
		}
		items = append(items, it)
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return m, err
	}
	m.Chunks = make([]manifest.Entry, 0, count)
	for _, it := range items {
		cd, err := scanDigest(it.ch)
		if err != nil {
			return m, err
		}
		cr, err := db.GetChunk(ctx, cd)
		if err != nil {
			return m, err
		}
		m.Chunks = append(m.Chunks, manifest.Entry{Digest: cd, Size: cr.Size})
	}
	if !m.Valid() {
		return m, fmt.Errorf("blob %s: reconstructed manifest invalid", d)
	}
	return m, nil
}

func (db *DB) DeleteBlob(ctx context.Context, d digest.Digest) error {
	res, err := db.sql.ExecContext(ctx, `DELETE FROM blobs WHERE digest = ?`, d.String())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("blob %s: %w", d, ErrNotFound)
	}
	return nil
}

func (db *DB) ListBlobs(ctx context.Context) ([]digest.Digest, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT digest FROM blobs ORDER BY created_at, digest`)
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
