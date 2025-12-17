package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sazardev/go-money/internal/auth"
	"github.com/sazardev/go-money/internal/extractor"
	"github.com/sazardev/go-money/internal/gmail"
	"github.com/sazardev/go-money/internal/models"
	"github.com/spf13/cobra"
)

var Version = "1.0.0"

var rootCmd = &cobra.Command{
	Use:   "gm",
	Short: "GO Money - CLI for managing expenses from Gmail",
	Long: `GO Money helps you manage your finances by extracting 
transaction data from your Gmail account.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(calculateCmd)
	rootCmd.AddCommand(graphCmd)

	// Add subcommands
	authCmd.AddCommand(loginCmd)

	// Add flags to calculateCmd
	calculateCmd.Flags().BoolP("debug", "d", false, "Enable debug mode")
	calculateCmd.Flags().StringP("from", "f", "", "Start date (YYYY-MM-DD format)")
	calculateCmd.Flags().StringP("to", "t", "", "End date (YYYY-MM-DD format)")
	calculateCmd.Flags().StringP("month", "m", "", "Specific month (YYYY-MM format)")
	calculateCmd.Flags().StringP("currency", "c", "", "Filter by currency (USD, MXN, EUR, GBP, JPY, CAD)")
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("GO Money v%s\n", Version)
	},
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Google",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Create authenticator
		authenticator := auth.NewAuthenticator()

		// Get token (this will open browser or request manual auth)
		token, err := authenticator.GetToken(ctx)
		if err != nil {
			log.Printf("❌ Authentication failed: %v\n", err)
			return err
		}

		// Success
		fmt.Println("✅ Successfully authenticated with Google!")
		fmt.Printf("📧 Access token obtained. Token expires at: %v\n", token.Expiry)
		fmt.Println("🎉 You can now use 'gm calculate' to extract your expenses!")

		return nil
	},
}

var calculateCmd = &cobra.Command{
	Use:   "calculate",
	Short: "Calculate and summarize expenses",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		debug, _ := cmd.Flags().GetBool("debug")
		fromStr, _ := cmd.Flags().GetString("from")
		toStr, _ := cmd.Flags().GetString("to")
		month, _ := cmd.Flags().GetString("month")
		currency, _ := cmd.Flags().GetString("currency")

		// Parse date filters
		var fromDate, toDate time.Time
		var err error

		if fromStr != "" {
			fromDate, err = parseDate(fromStr)
			if err != nil {
				fmt.Printf("❌ Invalid --from date: %v (use YYYY-MM-DD)\n", err)
				return nil
			}
		}

		if toStr != "" {
			toDate, err = parseDate(toStr)
			if err != nil {
				fmt.Printf("❌ Invalid --to date: %v (use YYYY-MM-DD)\n", err)
				return nil
			}
		}

		// Handle month filter (YYYY-MM format)
		if month != "" {
			parts := strings.Split(month, "-")
			if len(parts) == 2 {
				year := parts[0]
				monthNum := parts[1]
				dateStr := year + "-" + monthNum + "-01"
				if monthDate, err := parseDate(dateStr); err == nil {
					fromDate = monthDate
					// Set toDate to last day of month
					toDate = monthDate.AddDate(0, 1, -1).Add(24*time.Hour - time.Nanosecond)
				}
			}
		}

		// Step 1: Load existing token
		fmt.Println("📊 Loading your authentication token...")
		authenticator := auth.NewAuthenticator()
		token, err := authenticator.GetToken(ctx)
		if err != nil {
			fmt.Printf("❌ Failed to load authentication: %v\n", err)
			fmt.Println("💡 Tip: Run 'gm auth login' first to authenticate")
			return err
		}
		fmt.Println("✅ Token loaded successfully!")

		// Step 2: Connect to Gmail
		fmt.Println("\n📧 Connecting to Gmail...")
		gmailService, err := gmail.NewGmailService(ctx, token)
		if err != nil {
			fmt.Printf("❌ Failed to connect to Gmail: %v\n", err)
			return err
		}
		fmt.Println("✅ Connected to Gmail!")

		// Step 3: Get messages with transaction queries
		fmt.Println("\n🔍 Searching for transaction emails...")

		// Search queries for common transaction keywords
		queries := []string{
			"receipt",
			"payment",
			"transaction",
			"order confirmation",
			"booking confirmation",
		}

		var allMessages []*models.Message
		for _, query := range queries {
			messages, err := gmailService.GetMessages(ctx, query)
			if err != nil {
				log.Printf("⚠️  Warning: Could not search for '%s': %v\n", query, err)
				continue
			}
			allMessages = append(allMessages, messages...)
		}

		fmt.Printf("✅ Found %d transaction emails!\n", len(allMessages))

		if len(allMessages) == 0 {
			fmt.Println("\n⚠️  No transaction emails found.")
			fmt.Println("💡 Tip: Make sure you have emails from services like Uber, Amazon, Netflix, etc.")
			return nil
		}

		// Step 4: Extract transactions
		fmt.Println("\n💰 Extracting transactions...")
		txExtractor, err := extractor.NewTransactionExtractor()
		if err != nil {
			fmt.Printf("❌ Failed to initialize transaction extractor: %v\n", err)
			return err
		}

		transactions := txExtractor.ExtractTransactions(allMessages)

		// Filter by date range if provided
		if !fromDate.IsZero() || !toDate.IsZero() {
			var filtered []*models.Transaction
			for _, tx := range transactions {
				txDate := tx.Date
				if !fromDate.IsZero() && txDate.Before(fromDate) {
					continue
				}
				if !toDate.IsZero() && txDate.After(toDate) {
					continue
				}
				filtered = append(filtered, tx)
			}
			transactions = filtered
			if len(transactions) == 0 {
				fmt.Println("⚠️  No transactions found in the specified date range")
				return nil
			}
		}

		// Filter by currency if provided
		if currency != "" {
			var filtered []*models.Transaction
			for _, tx := range transactions {
				if strings.EqualFold(tx.Currency, currency) {
					filtered = append(filtered, tx)
				}
			}
			transactions = filtered
			if len(transactions) == 0 {
				fmt.Printf("⚠️  No transactions found in %s currency\n", currency)
				return nil
			}
		}

		// Show debug information if requested
		if debug {
			// Show first 10 emails for debugging
			limit := 10
			if len(allMessages) < limit {
				limit = len(allMessages)
			}

			for i := 0; i < limit; i++ {
				msg := allMessages[i]
				fmt.Printf("\n📧 Email %d:\n", i+1)
				fmt.Printf("   From: %s\n", msg.From)
				fmt.Printf("   Subject: %s\n", msg.Subject)
				fmt.Printf("   Date: %s\n", msg.Date)
				fmt.Printf("   Body (first 200 chars): %s\n", truncateString(msg.Body, 200))
			}

			fmt.Println("\n💡 Tip: Check the email domains and keywords. You may need to update tracker-mails.json")
		}

		// Step 5: Display results
		if len(transactions) == 0 {
			fmt.Println("\n⚠️  No transactions could be extracted from the emails.")
			fmt.Println("💡 Tip: Some emails might not match the configured services.")
			if !debug {
				fmt.Println("💡 Try: gm calculate --debug  (to see unmatched emails)")
			}
			return nil
		}

		displayExpenseSummary(transactions)

		return nil
	},
}

// displayExpenseSummary displays a formatted expense summary
func displayExpenseSummary(transactions interface{}) {
	// For now, show basic info
	fmt.Println("\n" + "═══════════════════════════════════════════════════")
	fmt.Println("           💸 EXPENSE SUMMARY 💸")
	fmt.Println("═══════════════════════════════════════════════════")

	// Display basic info
	switch t := transactions.(type) {
	case []*models.Transaction:
		if len(t) == 0 {
			fmt.Println("No transactions found")
			return
		}

		// Show individual transactions
		fmt.Println("\n📝 Transactions:")
		fmt.Println("─────────────────────────────────────────────────")

		totalAmount := 0.0
		byCategory := make(map[string]float64)
		byService := make(map[string]float64)
		currenciesSeen := make(map[string]string)

		for i, tx := range t {
			fmt.Printf("%d. %s - %s%.2f %s\n", i+1, tx.ServiceName, tx.CurrencySymbol, tx.Amount, tx.Currency)
			fmt.Printf("   Category: %s | Date: %s\n", tx.Category, tx.Date.Format("2006-01-02"))
			fmt.Printf("   Subject: %s\n", tx.Subject)

			totalAmount += tx.Amount
			byCategory[tx.Category] += tx.Amount
			byService[tx.ServiceName] += tx.Amount
			currenciesSeen[tx.Currency] = tx.CurrencySymbol
		}

		// Get symbol for summary (use first found)
		summarySymbol := "$"
		for _, sym := range currenciesSeen {
			summarySymbol = sym
			break
		}

		// Summary by category
		fmt.Println("\n📊 Summary by Category:")
		fmt.Println("─────────────────────────────────────────────────")
		for category, amount := range byCategory {
			percentage := (amount / totalAmount) * 100
			fmt.Printf("%-20s: %s%8.2f (%.1f%%)\n", category, summarySymbol, amount, percentage)
		}

		// Summary by service
		fmt.Println("\n🏪 Summary by Service (Top 5):")
		fmt.Println("─────────────────────────────────────────────────")

		// Sort services by amount (simple bubble sort for demo)
		type kv struct {
			service string
			amount  float64
		}
		var services []kv
		for k, v := range byService {
			services = append(services, kv{k, v})
		}

		// Sort descending
		for i := 0; i < len(services); i++ {
			for j := i + 1; j < len(services); j++ {
				if services[j].amount > services[i].amount {
					services[i], services[j] = services[j], services[i]
				}
			}
		}

		// Show top 5
		limit := 5
		if len(services) < limit {
			limit = len(services)
		}

		for i := 0; i < limit; i++ {
			percentage := (services[i].amount / totalAmount) * 100
			fmt.Printf("%-20s: %s%8.2f (%.1f%%)\n", services[i].service, summarySymbol, services[i].amount, percentage)
		}

		// Total
		fmt.Println("\n═══════════════════════════════════════════════════")
		fmt.Printf("💰 TOTAL EXPENSES: %s%.2f\n", summarySymbol, totalAmount)
		fmt.Printf("📈 Number of Transactions: %d\n", len(t))
		if len(t) > 0 {
			fmt.Printf("📅 Date Range: %s to %s\n",
				getEarliestDate(t).Format("2006-01-02"),
				getLatestDate(t).Format("2006-01-02"))
		}
		fmt.Println("═══════════════════════════════════════════════════\n")

	default:
		fmt.Println("Unknown transaction type")
	}
}

// Helper functions
func getEarliestDate(transactions []*models.Transaction) time.Time {
	if len(transactions) == 0 {
		return time.Now()
	}
	earliest := transactions[0].Date
	for _, tx := range transactions {
		if tx.Date.Before(earliest) {
			earliest = tx.Date
		}
	}
	return earliest
}

func getLatestDate(transactions []*models.Transaction) time.Time {
	if len(transactions) == 0 {
		return time.Now()
	}
	latest := transactions[0].Date
	for _, tx := range transactions {
		if tx.Date.After(latest) {
			latest = tx.Date
		}
	}
	return latest
}

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Generate graph",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: Implement graph")
	},
}

// Helper function to truncate strings
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Helper function to parse date strings (YYYY-MM-DD format)
func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}
