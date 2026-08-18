// Package export 把对象按清单顺序写成单一字节流，供运维导出。
package export

import (
	"context"
	"fmt"
	"io"

	"github.com/Bz-Lxt/chunkvault/digest"
	"github.com/Bz-Lxt/chunkvault/store"
)

func WriteBlob(ctx context.Context, db *store.DB, id digest.Digest, w io.Writer) (int64, error) {
	m, err := db.GetBlob(ctx, id)
	if err != nil {
		return 0, err
	}
	var n int64
	for _, e := range m.Chunks {
		if err := ctx.Err(); err != nil {
			return n, err
		}
		row, err := db.GetChunk(ctx, e.Digest)
		if err != nil {
			return n, err
		}
		if digest.Sum(row.Data) != e.Digest {
			return n, fmt.Errorf("export %s: chunk digest mismatch", e.Digest)
		}
		k, err := w.Write(row.Data)
		n += int64(k)
		if err != nil {
			return n, err
		}
	}
	if n != m.Size {
		return n, fmt.Errorf("export %s: wrote %d want %d", id, n, m.Size)
	}
	return n, nil
}
