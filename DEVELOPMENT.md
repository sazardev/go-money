# Development Guide

## Project Structure

```
go-money/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── cmd/                    # CLI commands (auth, calculate, graph, version)
│   ├── auth/                   # OAuth2 authentication with Google
│   ├── config/                 # Configuration management
│   ├── gmail/                  # Gmail API integration
│   ├── models/                 # Data models
│   └── extractor/              # Transaction extraction logic
├── pkg/
│   ├── logger/                 # Logging utilities
│   └── utils/                  # General utilities
├── tracker-mails.json          # Service configurations
├── go.mod                       # Go modules file
├── go.sum                       # Go modules checksums
├── Makefile                    # Build automation
└── README.md                   # Documentation
```

## Setup Instructions

### 1. Environment Setup

Create a `.env` file with your Google OAuth credentials:

```bash
GOOGLE_CLIENT_ID=your_client_id
GOOGLE_CLIENT_SECRET=your_client_secret
GOOGLE_PROJECT_ID=your_project_id
GOOGLE_AUTH_URI=https://accounts.google.com/o/oauth2/auth
GOOGLE_TOKEN_URI=https://oauth2.googleapis.com/token
GOOGLE_AUTH_PROVIDER_CERT_URL=https://www.googleapis.com/oauth2/v1/certs
GOOGLE_REDIRECT_URI=http://localhost
```

### 2. Install Dependencies

```bash
make install
# or
go mod download
go mod tidy
```

### 3. Build the Application

```bash
make build
# Binary will be in bin/gm
```

### 4. Run the Application

```bash
make run
# or
./bin/gm --help
```

## Available Commands

- `gm auth login` - Authenticate with Google
- `gm calculate` - Calculate and summarize expenses
- `gm graph` - Generate expense graph
- `gm version` - Show version
- `gm help` - Show help

## Development Workflow

### Code Organization

- **cmd/**: Application entry point
- **internal/cmd/**: Cobra CLI command definitions
- **internal/auth/**: OAuth2 token management
- **internal/config/**: Environment configuration
- **internal/gmail/**: Gmail API wrapper
- **internal/models/**: Data structures
- **internal/extractor/**: Transaction extraction engine
- **pkg/logger/**: Logging interface
- **pkg/utils/**: Helper functions

### Key Components

#### Authentication (internal/auth/auth.go)
- Handles OAuth2 flow with Google
- Manages token storage and refresh
- Provides HTTP client with authentication

#### Gmail Service (internal/gmail/gmail.go)
- Retrieves messages from Gmail
- Parses email headers and body
- Searches messages with queries

#### Transaction Extractor (internal/extractor/extractor.go)
- Loads service configurations from tracker-mails.json
- Matches emails to services
- Extracts transaction amounts

## Extending the Project

### Adding New Services

Edit `tracker-mails.json` and add a new service object:

```json
{
  "id": "service-id",
  "name": "Service Name",
  "category": "Category",
  "emailDomains": ["email@domain.com"],
  "transactionTypes": ["type1", "type2"],
  "keywords": ["keyword1", "keyword2"],
  "pricePattern": {
    "currency": "USD",
    "fields": ["field1", "field2"]
  }
}
```

### Adding New Commands

Create a new file in `internal/cmd/` and add it to the root command:

```go
var myCmd = &cobra.Command{
  Use: "mycommand",
  Short: "Description",
  RunE: func(cmd *cobra.Command, args []string) error {
    // Implementation
    return nil
  },
}

func init() {
  rootCmd.AddCommand(myCmd)
}
```

## Testing

```bash
make test
```

## Formatting and Linting

```bash
make fmt    # Format code
make lint   # Run linter
```

## Useful Go Commands

```bash
# View project structure
go list ./...

# Get dependency info
go mod graph

# Update dependencies
go get -u

# Run with race detection
go run -race ./cmd/main.go
```
