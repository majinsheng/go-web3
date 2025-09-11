package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/em/go-web3/internal/api"
	"github.com/em/go-web3/internal/config"
	"github.com/em/go-web3/internal/ethereum"
	"github.com/em/go-web3/internal/events"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create Ethereum client
	ethClient, err := ethereum.NewClient(&cfg.Ethereum)
	if err != nil {
		log.Fatalf("Failed to create Ethereum client: %v", err)
	}

	// Create event service
	eventService := events.NewService(ethClient.Client)
	if err := eventService.Start(); err != nil {
		log.Fatalf("Failed to start event service: %v", err)
	}

	// Create API handler
	handler := api.NewHandler(ethClient, eventService)

	// Create and start server
	server := api.NewServer(&cfg.Server, handler)

	// Handle graceful shutdown
	quit := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		// Stop the event service
		eventService.Stop()

		close(quit)
	}()

	go server.GracefulShutdown(quit)

	// Start server
	if err := server.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
