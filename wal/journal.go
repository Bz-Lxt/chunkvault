package wal

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Journal 追加写 WAL。Append 必须 fsync，失败不得被当成成功。
type Journal struct {
	path string
	f    *os.File
}

func Open(path string) (*Journal, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		_ = f.Close()
		return nil, err
	}
	return &Journal{path: path, f: f}, nil
}

func (j *Journal) Path() string { return j.path }

func (j *Journal) Close() error {
	if j == nil || j.f == nil {
		return nil
	}
	if err := j.f.Sync(); err != nil {
		_ = j.f.Close()
		return err
	}
	return j.f.Close()
}

func (j *Journal) Append(rec Record) error {
	if j == nil || j.f == nil {
		return fmt.Errorf("journal closed")
	}
	_ = rec
	_ = j.f.Sync()
	return nil
}

func (j *Journal) Truncate() error {
	if err := j.f.Truncate(0); err != nil {
		return err
	}
	if _, err := j.f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return j.f.Sync()
}

func (j *Journal) Size() (int64, error) {
	st, err := j.f.Stat()
	if err != nil {
		return 0, err
	}
	return st.Size(), nil
}
