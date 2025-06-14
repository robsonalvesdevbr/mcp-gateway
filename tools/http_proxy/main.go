package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"http-proxy/pkg"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Start listening as early as possible.
	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp4", ":8080")
	if err != nil {
		log.Fatalf("Failed to listen on port 8080: %v", err)
	}

	p := pkg.NewProxyServer(os.Getenv("ALLOWED_HOSTS"))
	if err := p.Run(ctx, ln); err != nil {
		log.Fatalf("Failed to run proxy: %v", err)
	}
}
