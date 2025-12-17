package config

import (
	"os"

	"github.com/sazardev/go-money/pkg/logger"
)

type Config struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleProjectID    string
	GoogleAuthURI      string
	GoogleTokenURI     string
	GoogleRedirectURI  string
	TokenFile          string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	log := logger.GetLogger()

	config := &Config{
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleProjectID:    os.Getenv("GOOGLE_PROJECT_ID"),
		GoogleAuthURI:      os.Getenv("GOOGLE_AUTH_URI"),
		GoogleTokenURI:     os.Getenv("GOOGLE_TOKEN_URI"),
		GoogleRedirectURI:  os.Getenv("GOOGLE_REDIRECT_URI"),
		TokenFile:          ".credentials/token.json",
	}

	// Validate required fields
	if config.GoogleClientID == "" || config.GoogleClientSecret == "" {
		log.Warn("Missing Google OAuth credentials. Please set GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET")
	}

	return config
}

// IsValid checks if the configuration is valid
func (c *Config) IsValid() bool {
	return c.GoogleClientID != "" && c.GoogleClientSecret != ""
}
