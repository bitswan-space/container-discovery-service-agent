package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bitswan.space/container-discovery-service-agent/internal/config"
	"bitswan.space/container-discovery-service-agent/internal/docker"
	"bitswan.space/container-discovery-service-agent/internal/logger"
	"bitswan.space/container-discovery-service-agent/internal/mqtt"
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
	cfg := config.GetConfig()

	err = mqtt.Init()
	if err != nil {
		logger.Error.Fatalf("Failed to initialize MQTT client: %v", err)
		os.Exit(1)
	}

	err = docker.Init()
	if err != nil {
		logger.Error.Fatalf("Failed to initialize MQTT client: %v", err)
		os.Exit(1)
	}


	ticker := time.NewTicker(time.Duration(cfg.PollingInterval) * time.Second)
	defer ticker.Stop()


	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				docker.SendTopology()

			case <-sigChan:
				// Received an exit signal
				logger.Info.Println("Shutting down gracefully...")
				// Perform any necessary cleanup here
				mqtt.Close()
				docker.Close()
				logger.Info.Println("Shutdown complete")
				done <- true
				return
			}
		}
	}()

	// Block until shutdown signal is received
	<-done
}
