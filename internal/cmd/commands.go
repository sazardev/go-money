package cmd

import (
	"context"
	"fmt"
	"log"
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
		fmt.Printf("✅ Extracted %d transactions!\n", len(transactions))

		// Debug mode: show unmatched emails
		if debug && len(transactions) == 0 && len(allMessages) > 0 {
			fmt.Println("\n🔍 DEBUG: Analyzing unmatched emails...")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

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
				if debug {
					fmt.Printf("   Body (first 200 chars): %s\n", truncateString(msg.Body, 200))
				}
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

		for i, tx := range t {
			fmt.Printf("%d. %s - $%.2f %s\n", i+1, tx.ServiceName, tx.Amount, tx.Currency)
			fmt.Printf("   Category: %s | Date: %s\n", tx.Category, tx.Date.Format("2006-01-02"))
			fmt.Printf("   Subject: %s\n", tx.Subject)

			totalAmount += tx.Amount
			byCategory[tx.Category] += tx.Amount
			byService[tx.ServiceName] += tx.Amount
		}

		// Summary by category
		fmt.Println("\n📊 Summary by Category:")
		fmt.Println("─────────────────────────────────────────────────")
		for category, amount := range byCategory {
			percentage := (amount / totalAmount) * 100
			fmt.Printf("%-20s: $%8.2f (%.1f%%)\n", category, amount, percentage)
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
			fmt.Printf("%-20s: $%8.2f (%.1f%%)\n", services[i].service, services[i].amount, percentage)
		}

		// Total
		fmt.Println("\n═══════════════════════════════════════════════════")
		fmt.Printf("💰 TOTAL EXPENSES: $%.2f\n", totalAmount)
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

func init() {
	authCmd.AddCommand(loginCmd)
	calculateCmd.Flags().BoolP("debug", "d", false, "Show debug information about unmatched emails")
}
