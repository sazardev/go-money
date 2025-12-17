package cmd

import (
	"fmt"

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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: Implement login")
	},
}

var calculateCmd = &cobra.Command{
	Use:   "calculate",
	Short: "Calculate expenses",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: Implement calculate")
	},
}

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Generate graph",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: Implement graph")
	},
}

func init() {
	authCmd.AddCommand(loginCmd)
}
