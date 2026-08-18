package manifest

import (
	"encoding/binary"
	"fmt"

	"github.com/Bz-Lxt/chunkvault/digest"
)

const entryBytes = digest.Size + 4

// Encode 把清单编成稳定二进制：digest + size + count + entries。
func Encode(m Manifest) ([]byte, error) {
	if !m.Valid() {
		return nil, fmt.Errorf("invalid manifest")
	}
	buf := make([]byte, 0, digest.Size+8+4+len(m.Chunks)*entryBytes)
	buf = append(buf, m.Digest[:]...)
	var sz [8]byte
	binary.BigEndian.PutUint64(sz[:], uint64(m.Size))
	buf = append(buf, sz[:]...)
	var n [4]byte
	binary.BigEndian.PutUint32(n[:], uint32(len(m.Chunks)))
	buf = append(buf, n[:]...)
	for _, e := range m.Chunks {
		buf = append(buf, e.Digest[:]...)
		var es [4]byte
		binary.BigEndian.PutUint32(es[:], uint32(e.Size))
		buf = append(buf, es[:]...)
	}
	return digest.AppendCRC(buf, digest.CRC32(buf)), nil
}

// Decode 解析 Encode 的输出，校验 CRC 与尺寸和。
func Decode(raw []byte) (Manifest, error) {
	var m Manifest
	payload, want, ok := digest.SplitCRC(raw)
	if !ok {
		return m, fmt.Errorf("manifest too short")
	}
	if digest.CRC32(payload) != want {
		return m, fmt.Errorf("manifest crc mismatch")
	}
	if len(payload) < digest.Size+8+4 {
		return m, fmt.Errorf("manifest header truncated")
	}
	copy(m.Digest[:], payload[:digest.Size])
	m.Size = int64(binary.BigEndian.Uint64(payload[digest.Size : digest.Size+8]))
	count := int(binary.BigEndian.Uint32(payload[digest.Size+8 : digest.Size+12]))
	rest := payload[digest.Size+12:]
	if count < 0 || len(rest) != count*entryBytes {
		return m, fmt.Errorf("manifest entry count %d vs %d bytes", count, len(rest))
	}
	m.Chunks = make([]Entry, count)
	for i := 0; i < count; i++ {
		off := i * entryBytes
		copy(m.Chunks[i].Digest[:], rest[off:off+digest.Size])
		m.Chunks[i].Size = int(binary.BigEndian.Uint32(rest[off+digest.Size : off+entryBytes]))
	}
	if !m.Valid() {
		return m, fmt.Errorf("manifest failed validation")
	}
	return m, nil
}
