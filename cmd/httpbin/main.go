package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TykTechnologies/tyk-devops-assignement/internal/server"
)

// Build information - set via ldflags during build
var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

// printVersion prints version information
func printVersion() {
	fmt.Printf("\tVersion:      %s\n", version)
	fmt.Printf("\tCommit:       %s\n", commit)
	fmt.Printf("\tBuild time:   %s\n", buildTime)
}

func main() {
	// Parse command-line flags
	host := flag.String("host", "0.0.0.0", "Host to bind the server to")
	port := flag.Int("port", 8088, "Port to bind the server to")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	// Show version and exit if requested
	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	// Create server
	addr := fmt.Sprintf("%s:%d", *host, *port)
	srv := server.New(addr)

	// Start server in a goroutine
	go func() {
		log.Printf("Starting httpbin server on %s (version: %s, commit: %s)", addr, version, commit)
		if err := srv.Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
