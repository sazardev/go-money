package models

import "time"

// Transaction represents a financial transaction
type Transaction struct {
	ID             string    `json:"id"`
	ServiceID      string    `json:"service_id"`
	ServiceName    string    `json:"service_name"`
	Category       string    `json:"category"`
	Amount         float64   `json:"amount"`
	Currency       string    `json:"currency"`        // USD, MXN, EUR, GBP, etc.
	CurrencySymbol string    `json:"currency_symbol"` // $, €, £, ¥, etc.
	Date           time.Time `json:"date"`
	Description    string    `json:"description"`
	Email          string    `json:"email"`
	Subject        string    `json:"subject"`
	Timestamp      time.Time `json:"timestamp"`
	RawAmount      string    `json:"raw_amount"` // Original text extracted
}

// ExpenseSummary represents a summary of expenses
type ExpenseSummary struct {
	TotalAmount float64
	TotalCount  int
	ByCategory  map[string]float64
	ByService   map[string]float64
	DateRange   [2]time.Time
}

// Message represents a Gmail message
type Message struct {
	ID       string
	ThreadID string
	From     string
	To       string
	Subject  string
	Body     string
	Date     time.Time
	Labels   []string
}
