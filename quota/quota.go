// Package quota 按已占用字节限制新写入，避免保险库无限膨胀。
package quota

import "fmt"

// Limit 描述对象总字节上限。Zero 表示不限制。
type Limit struct {
	MaxBytes int64
}

func (l Limit) Enabled() bool { return l.MaxBytes > 0 }

func (l Limit) Allow(used, incoming int64) error {
	if !l.Enabled() {
		return nil
	}
	if incoming < 0 {
		return fmt.Errorf("quota: negative incoming")
	}
	if used < 0 {
		return fmt.Errorf("quota: negative used")
	}
	if used+incoming > l.MaxBytes {
		return fmt.Errorf("quota: %d+%d exceeds %d", used, incoming, l.MaxBytes)
	}
	return nil
}

func (l Limit) Remaining(used int64) int64 {
	if !l.Enabled() {
		return -1
	}
	if used >= l.MaxBytes {
		return 0
	}
	return l.MaxBytes - used
}

// Headroom 按窗口估算还能再写几段。
func (l Limit) Headroom(used int64, chunk int) int {
	left := l.Remaining(used)
	if left < 0 {
		return -1
	}
	if chunk <= 0 {
		return 0
	}
	return int(left / int64(chunk))
}
