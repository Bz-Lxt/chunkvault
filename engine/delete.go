package engine

import (
	"context"

	"github.com/Bz-Lxt/chunkvault/digest"
	"github.com/Bz-Lxt/chunkvault/wal"
)

func (v *Vault) Unlink(ctx context.Context, d digest.Digest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if err := v.guard(ctx); err != nil {
		return err
	}
	if v.cfg.ReadOnly {
		return ErrReadOnly
	}
	m, err := v.db.GetBlob(ctx, d)
	if err != nil {
		return err
	}
	rec := wal.Record{Op: wal.OpUnlinkBlob, Digest: d}.Seal()
	if err := v.journal.Append(rec); err != nil {
		return err
	}
	for _, e := range m.Chunks {
		if _, err := v.db.AddRefs(ctx, e.Digest, -1); err != nil {
			return err
		}
		v.refs.Add(e.Digest, -1)
	}
	pins, err := v.db.ListPins(ctx)
	if err != nil {
		return err
	}
	for name, pd := range pins {
		if pd == d {
			if err := v.journal.Append(wal.Record{Op: wal.OpUnpin, Name: name}.Seal()); err != nil {
				return err
			}
			if err := v.db.Unpin(ctx, name); err != nil {
				return err
			}
		}
	}
	if err := v.db.DeleteBlob(ctx, d); err != nil {
		return err
	}
	return v.markApplied()
}

func (v *Vault) Checkpoint(ctx context.Context) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if err := v.guard(ctx); err != nil {
		return err
	}
	// 检查点只截断已落盘的 WAL：每条记录在 Put/Pin/Unlink 末尾都标记过已应用，
	// SQLite 里已持有全部对象清单与段数据。这里只能 Truncate + 重置已应用偏移，
	// 绝不能动 blobs/pins，否则按摘要读取会 404、丢掉已落库对象。
	if err := v.journal.Truncate(); err != nil {
		return err
	}
	return v.markApplied()
}
