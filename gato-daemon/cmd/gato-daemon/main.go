package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/veinticinco/gato-daemon/internal/folders"
)

const version = "0.1.0"

func main() {
	showVersion := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("gato-daemon v%s\n", version)
		fmt.Println("The intelligent daemon for Gato OS")
		os.Exit(0)
	}

	log.Printf("Starting gato-daemon v%s", version)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start folder manager
	mgr := folders.New()

	go func() {
		if err := mgr.Start(ctx); err != nil {
			log.Printf("Folder manager error: %v", err)
		}
	}()

	log.Println("Gato daemon running. Press Ctrl+C to stop.")

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	cancel()
}
