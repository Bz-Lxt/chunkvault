// Package config 描述保险库打开参数。零值可用。
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Bz-Lxt/chunkvault/clock"
)

const (
	DefaultChunkSize = 4096
	DefaultAddr      = ":8080"
	SQLiteName       = "vault.sqlite"
	JournalName      = "journal.wal"
	WebDirName       = "web"
)

// Config 打开保险库与 HTTP 服务所需的全部旋钮。
type Config struct {
	Dir       string
	Addr      string
	ChunkSize int
	Clock       clock.Clock
	ReadOnly    bool
	SlowChunkMS int
}

func (c Config) withDefaults() Config {
	if c.Dir == "" {
		c.Dir = "data"
	}
	if c.Addr == "" {
		c.Addr = DefaultAddr
	}
	if c.ChunkSize <= 0 {
		c.ChunkSize = DefaultChunkSize
	}
	if c.Clock == nil {
		c.Clock = clock.Beijing{}
	}
	return c
}

// Normalize 校验并填默认值。
func Normalize(c Config) (Config, error) {
	c = c.withDefaults()
	if c.ChunkSize < 32 {
		return c, fmt.Errorf("chunk size %d too small", c.ChunkSize)
	}
	if c.ChunkSize > 1<<20 {
		return c, fmt.Errorf("chunk size %d too large", c.ChunkSize)
	}
	abs, err := filepath.Abs(c.Dir)
	if err != nil {
		return c, err
	}
	c.Dir = abs
	return c, nil
}

func (c Config) SQLitePath() string {
	return filepath.Join(c.Dir, SQLiteName)
}

func (c Config) JournalPath() string {
	return filepath.Join(c.Dir, JournalName)
}

func (c Config) WebDir() string {
	if _, err := os.Stat(filepath.Join(c.Dir, WebDirName)); err == nil {
		return filepath.Join(c.Dir, WebDirName)
	}
	return WebDirName
}

// FromEnv 读 CHUNKVAULT_* 环境变量覆盖字段。
func FromEnv(c Config) Config {
	if v := os.Getenv("CHUNKVAULT_DIR"); v != "" {
		c.Dir = v
	}
	if v := os.Getenv("CHUNKVAULT_ADDR"); v != "" {
		c.Addr = v
	}
	if v := os.Getenv("CHUNKVAULT_CHUNK"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.ChunkSize = n
		}
	}
	if v := os.Getenv("CHUNKVAULT_SLOW_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.SlowChunkMS = n
		}
	}
	return c
}
