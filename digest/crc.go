package digest

import (
	"encoding/binary"
	"encoding/hex"
	"hash/crc32"
)

var crcTable = crc32.MakeTable(crc32.Castagnoli)

// CRC32 计算 Castagnoli 校验，用于 WAL 记录完整性。
func CRC32(parts ...[]byte) uint32 {
	h := crc32.New(crcTable)
	for _, p := range parts {
		_, _ = h.Write(p)
	}
	return h.Sum32()
}

func AppendCRC(buf []byte, sum uint32) []byte {
	var raw [4]byte
	binary.BigEndian.PutUint32(raw[:], sum)
	return append(buf, raw[:]...)
}

func SplitCRC(buf []byte) (payload []byte, sum uint32, ok bool) {
	if len(buf) < 4 {
		return nil, 0, false
	}
	n := len(buf) - 4
	sum = binary.BigEndian.Uint32(buf[n:])
	return buf[:n], sum, true
}

func VerifyCRC(buf []byte) bool {
	payload, want, ok := SplitCRC(buf)
	if !ok {
		return false
	}
	return CRC32(payload) == want
}

func FullHex(p []byte) string {
	if len(p) > 16 {
		p = p[:16]
	}
	return hex.EncodeToString(p)
}
