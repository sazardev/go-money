package extractor

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"regexp"
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

	// Check email domains
	for _, service := range te.tracker.Services {
		for _, domain := range service.EmailDomains {
			if strings.Contains(sender, strings.ToLower(domain)) {
				return &service
			}
		}
	}

	// Check keywords
	body := strings.ToLower(msg.Body + msg.Subject)
	for _, service := range te.tracker.Services {
		for _, keyword := range service.Keywords {
			if strings.Contains(body, strings.ToLower(keyword)) {
				return &service
			}
		}
	}

	return nil
}

// extractAmount extracts the amount from text
func (te *TransactionExtractor) extractAmount(text string) float64 {
	// Common patterns for amounts
	patterns := []string{
		`\$[\d,]+\.?\d{0,2}`,
		`USD\s*[\d,]+\.?\d{0,2}`,
		`Total[:\s]*\$?[\d,]+\.?\d{0,2}`,
		`Amount[:\s]*\$?[\d,]+\.?\d{0,2}`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(text, 1)
		if len(matches) > 0 {
			// TODO: Parse the amount properly
			return 0
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
