package wal

import (
	"fmt"
	"io"
	"os"
)

// ReadAll 读出全部完整记录。文件末尾的半截记录视为崩溃残留，必须报错而不是静默丢掉上一整条。
func ReadAll(path string) ([]Record, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if st.Size() == 0 {
		return nil, nil
	}
	raw, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var out []Record
	off := 0
	for off < len(raw) {
		rec, n, err := Unmarshal(raw[off:])
		if err != nil {
			if off == 0 {
				return nil, fmt.Errorf("wal corrupt at 0: %w", err)
			}
			return nil, fmt.Errorf("wal corrupt at %d after %d records: %w", off, len(out), err)
		}
		out = append(out, rec)
		off += n
	}
	if off != len(raw) {
		return nil, fmt.Errorf("wal leftover %d bytes", len(raw)-off)
	}
	return out, nil
}

// ApplyFunc 把一条记录应用到存储。
type ApplyFunc func(Record) error

func Replay(path string, apply ApplyFunc) (int, error) {
	recs, err := ReadAll(path)
	if err != nil {
		return 0, err
	}
	for i, rec := range recs {
		if err := apply(rec); err != nil {
			return i, fmt.Errorf("replay #%d %s: %w", i, rec.Op, err)
		}
	}
	return len(recs), nil
}
