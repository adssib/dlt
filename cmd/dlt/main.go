// Command dlt is the distributed load tester: one binary with role subcommands
// (coordinator | worker) plus a local launcher (test). See ADR-0001.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/adssib/dlt/internal/config"
	"github.com/adssib/dlt/internal/coordinator"
	"github.com/adssib/dlt/internal/launcher"
	"github.com/adssib/dlt/internal/worker"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "coordinator":
		runCoordinator(os.Args[2:])
	case "worker":
		runWorker(os.Args[2:])
	case "test":
		runTest(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `dlt — distributed load tester

usage:
  dlt coordinator --config configs/coordinator.yaml
  dlt worker      --config configs/worker.yaml
  dlt test        --coordinator-config configs/coordinator.yaml --worker-config configs/worker.yaml --workers N
`)
}

func runCoordinator(args []string) {
	fs := flag.NewFlagSet("coordinator", flag.ExitOnError)
	var cfgPath string
	fs.StringVar(&cfgPath, "config", "configs/coordinator.yaml", "path to coordinator config")
	fs.StringVar(&cfgPath, "c", "configs/coordinator.yaml", "shorthand for --config")
	_ = fs.Parse(args)

	cfg, err := config.LoadCoordinator(cfgPath)
	if err != nil {
		fatal(err)
	}
	ctx, stop := signalContext()
	defer stop()
	if err := coordinator.New(cfg).Run(ctx); err != nil {
		fatal(err)
	}
}

func runWorker(args []string) {
	fs := flag.NewFlagSet("worker", flag.ExitOnError)
	var cfgPath string
	fs.StringVar(&cfgPath, "config", "configs/worker.yaml", "path to worker config")
	fs.StringVar(&cfgPath, "c", "configs/worker.yaml", "shorthand for --config")
	_ = fs.Parse(args)

	cfg, err := config.LoadWorker(cfgPath)
	if err != nil {
		fatal(err)
	}
	ctx, stop := signalContext()
	defer stop()
	if err := worker.New(cfg).Run(ctx); err != nil {
		fatal(err)
	}
}

func runTest(args []string) {
	fs := flag.NewFlagSet("test", flag.ExitOnError)
	coordCfg := fs.String("coordinator-config", "configs/coordinator.yaml", "path to coordinator config")
	workerCfg := fs.String("worker-config", "configs/worker.yaml", "path to worker config")
	workers := fs.Int("workers", 1, "number of local workers to spawn")
	_ = fs.Parse(args)

	ctx, stop := signalContext()
	defer stop()
	if err := launcher.Run(ctx, *coordCfg, *workerCfg, *workers); err != nil {
		fatal(err)
	}
}

// signalContext returns a context cancelled on SIGINT / SIGTERM.
func signalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
