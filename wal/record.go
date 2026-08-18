// Package wal 预写日志：先落盘再改 SQLite。
package wal

import (
	"encoding/binary"
	"fmt"

	"github.com/Bz-Lxt/chunkvault/digest"
)

type Op uint8

const (
	OpPutChunk   Op = 1
	OpLinkBlob   Op = 2
	OpUnlinkBlob Op = 3
	OpPin        Op = 4
	OpUnpin      Op = 5
	OpSetMeta    Op = 6
)

func (o Op) String() string {
	switch o {
	case OpPutChunk:
		return "put_chunk"
	case OpLinkBlob:
		return "link_blob"
	case OpUnlinkBlob:
		return "unlink_blob"
	case OpPin:
		return "pin"
	case OpUnpin:
		return "unpin"
	case OpSetMeta:
		return "set_meta"
	default:
		return fmt.Sprintf("op_%d", o)
	}
}

// Record 一条完整日志。CRC 覆盖 Op+Digest+Name+Payload。
type Record struct {
	Op      Op
	Digest  digest.Digest
	Name    string
	Payload []byte
	CRC     uint32
}

func (r Record) crcParts() [][]byte {
	return [][]byte{
		{byte(r.Op)},
		r.Digest[:],
		[]byte(r.Name),
		r.Payload,
	}
}

func (r Record) ComputeCRC() uint32 {
	return digest.CRC32(r.crcParts()...)
}

func (r Record) Seal() Record {
	r.CRC = r.ComputeCRC()
	return r
}

func (r Record) Valid() bool {
	return r.CRC == r.ComputeCRC()
}

// Marshal 编码：u32len | op | digest | u16namelen | name | u32paylen | payload | crc32
func Marshal(r Record) []byte {
	r = r.Seal()
	name := []byte(r.Name)
	pay := r.Payload
	if pay == nil {
		pay = []byte{}
	}
	body := make([]byte, 0, 1+digest.Size+2+len(name)+4+len(pay)+4)
	body = append(body, byte(r.Op))
	body = append(body, r.Digest[:]...)
	var nl [2]byte
	binary.BigEndian.PutUint16(nl[:], uint16(len(name)))
	body = append(body, nl[:]...)
	body = append(body, name...)
	var pl [4]byte
	binary.BigEndian.PutUint32(pl[:], uint32(len(pay)))
	body = append(body, pl[:]...)
	body = append(body, pay...)
	var crc [4]byte
	binary.BigEndian.PutUint32(crc[:], r.CRC)
	body = append(body, crc[:]...)
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(body)))
	out := make([]byte, 0, 4+len(body))
	out = append(out, hdr[:]...)
	out = append(out, body...)
	return out
}

func Unmarshal(raw []byte) (Record, int, error) {
	var r Record
	if len(raw) < 4 {
		return r, 0, fmt.Errorf("wal header truncated")
	}
	n := int(binary.BigEndian.Uint32(raw[:4]))
	if n < 1+digest.Size+2+4+4 {
		return r, 0, fmt.Errorf("wal record too small: %d", n)
	}
	if len(raw) < 4+n {
		return r, 0, fmt.Errorf("wal record truncated: have %d want %d", len(raw)-4, n)
	}
	body := raw[4 : 4+n]
	r.Op = Op(body[0])
	copy(r.Digest[:], body[1:1+digest.Size])
	off := 1 + digest.Size
	nl := int(binary.BigEndian.Uint16(body[off : off+2]))
	off += 2
	if off+nl+4 > len(body)-4 {
		return r, 0, fmt.Errorf("wal name truncated")
	}
	r.Name = string(body[off : off+nl])
	off += nl
	pl := int(binary.BigEndian.Uint32(body[off : off+4]))
	off += 4
	if off+pl+4 != len(body) {
		return r, 0, fmt.Errorf("wal payload length mismatch")
	}
	if pl > 0 {
		r.Payload = append([]byte(nil), body[off:off+pl]...)
	}
	off += pl
	r.CRC = binary.BigEndian.Uint32(body[off:])
	if !r.Valid() {
		return r, 0, fmt.Errorf("wal crc mismatch op=%s", r.Op)
	}
	return r, 4 + n, nil
}
