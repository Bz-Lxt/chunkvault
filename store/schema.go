package store

const schemaSQL = `
PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;
PRAGMA busy_timeout=5000;

CREATE TABLE IF NOT EXISTS blobs (
  digest TEXT PRIMARY KEY,
  size INTEGER NOT NULL,
  chunk_count INTEGER NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS chunks (
  digest TEXT PRIMARY KEY,
  size INTEGER NOT NULL,
  refs INTEGER NOT NULL,
  data BLOB NOT NULL,
  created_at TEXT NOT NULL,
  generation INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS blob_chunks (
  blob_digest TEXT NOT NULL,
  seq INTEGER NOT NULL,
  chunk_digest TEXT NOT NULL,
  PRIMARY KEY (blob_digest, seq),
  FOREIGN KEY (blob_digest) REFERENCES blobs(digest) ON DELETE CASCADE,
  FOREIGN KEY (chunk_digest) REFERENCES chunks(digest)
);

CREATE TABLE IF NOT EXISTS pins (
  name TEXT PRIMARY KEY,
  blob_digest TEXT NOT NULL,
  FOREIGN KEY (blob_digest) REFERENCES blobs(digest)
);

CREATE TABLE IF NOT EXISTS meta (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
`

const (
	MetaGeneration = "generation"
	MetaOpenedAt   = "opened_at"
	MetaWALApplied = "wal_applied"
)
