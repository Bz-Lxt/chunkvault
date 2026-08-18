package chunker

import (
	"fmt"

	"github.com/Bz-Lxt/chunkvault/digest"
)

// Join 按顺序拼回。段为 nil 视为损坏。
func Join(parts []Piece) ([]byte, error) {
	total := 0
	for i, p := range parts {
		if p.Data == nil {
			return nil, fmt.Errorf("chunk %d is nil", i)
		}
		if digest.Sum(p.Data) != p.Digest {
			return nil, fmt.Errorf("chunk %d digest mismatch", i)
		}
		total += len(p.Data)
	}
	out := make([]byte, 0, total)
	for _, p := range parts {
		out = append(out, p.Data...)
	}
	return out, nil
}
