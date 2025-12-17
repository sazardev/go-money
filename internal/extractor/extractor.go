package extractor

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sazardev/go-money/internal/models"
)

type ServiceTracker struct {
	Services map[string]Service `json:"services"`
}

type Service struct {
	ID               string             `json:"id"`
	Name             string             `json:"name"`
	Category         string             `json:"category"`
	EmailDomains     []string           `json:"emailDomains"`
	TransactionTypes []string           `json:"transactionTypes"`
	Keywords         []string           `json:"keywords"`
	PricePattern     PricePatternConfig `json:"pricePattern"`
}

type PricePatternConfig struct {
	Currency string   `json:"currency"`
	Fields   []string `json:"fields"`
}

// TransactionExtractor handles extraction of transactions from emails
type TransactionExtractor struct {
	tracker *ServiceTracker
}

// NewTransactionExtractor creates a new extractor
func NewTransactionExtractor() (*TransactionExtractor, error) {
	tracker, err := loadServiceTracker()
	if err != nil {
		return nil, err
	}

	return &TransactionExtractor{
		tracker: tracker,
	}, nil
}

// loadServiceTracker loads the service configuration from tracker-mails.json
func loadServiceTracker() (*ServiceTracker, error) {
	data, err := ioutil.ReadFile("tracker-mails.json")
	if err != nil {
		log.Fatalf("Failed to load tracker-mails.json: %v", err)
		return nil, err
	}

	var trackerData struct {
		Services []Service `json:"services"`
	}

	if err := json.Unmarshal(data, &trackerData); err != nil {
		return nil, err
	}

	// Convert to map
	tracker := &ServiceTracker{
		Services: make(map[string]Service),
	}
	for _, service := range trackerData.Services {
		tracker.Services[service.ID] = service
	}

	return tracker, nil
}

// ExtractTransactions extracts transactions from messages
func (te *TransactionExtractor) ExtractTransactions(messages []*models.Message) []*models.Transaction {
	var transactions []*models.Transaction

	for _, msg := range messages {
		if txn := te.extractTransactionFromMessage(msg); txn != nil {
			transactions = append(transactions, txn)
		}
	}

	return transactions
}

// extractTransactionFromMessage extracts transaction from a single message
func (te *TransactionExtractor) extractTransactionFromMessage(msg *models.Message) *models.Transaction {
	// Check email domain
	service := te.matchService(msg)
	if service == nil {
		return nil
	}

	// Extract amount
	amount := te.extractAmount(msg.Body)
	if amount <= 0 {
		return nil
	}

	// Create transaction
	txn := &models.Transaction{
		ID:          msg.ID,
		ServiceID:   service.ID,
		ServiceName: service.Name,
		Category:    service.Category,
		Amount:      amount,
		Currency:    service.PricePattern.Currency,
		Date:        msg.Date,
		Description: msg.Subject,
		Email:       msg.From,
		Subject:     msg.Subject,
		Timestamp:   time.Now(),
	}

	return txn
}

// matchService finds the matching service for a message
func (te *TransactionExtractor) matchService(msg *models.Message) *Service {
	sender := strings.ToLower(msg.From)
	body := strings.ToLower(msg.Body + " " + msg.Subject)

	// Priority 1: Check email domains (most reliable)
	for _, service := range te.tracker.Services {
		for _, domain := range service.EmailDomains {
			if strings.Contains(sender, strings.ToLower(domain)) {
				// Found match by email domain
				return &service
			}
		}
	}

	// Priority 2: Check keywords
	for _, service := range te.tracker.Services {
		for _, keyword := range service.Keywords {
			if strings.Contains(body, strings.ToLower(keyword)) {
				// Found match by keyword
				return &service
			}
		}
	}

	return nil
}

// extractAmount extracts the amount from text with multiple patterns
func (te *TransactionExtractor) extractAmount(text string) float64 {
	if text == "" {
		return 0
	}

	// Common patterns for monetary amounts
	patterns := []string{
		// Dollar amounts: $123.45
		`\$\s*([\d,]+\.?\d{0,2})`,
		// USD amounts: USD 123.45
		`USD\s+([\d,]+\.?\d{0,2})`,
		// Total field: Total: $123.45 or Total: 123.45
		`(?i)total\s*:?\s*\$?\s*([\d,]+\.?\d{0,2})`,
		// Amount field: Amount: $123.45
		`(?i)amount\s*:?\s*\$?\s*([\d,]+\.?\d{0,2})`,
		// Charge field: Charge: $123.45
		`(?i)charge\s*:?\s*\$?\s*([\d,]+\.?\d{0,2})`,
		// Price field: Price: $123.45
		`(?i)price\s*:?\s*\$?\s*([\d,]+\.?\d{0,2})`,
		// Generic number pattern with currency symbol
		`[\$£€]\s*([\d,]+\.?\d{0,2})`,
		// Generic number at end of likely currency string
		`[\d,]+\.\d{2}\s*(USD|EUR|GBP)`,
	}

	// Try each pattern
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(text, -1)

		// Get the largest match (in case of multiple amounts, pick the biggest one)
		var maxAmount float64
		for _, match := range matches {
			if len(match) >= 2 {
				// Extract the number group
				amountStr := match[1]
				// Remove commas
				amountStr = strings.ReplaceAll(amountStr, ",", "")

				if amount, err := strconv.ParseFloat(amountStr, 64); err == nil {
					if amount > maxAmount {
						maxAmount = amount
					}
				}
			} else if len(match) >= 1 {
				// Try parsing the whole match
				amountStr := match[0]
				// Remove currency symbols
				amountStr = strings.TrimPrefix(amountStr, "$")
				amountStr = strings.TrimPrefix(amountStr, "£")
				amountStr = strings.TrimPrefix(amountStr, "€")
				amountStr = strings.TrimSpace(amountStr)
				amountStr = strings.ReplaceAll(amountStr, ",", "")

				if amount, err := strconv.ParseFloat(amountStr, 64); err == nil {
					if amount > maxAmount {
						maxAmount = amount
					}
				}
			}
		}

		if maxAmount > 0 {
			return maxAmount
		}
	}

	// Fallback: find any number that looks like a price
	// Match any number with 2 decimal places
	re := regexp.MustCompile(`\d+\.\d{2}`)
	matches := re.FindAllString(text, -1)
	if len(matches) > 0 {
		// Get the last match (often amounts are listed last in receipts)
		for i := len(matches) - 1; i >= 0; i-- {
			if amount, err := strconv.ParseFloat(matches[i], 64); err == nil && amount > 0 {
				return amount
			}
		}
	}

	// Last resort: find largest number
	re = regexp.MustCompile(`[\d,]+\.?\d{0,2}`)
	matches = re.FindAllString(text, -1)
	if len(matches) > 0 {
		var maxAmount float64
		for _, match := range matches {
			match = strings.ReplaceAll(match, ",", "")
			if amount, err := strconv.ParseFloat(match, 64); err == nil {
				if amount > maxAmount && amount < 100000 { // Sanity check
					maxAmount = amount
				}
			}
		}
		if maxAmount > 0 {
			return maxAmount
		}
	}

	return 0
}

// GetServiceByID returns a service by its ID
func (te *TransactionExtractor) GetServiceByID(id string) *Service {
	if service, ok := te.tracker.Services[id]; ok {
		return &service
	}
	return nil
}

// GetAllServices returns all services
func (te *TransactionExtractor) GetAllServices() []Service {
	var services []Service
	for _, service := range te.tracker.Services {
		services = append(services, service)
	}
	return services
}
