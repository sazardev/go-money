package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sazardev/go-money/internal/config"
	"github.com/sazardev/go-money/pkg/logger"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Authenticator struct {
	config       *config.Config
	oauth2Config *oauth2.Config
	log          logger.Logger
}

// NewAuthenticator creates a new Authenticator instance
func NewAuthenticator() *Authenticator {
	log := logger.GetLogger()
	cfg := config.LoadConfig()

	oauthConfig := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURI,
		Scopes: []string{
			"https://www.googleapis.com/auth/gmail.readonly",
		},
		Endpoint: google.Endpoint,
	}

	return &Authenticator{
		config:       cfg,
		oauth2Config: oauthConfig,
		log:          log,
	}
}

// GetToken retrieves a valid OAuth2 token
func (a *Authenticator) GetToken(ctx context.Context) (*oauth2.Token, error) {
	// Try to load from file first
	token, err := a.loadTokenFromFile()
	if err == nil && token.Valid() {
		a.log.Info("Using cached token")
		return token, nil
	}

	// If no valid token, request a new one
	a.log.Info("Requesting new token from user...")
	return a.requestNewToken(ctx)
}

// requestNewToken initiates OAuth2 flow
func (a *Authenticator) requestNewToken(ctx context.Context) (*oauth2.Token, error) {
	// Generate authorization code
	authURL := a.oauth2Config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser:\n%v\n", authURL)

	// Wait for authorization code
	var authCode string
	fmt.Print("Enter authorization code: ")
	fmt.Scanln(&authCode)

	// Exchange code for token
	token, err := a.oauth2Config.Exchange(ctx, authCode)
	if err != nil {
		a.log.Error(fmt.Sprintf("Unable to retrieve token: %v", err))
		return nil, err
	}

	// Save token to file
	if err := a.saveTokenToFile(token); err != nil {
		a.log.Warn(fmt.Sprintf("Unable to save token: %v", err))
	}

	return token, nil
}

// saveTokenToFile saves the OAuth2 token to a file
func (a *Authenticator) saveTokenToFile(token *oauth2.Token) error {
	credDir := ".credentials"
	if err := os.MkdirAll(credDir, 0700); err != nil {
		return err
	}

	tokFile := filepath.Join(credDir, "token.json")
	f, err := os.OpenFile(tokFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(token)
}

// loadTokenFromFile loads the OAuth2 token from file
func (a *Authenticator) loadTokenFromFile() (*oauth2.Token, error) {
	tokFile := filepath.Join(".credentials", "token.json")
	b, err := ioutil.ReadFile(tokFile)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(b, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// GetHTTPClient returns an HTTP client with the OAuth2 token
func (a *Authenticator) GetHTTPClient(ctx context.Context, token *oauth2.Token) *http.Client {
	return a.oauth2Config.Client(ctx, token)
}
