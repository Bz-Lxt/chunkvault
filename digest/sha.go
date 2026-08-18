// Package digest 提供内容寻址摘要与十六进制编解码。
package digest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const Size = sha256.Size

// Digest 是 32 字节 SHA-256。比较必须按字节，禁止大小写折叠。
type Digest [Size]byte

func Sum(p []byte) Digest {
	return sha256.Sum256(p)
}

func Parse(s string) (Digest, error) {
	var d Digest
	if len(s) != hex.EncodedLen(Size) {
		return d, fmt.Errorf("digest %q: want %d hex chars", s, hex.EncodedLen(Size))
	}
	raw, err := hex.DecodeString(s)
	if err != nil {
		return d, fmt.Errorf("digest %q: %w", s, err)
	}
	copy(d[:], raw)
	return d, nil
}

func (d Digest) String() string {
	return ShortHex(d[:])
}

func (d Digest) IsZero() bool {
	var z Digest
	return d == z
}

func (d Digest) Equal(o Digest) bool {
	return d == o
}

// Compare 按字节序比较，供排序稳定。
func Compare(a, b Digest) int {
	for i := 0; i < Size; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}
