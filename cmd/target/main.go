// Command target runs the deliberately-breakable HTTP server (the system under
// test). Configure it via -c <path-to-target.yaml>.
package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/adssib/dlt/internal/config"
	"github.com/adssib/dlt/internal/target"
)

func main() {
	configPath := flag.String("c", "configs/target.yaml", "path to target config file")
	flag.Parse()

	cfg, err := config.LoadTarget(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := target.NewServer(*cfg)
	log.Printf("target listening on %s", cfg.Target.Listen)
	if err := srv.Run(ctx); err != nil {
		log.Fatalf("server error: %v", err)
	}
	log.Println("target stopped")
}
