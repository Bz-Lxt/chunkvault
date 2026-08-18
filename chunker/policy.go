// Package chunker 把对象切成固定窗口，每段独立寻址。
package chunker

import "fmt"

const (
	MinSize = 32
	MaxSize = 1 << 20
)

// Policy 描述切段窗口。Size 是每段最大字节数。
type Policy struct {
	Size int
}

func DefaultPolicy() Policy {
	return Policy{Size: 4096}
}

func (p Policy) Normalize() (Policy, error) {
	if p.Size <= 0 {
		p.Size = 4096
	}
	if p.Size < MinSize {
		return p, fmt.Errorf("chunk size %d < %d", p.Size, MinSize)
	}
	if p.Size > MaxSize {
		return p, fmt.Errorf("chunk size %d > %d", p.Size, MaxSize)
	}
	return p, nil
}

func (p Policy) Count(n int) int {
	if n <= 0 {
		return 0
	}
	return (n + p.Size - 1) / p.Size
}
