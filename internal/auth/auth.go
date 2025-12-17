package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

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

// requestNewToken initiates OAuth2 flow with automatic browser and code capture
func (a *Authenticator) requestNewToken(ctx context.Context) (*oauth2.Token, error) {
	// Start local HTTP server to capture the authorization code
	codeChan := make(chan string)
	errChan := make(chan error)

	// Create a listener on port 8080
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		a.log.Error(fmt.Sprintf("Failed to start local server: %v", err))
		return nil, err
	}

	// Start HTTP server in a goroutine
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Extract authorization code from URL
			code := r.URL.Query().Get("code")
			if code == "" {
				http.Error(w, "No authorization code received", http.StatusBadRequest)
				errChan <- fmt.Errorf("no authorization code received")
				return
			}

			// Send success response to browser
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head>
    <title>GO Money - Authorization Successful</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; margin-top: 50px; }
        .success { color: #27ae60; font-size: 24px; }
        .message { color: #555; margin-top: 20px; font-size: 16px; }
    </style>
</head>
<body>
    <div class="success">✅ Authorization Successful!</div>
    <p class="message">You can close this window and return to the terminal.</p>
    <p class="message">GO Money is now authenticated with your Google account.</p>
</body>
</html>
			`)

			// Send code to channel
			codeChan <- code
		})

		server := &http.Server{Handler: mux}
		server.Serve(listener)
	}()

	// Generate authorization URL
	authURL := a.oauth2Config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Printf("🔐 Opening browser for authentication...\n")
	fmt.Printf("📱 If browser doesn't open, visit: %s\n\n", authURL)

	// Try to open browser automatically
	openBrowser(authURL)

	// Wait for code or error
	select {
	case code := <-codeChan:
		listener.Close()
		a.log.Info("Authorization code received successfully")

		// Exchange code for token
		token, err := a.oauth2Config.Exchange(ctx, code)
		if err != nil {
			a.log.Error(fmt.Sprintf("Failed to exchange code: %v", err))
			return nil, err
		}

		// Save token to file
		err = a.saveTokenToFile(token)
		if err != nil {
			a.log.Error(fmt.Sprintf("Failed to save token: %v", err))
			return nil, err
		}

		return token, nil

	case err := <-errChan:
		listener.Close()
		return nil, err

	case <-ctx.Done():
		listener.Close()
		return nil, ctx.Err()
	}
}

// openBrowser opens the default browser with the given URL
func openBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		fmt.Printf("Unable to open browser on %s\n", runtime.GOOS)
		return
	}

	err := cmd.Start()
	if err != nil {
		fmt.Printf("Could not open browser: %v\n", err)
	}
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
