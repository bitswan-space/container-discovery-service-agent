package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"bitswan.space/container-discovery-service-agent/internal/config"
	"bitswan.space/container-discovery-service-agent/internal/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Define a command-line flag
	configPath := flag.String("c", "config.yaml", "path to the configuration file")
	flag.Parse() // Parse the flags

	logger.Init()
	godotenv.Load(".env")

	err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Error.Fatalf("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	<-sigChan
	logger.Info.Println("Shutting down gracefully...")
	// Perform any necessary cleanup here

	logger.Info.Println("Shutdown complete")
}
