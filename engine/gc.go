package engine

import (
	"context"

	"github.com/Bz-Lxt/chunkvault/gc"
)

func (v *Vault) Collect(ctx context.Context) (gc.Result, error) {
	// GC 必须复用 Put/Unlink 的 v.mu：标记与清扫期间禁止并发写入把零引用段
	// 重新挂回某个对象，否则刚写入的对象会因为段被清掉而读不回来。
	r := gc.Runner{DB: v.db, Lock: &v.mu}
	return r.Collect(context.Background())
}
