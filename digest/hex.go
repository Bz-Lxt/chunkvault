package digest

import (
	"encoding/hex"
	"strings"
)

func Encode(d Digest) string {
	return hex.EncodeToString(d[:])
}

func Decode(s string) (Digest, error) {
	return Parse(strings.TrimSpace(s))
}

func EqualHex(a, b string) bool {
	left, err1 := Parse(strings.TrimSpace(a))
	right, err2 := Parse(strings.TrimSpace(b))
	if err1 != nil || err2 != nil {
		return false
	}
	return left == right
}
