package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/sazardev/go-money/internal/cmd"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Execute root command
	if err := cmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
		os.Exit(1)
	}
}
