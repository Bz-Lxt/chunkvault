package chunker

import (
	"fmt"

	"github.com/Bz-Lxt/chunkvault/digest"
)

// Piece 是一段独立寻址的窗口。Data 必须是调用方缓冲的副本。
type Piece struct {
	Digest digest.Digest
	Data   []byte
}

// Split 按固定窗口切开。空输入返回零段，不返回空段。
// 每段 Data 独立分配，调用方改原文不得污染已切出的段。
func Split(body []byte, pol Policy) ([]Piece, error) {
	pol, err := pol.Normalize()
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, nil
	}
	n := pol.Count(len(body))
	out := make([]Piece, 0, n)
	for off := 0; off < len(body); off += pol.Size {
		end := off + pol.Size
		if end > len(body) {
			end = len(body)
		}
		win := body[off:end]
		cp := make([]byte, len(win))
		copy(cp, win)
		out = append(out, Piece{Digest: digest.Sum(cp), Data: cp})
	}
	if len(out) != n {
		return nil, fmt.Errorf("split count %d != %d", len(out), n)
	}
	return out, nil
}
