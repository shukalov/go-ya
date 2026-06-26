package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shukalov/go-ya/internal/agent"
)

func main() {
	addr := flag.String("a", "localhost:8080", "address endpoint")
	reportInterval := flag.Int("r", 10, "report interval in seconds")
	pollInterval := flag.Int("p", 2, "poll interval in seconds")

	flag.Parse()

	if *reportInterval <= 0 || *pollInterval <= 0 {
		fmt.Fprintf(os.Stderr, "intervals must be positive")
		os.Exit(1)
	}

	config := agent.Config{
		PollInterval:   time.Duration(*pollInterval) * time.Second,
		ReportInterval: time.Duration(*reportInterval) * time.Second,
		ServerAddress:  fmt.Sprintf("http://%s", *addr),
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a := agent.NewAgent(config)
	a.Run(ctx)
}
