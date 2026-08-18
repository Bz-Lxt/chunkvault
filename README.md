# ChunkVault

内容寻址分段保险库。对象按固定窗口切开，每段用 SHA-256 寻址，引用计数归零后才能被世代 GC 清扫。写入先追加 `journal.wal` 再进 SQLite，进程重启后按已应用偏移重放未检查点的日志。

## 启动

```bash
docker compose up --build -d
```

浏览器打开 `http://127.0.0.1:18081/`，健康检查 `http://127.0.0.1:18081/health`。

本地：

```bash
go test ./...
go run ./cmd/chunkvault -data ./data -addr :8080
```
