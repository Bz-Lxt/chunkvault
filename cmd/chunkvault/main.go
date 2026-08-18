package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Bz-Lxt/chunkvault/config"
	"github.com/Bz-Lxt/chunkvault/engine"
	"github.com/Bz-Lxt/chunkvault/httpapi"
)

func main() {
	var (
		dir   = flag.String("data", "data", "vault directory")
		addr  = flag.String("addr", ":8080", "listen address")
		chunk = flag.Int("chunk", config.DefaultChunkSize, "chunk window")
		web   = flag.String("web", "web", "static ui directory")
	)
	flag.Parse()
	cfg, err := config.Normalize(config.FromEnv(config.Config{Dir: *dir, Addr: *addr, ChunkSize: *chunk, SlowChunkMS: 0}))
	if err != nil {
		log.Fatal(err)
	}
	v, err := engine.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer v.Close()
	srv := httpapi.New(v, cfg.Addr, *web)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()
	log.Printf("chunkvault listen %s data=%s", cfg.Addr, cfg.Dir)
	if err := srv.ListenAndServe(); err != nil {
		log.Print(err)
	}
}
