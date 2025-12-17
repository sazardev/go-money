package extractor

import (
	"encoding/json"
	"fmt"
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

	// Extract amount and currency
	amount, currency, currencySymbol, rawAmount := te.extractAmountWithCurrency(msg.Body)
	if amount <= 0 {
		return nil
	}

	// Try to extract transaction date from email body
	txDate := te.extractTransactionDate(msg.Body, msg.Subject)
	if txDate.IsZero() {
		txDate = msg.Date
	}

	// Create transaction
	txn := &models.Transaction{
		ID:             msg.ID,
		ServiceID:      service.ID,
		ServiceName:    service.Name,
		Category:       service.Category,
		Amount:         amount,
		Currency:       currency,
		CurrencySymbol: currencySymbol,
		Date:           txDate,
		Description:    msg.Subject,
		Email:          msg.From,
		Subject:        msg.Subject,
		Timestamp:      time.Now(),
		RawAmount:      rawAmount,
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

// extractAmountWithCurrency extracts amount AND currency from text
func (te *TransactionExtractor) extractAmountWithCurrency(text string) (float64, string, string, string) {
	if text == "" {
		return 0, "USD", "$", ""
	}

	// Pattern: match currency codes and symbols with amounts
	currencyPatterns := []struct {
		pattern  string // regex pattern
		currency string // currency code
		symbol   string // currency symbol
	}{
		{`(\$|\$\s*)[\s]*(\d[\d,]*\.?\d{0,2})\s*(USD)?`, "USD", "$"},
		{`(\d[\d,]*\.?\d{0,2})\s*(USD)`, "USD", "$"},
		{`(MXN|M\$|MEX|\$\s*M)\s*(\d[\d,]*\.?\d{0,2})`, "MXN", "$"},
		{`(\d[\d,]*\.?\d{0,2})\s*(MXN|M\$|MEX)`, "MXN", "$"},
		{`(€)\s*(\d[\d,]*\.?\d{0,2})`, "EUR", "€"},
		{`(\d[\d,]*\.?\d{0,2})\s*(EUR|€)`, "EUR", "€"},
		{`(£)\s*(\d[\d,]*\.?\d{0,2})`, "GBP", "£"},
		{`(\d[\d,]*\.?\d{0,2})\s*(GBP|£)`, "GBP", "£"},
		{`(¥|JPY)\s*(\d[\d,]*\.?\d{0,2})`, "JPY", "¥"},
		{`(\d[\d,]*\.?\d{0,2})\s*(JPY|¥)`, "JPY", "¥"},
		{`(CAD|\$\s*C)\s*(\d[\d,]*\.?\d{0,2})`, "CAD", "$"},
		{`(\d[\d,]*\.?\d{0,2})\s*(CAD)`, "CAD", "$"},
	}

	// Try each currency pattern
	for _, cp := range currencyPatterns {
		re := regexp.MustCompile("(?i)" + cp.pattern)
		matches := re.FindAllStringSubmatch(text, -1)

		var maxAmount float64
		var rawAmount string

		for _, match := range matches {
			var amountStr string

			// Extract the numeric part (usually second or third group depending on pattern)
			if len(match) >= 3 {
				// Try the last group (usually the number)
				for i := len(match) - 1; i >= 1; i-- {
					if match[i] != "" && !strings.ContainsAny(match[i], "$€£¥") {
						if num, err := strconv.Atoi(strings.ReplaceAll(match[i], ",", "")); err == nil && num > 0 {
							amountStr = match[i]
							break
						}
					}
				}
			}

			// Remove commas and parse
			amountStr = strings.ReplaceAll(amountStr, ",", "")
			if amount, err := strconv.ParseFloat(amountStr, 64); err == nil {
				if amount > maxAmount && amount < 1000000 { // Sanity check
					maxAmount = amount
					rawAmount = match[0]
				}
			}
		}

		if maxAmount > 0 {
			return maxAmount, cp.currency, cp.symbol, rawAmount
		}
	}

	// Fallback: use extractAmount without currency detection, default to USD
	amount := te.extractAmount(text)
	if amount > 0 {
		return amount, "USD", "$", ""
	}

	return 0, "USD", "$", ""
}

// extractAmount extracts the amount from text (fallback function)
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
				if amount > maxAmount && amount < 1000000 { // Sanity check
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

// cleanHTMLTags removes HTML tags and common HTML entities from text
func (te *TransactionExtractor) cleanHTMLTags(text string) string {
	// Remove HTML tags
	reHtmlTag := regexp.MustCompile("<[^>]*>")
	text = reHtmlTag.ReplaceAllString(text, " ")

	// Remove common HTML entities
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&apos;", "'")

	// Collapse multiple spaces
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	return text
}

// extractTransactionDate tries to extract the transaction date from email body and subject
func (te *TransactionExtractor) extractTransactionDate(body, subject string) time.Time {
	// Clean HTML from body
	cleanBody := te.cleanHTMLTags(body)
	fullText := cleanBody + " " + subject
	fullText = strings.ToLower(fullText)

	// Try exact date patterns first (YYYY-MM-DD, MM/DD/YYYY, etc.)
	datePatterns := []struct {
		pattern string
		format  string
	}{
		// Exact patterns with word boundaries
		{`\b(\d{4})-(\d{2})-(\d{2})\b`, "2006-01-02"},
		{`\b(\d{1,2})/(\d{1,2})/(\d{4})\b`, "01/02/2006"},
		// Month Day, Year format (Dec 14, 2025)
		{`\b(jan|january|feb|february|mar|march|apr|april|may|jun|june|jul|july|aug|august|sep|september|oct|october|nov|november|dec|december)\s+(\d{1,2}),?\s+(\d{4})\b`, "January 02 2006"},
		// Day Month Year format (14 Dec 2025)
		{`\b(\d{1,2})\s+(jan|january|feb|february|mar|march|apr|april|may|jun|june|jul|july|aug|august|sep|september|oct|october|nov|november|dec|december)\s+(\d{4})\b`, "02 January 2006"},
	}

	for _, dp := range datePatterns {
		re := regexp.MustCompile("(?i)" + dp.pattern)
		matches := re.FindAllStringSubmatch(fullText, -1)
		if len(matches) > 0 {
			// Get the last match (usually most relevant - actual transaction date, not email sent date)
			lastMatch := matches[len(matches)-1]

			// Extract and reconstruct the date string
			var dateStr string
			if len(lastMatch) >= 4 && strings.Contains(lastMatch[0], ",") {
				// Month Day, Year format (Dec 14, 2025)
				dateStr = strings.Title(lastMatch[1]) + " " + lastMatch[2] + " " + lastMatch[3]
			} else if len(lastMatch) >= 4 {
				// Day Month Year format (14 Dec 2025)
				dateStr = lastMatch[1] + " " + strings.Title(lastMatch[2]) + " " + lastMatch[3]
			} else if len(lastMatch) >= 3 {
				dateStr = lastMatch[1] + "-" + lastMatch[2] + "-" + lastMatch[3]
			} else {
				continue
			}

			// Try to parse
			if t, err := time.Parse(dp.format, dateStr); err == nil {
				return t
			}
		}
	}

	// Try month day pattern without year (use current year)
	today := time.Now()
	monthDayPattern := regexp.MustCompile(`(?i)\b(jan|january|feb|february|mar|march|apr|april|may|jun|june|jul|july|aug|august|sep|september|oct|october|nov|november|dec|december)\s+(\d{1,2})\b`)
	if matches := monthDayPattern.FindAllStringSubmatch(fullText, -1); len(matches) > 0 {
		lastMatch := matches[len(matches)-1]
		monthStr := strings.Title(lastMatch[1])
		dayStr := lastMatch[2]

		// Parse with current year
		dateStr := monthStr + " " + dayStr + " " + fmt.Sprintf("%d", today.Year())
		if t, err := time.Parse("January 02 2006", dateStr); err == nil {
			return t
		}
	}

	return time.Time{}
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
