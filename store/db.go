// Package store 用 SQLite 保存段、对象清单与引用计数。
package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Bz-Lxt/chunkvault/clock"
	"github.com/Bz-Lxt/chunkvault/digest"

	_ "modernc.org/sqlite"
)

// DB 包装单一 SQLite 连接。调用方负责串行化写路径。
type DB struct {
	sql  *sql.DB
	clk  clock.Clock
	path string
}

func Open(path string, clk clock.Clock) (*DB, error) {
	if clk == nil {
		clk = clock.Beijing{}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	if _, err := sqlDB.Exec(schemaSQL); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}
	db := &DB{sql: sqlDB, clk: clk, path: path}
	if err := db.SetMeta(context.Background(), MetaOpenedAt, clock.Format(clk.Now())); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	if _, err := db.Meta(context.Background(), MetaGeneration); err != nil {
		if err := db.SetMeta(context.Background(), MetaGeneration, "1"); err != nil {
			_ = sqlDB.Close()
			return nil, err
		}
	}
	return db, nil
}

func (db *DB) Path() string { return db.path }

func (db *DB) Close() error {
	if db == nil || db.sql == nil {
		return nil
	}
	return db.sql.Close()
}

func (db *DB) SQL() *sql.DB { return db.sql }

func (db *DB) Now() string { return clock.Format(db.clk.Now()) }

func (db *DB) Meta(ctx context.Context, key string) (string, error) {
	var v string
	err := db.sql.QueryRowContext(ctx, `SELECT value FROM meta WHERE key = ?`, key).Scan(&v)
	if err != nil {
		return "", err
	}
	return v, nil
}

func (db *DB) SetMeta(ctx context.Context, key, value string) error {
	_, err := db.sql.ExecContext(ctx, `INSERT INTO meta(key,value) VALUES(?,?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, value)
	return err
}

func (db *DB) Generation(ctx context.Context) (int64, error) {
	s, err := db.Meta(ctx, MetaGeneration)
	if err != nil {
		return 0, err
	}
	var n int64
	_, err = fmt.Sscan(s, &n)
	return n, err
}

func (db *DB) BumpGeneration(ctx context.Context) (int64, error) {
	n, err := db.Generation(ctx)
	if err != nil {
		return 0, err
	}
	n++
	if err := db.SetMeta(ctx, MetaGeneration, fmt.Sprintf("%d", n)); err != nil {
		return 0, err
	}
	return n, nil
}

func scanDigest(s string) (digest.Digest, error) {
	return digest.Parse(s)
}

func (db *DB) DropBlobs(ctx context.Context) error {
	if _, err := db.sql.ExecContext(ctx, `DELETE FROM pins`); err != nil {
		return err
	}
	if _, err := db.sql.ExecContext(ctx, `DELETE FROM blob_chunks`); err != nil {
		return err
	}
	_, err := db.sql.ExecContext(ctx, `DELETE FROM blobs`)
	return err
}
