# GO Money - CLI Project Setup

## Quick Start

### Prerequisites
- Go 1.25.5 or higher
- Google OAuth2 credentials
- Make (optional, for using Makefile)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd go-money
   ```

2. **Install dependencies**
   ```bash
   make install
   # or
   go mod download
   go mod tidy
   ```

3. **Configure Google OAuth**
   - Create a Google Cloud Project
   - Enable Gmail API
   - Create OAuth2 credentials (Desktop application)
   - Copy your credentials to `.env`

4. **Build the application**
   ```bash
   make build
   # Binary will be in bin/gm
   ```

5. **Run the application**
   ```bash
   make run
   ```

## Project Structure

See [DEVELOPMENT.md](DEVELOPMENT.md) for detailed project structure and development guidelines.

## Key Features

✅ OAuth2 authentication with Google  
✅ Gmail API integration  
✅ CLI command structure with Cobra  
✅ Service tracker configuration  
✅ Transaction extraction foundation  
✅ Modular architecture  
✅ Logging system  
✅ Configuration management  

## Available Commands

```bash
gm auth login      # Authenticate with Google
gm calculate       # Extract and summarize expenses
gm graph          # Generate expense visualization
gm version        # Show version
gm help           # Show help
```

## Development

### Build Commands

```bash
make build        # Build the binary
make run          # Run the application
make test         # Run tests
make clean        # Clean build artifacts
make fmt          # Format code
make lint         # Run linter
```

### Project Layout

```
├── cmd/                          # Entry point
├── internal/
│   ├── cmd/                      # CLI commands (Cobra)
│   ├── auth/                     # OAuth2 authentication
│   ├── config/                   # Configuration
│   ├── gmail/                    # Gmail API wrapper
│   ├── models/                   # Data models
│   └── extractor/                # Transaction extraction
├── pkg/
│   ├── logger/                   # Logging utilities
│   └── utils/                    # Helper functions
├── tracker-mails.json            # Service configurations
└── Makefile                      # Build automation
```

## Dependencies

- **github.com/spf13/cobra**: CLI framework
- **github.com/joho/godotenv**: Environment variables
- **google.golang.org/api**: Google API client
- **golang.org/x/oauth2**: OAuth2 implementation
- **github.com/go-echarts/go-echarts**: Charting library

## Configuration

Create a `.env` file with your Google OAuth credentials:

```env
GOOGLE_CLIENT_ID=your_client_id
GOOGLE_CLIENT_SECRET=your_client_secret
GOOGLE_PROJECT_ID=your_project_id
GOOGLE_AUTH_URI=https://accounts.google.com/o/oauth2/auth
GOOGLE_TOKEN_URI=https://oauth2.googleapis.com/token
GOOGLE_AUTH_PROVIDER_CERT_URL=https://www.googleapis.com/oauth2/v1/certs
GOOGLE_REDIRECT_URI=http://localhost
```

## Service Configuration

The `tracker-mails.json` file contains all service configurations including:
- Email domains for each service
- Transaction types
- Keywords for matching
- Price patterns and currencies

Add new services by editing this file.

## Next Steps

1. Review [DEVELOPMENT.md](DEVELOPMENT.md) for detailed development guide
2. Implement specific email parsing logic for each service
3. Complete the transaction extraction engine
4. Add data persistence and storage
5. Implement the graphing functionality
6. Add comprehensive error handling and logging
7. Write tests for each module

## Troubleshooting

### OAuth Authentication Issues
- Ensure Google OAuth credentials are correct in `.env`
- Check that the redirect URI matches your configuration
- Clear `.credentials/token.json` to force re-authentication

### Missing Dependencies
```bash
go mod tidy
go mod download
```

### Build Errors
```bash
go clean -modcache
go mod download
go mod tidy
```

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Google API Client Library for Go](https://github.com/googleapis/google-api-go-client)
- [Cobra CLI Framework](https://cobra.dev/)
- [Gmail API Documentation](https://developers.google.com/gmail/api/guides)

## License

This project is open source and available under the MIT License.
