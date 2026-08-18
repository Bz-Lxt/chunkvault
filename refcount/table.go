// Package refcount 在内存里跟踪本轮会话的引用增减，便于 GC 与诊断。
package refcount

import (
	"sync"

	"github.com/Bz-Lxt/chunkvault/digest"
)

// Table 记录 digest -> delta。零值不可用，必须 New。
type Table struct {
	mu   sync.Mutex
	refs map[digest.Digest]int64
}

func New() *Table {
	return &Table{refs: map[digest.Digest]int64{}}
}

func (t *Table) Add(d digest.Digest, delta int64) int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	next := t.refs[d] + delta
	if next < 0 {
		next = 0
	}
	t.refs[d] = next
	return next
}

func (t *Table) Get(d digest.Digest) int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.refs[d]
}

func (t *Table) Snapshot() map[digest.Digest]int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make(map[digest.Digest]int64, len(t.refs))
	for k, v := range t.refs {
		out[k] = v
	}
	return out
}

func (t *Table) Live() []digest.Digest {
	t.mu.Lock()
	defer t.mu.Unlock()
	var out []digest.Digest
	for k, v := range t.refs {
		if v > 0 {
			out = append(out, k)
		}
	}
	return out
}
