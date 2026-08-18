package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Bz-Lxt/chunkvault/digest"
)

var (
	ErrNotFound = errors.New("not found")
	ErrEmpty    = errors.New("empty")
)

func (db *DB) Pin(ctx context.Context, name string, d digest.Digest) error {
	if name == "" {
		return fmt.Errorf("pin: %w", ErrEmpty)
	}
	_, err := db.sql.ExecContext(ctx,
		`INSERT INTO pins(name, blob_digest) VALUES(?,?)
		 ON CONFLICT(name) DO UPDATE SET blob_digest=excluded.blob_digest`,
		name, d.String(),
	)
	return err
}

func (db *DB) Unpin(ctx context.Context, name string) error {
	res, err := db.sql.ExecContext(ctx, `DELETE FROM pins WHERE name = ?`, name)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("pin %s: %w", name, ErrNotFound)
	}
	return nil
}

func (db *DB) LookupPin(ctx context.Context, name string) (digest.Digest, error) {
	var hex string
	err := db.sql.QueryRowContext(ctx, `SELECT blob_digest FROM pins WHERE name = ?`, name).Scan(&hex)
	if err == sql.ErrNoRows {
		return digest.Digest{}, fmt.Errorf("pin %s: %w", name, ErrNotFound)
	}
	if err != nil {
		return digest.Digest{}, err
	}
	return scanDigest(hex)
}

func (db *DB) ListPins(ctx context.Context) (map[string]digest.Digest, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT name, blob_digest FROM pins ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]digest.Digest{}
	for rows.Next() {
		var name, hex string
		if err := rows.Scan(&name, &hex); err != nil {
			return nil, err
		}
		d, err := scanDigest(hex)
		if err != nil {
			return nil, err
		}
		out[name] = d
	}
	return out, rows.Err()
}
