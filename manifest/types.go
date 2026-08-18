// Package manifest 描述对象由哪些段按序组成。
package manifest

import "github.com/Bz-Lxt/chunkvault/digest"

// Entry 是清单里的一段。
type Entry struct {
	Digest digest.Digest
	Size   int
}

// Manifest 是对象级清单。Digest 是原文 SHA-256，不是清单编码的摘要。
type Manifest struct {
	Digest digest.Digest
	Size   int64
	Chunks []Entry
}

func (m Manifest) ChunkCount() int { return len(m.Chunks) }

func (m Manifest) SumSizes() int64 {
	var n int64
	for _, e := range m.Chunks {
		n += int64(e.Size)
	}
	return n
}

func (m Manifest) Valid() bool {
	if m.Digest.IsZero() && m.Size != 0 {
		return false
	}
	if m.Size != m.SumSizes() {
		return false
	}
	for _, e := range m.Chunks {
		if e.Size < 0 || e.Digest.IsZero() {
			return false
		}
	}
	return true
}

// Clone 深拷贝，调用方改 Chunks 不得污染原清单。
func (m Manifest) Clone() Manifest {
	cp := m
	if m.Chunks != nil {
		cp.Chunks = make([]Entry, len(m.Chunks))
		copy(cp.Chunks, m.Chunks)
	}
	return cp
}
