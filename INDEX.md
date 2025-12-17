# GO Money - Project Foundation Index

## 📋 Documentation

- [README.md](README.md) - Project overview and features
- [SETUP.md](SETUP.md) - Installation and quick start guide
- [DEVELOPMENT.md](DEVELOPMENT.md) - Development guide and architecture

## 🏗️ Project Structure

### Entry Point
- **[cmd/main.go](cmd/main.go)** - Application entry point, loads environment variables

### Internal Packages

#### CLI Commands (`internal/cmd/`)
- **[root.go](internal/cmd/root.go)** - Root command and CLI setup (Cobra)
- **[auth.go](internal/cmd/auth.go)** - OAuth2 authentication command
- **[calculate.go](internal/cmd/calculate.go)** - Expense calculation command
- **[graph.go](internal/cmd/graph.go)** - Graph generation command
- **[version.go](internal/cmd/version.go)** - Version display command

#### Authentication (`internal/auth/`)
- **[auth.go](internal/auth/auth.go)** - OAuth2 token management, browser-based login

#### Gmail Integration (`internal/gmail/`)
- **[gmail.go](internal/gmail/gmail.go)** - Gmail API client, message retrieval

#### Configuration (`internal/config/`)
- **[config.go](internal/config/config.go)** - Environment variable management

#### Data Models (`internal/models/`)
- **[models.go](internal/models/models.go)** - Transaction, Message, ExpenseSummary models
- **[time.go](internal/models/time.go)** - Time placeholder (WIP)

#### Transaction Extraction (`internal/extractor/`)
- **[extractor.go](internal/extractor/extractor.go)** - Service matching, amount extraction

### Public Packages

#### Logging (`pkg/logger/`)
- **[logger.go](pkg/logger/logger.go)** - Logger interface and implementation

#### Utilities (`pkg/utils/`)
- **[utils.go](pkg/utils/utils.go)** - Base64 decoding, email parsing, amount extraction

## 📦 Key Dependencies

```go
github.com/spf13/cobra              // CLI framework
github.com/joho/godotenv            // Environment loading
google.golang.org/api               // Google API client
golang.org/x/oauth2                 // OAuth2 implementation
github.com/go-echarts/go-echarts    // Chart generation
```

## 🔧 Configuration Files

- **[go.mod](go.mod)** - Go module definition with dependencies
- **[.env.example](.env.example)** - Example environment variables
- **[.gitignore](.gitignore)** - Git ignore patterns
- **[Makefile](Makefile)** - Build automation commands
- **[tracker-mails.json](tracker-mails.json)** - Service configurations (51 services)

## 🚀 Quick Commands

```bash
make install    # Install dependencies
make build      # Build binary
make run        # Run application
make test       # Run tests
make clean      # Clean build artifacts
make fmt        # Format code
```

## 📝 Available CLI Commands

```bash
gm auth login       # Authenticate with Google
gm calculate        # Extract and summarize expenses
gm graph           # Generate visualization
gm version         # Show version
gm help            # Show help
```

## 🎯 Architecture Overview

```
┌─────────────────────────────────────┐
│      CLI Entry (Cobra Commands)     │
│  auth | calculate | graph | version │
└──────────────┬──────────────────────┘
               │
       ┌───────┴────────┐
       │                │
   ┌───▼────┐      ┌────▼────────┐
   │  Auth  │      │Gmail Service │
   │(OAuth) │      │   (API)      │
   └────────┘      └──────┬───────┘
                          │
                    ┌─────▼────────┐
                    │  Extractor   │
                    │ (Service ID) │
                    └──────┬───────┘
                           │
                      ┌────▼─────┐
                      │ Models &  │
                      │ Logger    │
                      └───────────┘
```

## 📚 Module Dependencies

```
cmd/main.go
  └── internal/cmd (Cobra CLI)
      ├── internal/auth (OAuth)
      ├── internal/gmail (API)
      ├── internal/extractor (Extraction)
      ├── internal/config (Config)
      └── pkg/logger (Logging)

internal/gmail
  └── golang.org/x/oauth2 (OAuth client)
  └── google.golang.org/api (Gmail API)

internal/extractor
  └── tracker-mails.json (Service config)
  └── pkg/utils (Helper functions)
```

## 🔐 Security Notes

- OAuth2 tokens stored in `.credentials/` (add to .gitignore)
- Environment variables loaded from `.env` (not versioned)
- Never commit sensitive credentials

## 🎓 Next Development Steps

1. **Email Parsing**
   - Implement proper base64 decoding
   - Parse email headers and body
   - Extract structured data

2. **Transaction Extraction**
   - Complete amount extraction logic
   - Implement service matching
   - Add regex patterns for common formats

3. **Data Persistence**
   - Add database (SQLite/PostgreSQL)
   - Implement transaction storage
   - Add query capabilities

4. **Visualization**
   - Implement Go Echarts integration
   - Create expense charts
   - Add time-series analysis

5. **Testing**
   - Unit tests for each module
   - Integration tests
   - Mock Gmail API responses

6. **Enhancements**
   - Add CSV export
   - Implement caching
   - Add configuration file support
   - Multi-language support

## 📞 Support

Refer to [SETUP.md](SETUP.md) for troubleshooting and [DEVELOPMENT.md](DEVELOPMENT.md) for architectural details.

---

**Project Foundation Status**: ✅ Complete  
**Total Services Tracked**: 51  
**Project Categories**: 9  
**Ready for Development**: Yes
