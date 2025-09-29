package main

import (
	"fmt"
	"log"
	"os"

	"unipile-connector/internal/domain/service"
	"unipile-connector/internal/infrastructure/client"
	"unipile-connector/internal/infrastructure/config"
)

func main() {
	fmt.Println("=== Unipile Client Test ===")

	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Check if API key is set
	if cfg.Unipile.APIKey == "" || cfg.Unipile.APIKey == "your_unipile_api_key_here" {
		fmt.Println("❌ Error: Please set UNIPILE_API_KEY in your .env file")
		fmt.Println("   Example: UNIPILE_API_KEY=your_actual_api_key_here")
		os.Exit(1)
	}

	// Initialize Unipile client
	unipileClient := client.NewUnipileClient(cfg.Unipile.BaseURL, cfg.Unipile.APIKey)

	fmt.Printf("🔗 Testing connection to: %s\n", cfg.Unipile.BaseURL)
	fmt.Printf("🔑 Using API key: %s...\n", cfg.Unipile.APIKey[:8])

	TestConnection(unipileClient)
	// TestListAccounts(unipileClient)
	// TestDeleteAccount(unipileClient)
	// TestLinkedInConnectionWithCredentials(unipileClient, "your_linkedin_email@example.com", "your_linkedin_password")
	// TestLinkedInConnectionWithCookie(unipileClient, "your_li_at_cookie_here", "your_user_agent_here")
	// TestCheckpointSolving(unipileClient, "your_account_id", "your_code")
	// TestAccountStatusCheck(unipileClient, "your_account_id")

	fmt.Println("\n🎉 Unipile client test completed!")
}

// TestConnection tests the connection to Unipile API by calling ListAccounts
func TestConnection(unipileClient service.UnipileClient) {
	fmt.Println("\nTesting connection...")
	if err := unipileClient.TestConnection(); err != nil {
		fmt.Printf("❌ Connection test failed: %v\n", err)
		fmt.Println("   This might be normal if you don't have any accounts yet")
	} else {
		fmt.Println("✅ Connection test successful!")
	}
}

// TestListAccounts tests the list accounts functionality
func TestListAccounts(unipileClient service.UnipileClient) {
	fmt.Println("\nTesting list accounts...")
	accountsResp, err := unipileClient.ListAccounts()
	if err != nil {
		fmt.Printf("❌ List accounts failed: %v\n", err)
		fmt.Println("   This might be normal if you don't have any accounts yet")
	} else {
		fmt.Printf("✅ List accounts successful!\n")
		fmt.Printf("   Object: %s\n", accountsResp.Object)
		fmt.Printf("   Number of accounts: %d\n", len(accountsResp.Items))
		if accountsResp.Cursor != nil {
			fmt.Printf("   Cursor: %s\n", *accountsResp.Cursor)
		} else {
			fmt.Println("   Cursor: null")
		}

		// Display account details
		for i, account := range accountsResp.Items {
			fmt.Printf("   Account %d:\n", i+1)
			fmt.Printf("     ID: %s\n", account.ID)
			fmt.Printf("     Name: %s\n", account.Name)
			fmt.Printf("     Type: %s\n", account.Type)
			fmt.Printf("     Created: %s\n", account.CreatedAt)
			fmt.Printf("     Sources: %d\n", len(account.Sources))
			for j, source := range account.Sources {
				fmt.Printf("       Source %d: %s (Status: %s)\n", j+1, source.ID, source.Status)
			}
		}
	}
}

// TestDeleteAccount tests the delete account functionality
func TestDeleteAccount(unipileClient service.UnipileClient) {
	fmt.Println("\nTesting delete account...")
	fmt.Println("   Note: This will test with a non-existent account ID")
	fmt.Println("   Expected result: Account not found error")

	// Test with a non-existent account ID
	testAccountID := "non_existent_account_id_12345"
	fmt.Printf("   Testing deletion of account ID: %s\n", testAccountID)

	err := unipileClient.DeleteAccount(testAccountID)
	if err != nil {
		fmt.Printf("❌ Delete account failed (expected): %v\n", err)
		fmt.Println("   This is expected behavior for a non-existent account")
	} else {
		fmt.Printf("✅ Delete account successful!\n")
		fmt.Println("   Unexpected: Account deletion succeeded with non-existent ID")
	}
}

// TestLinkedInConnectionWithCredentials tests the LinkedIn connection with credentials
func TestLinkedInConnectionWithCredentials(unipileClient service.UnipileClient, username, password string) {
	fmt.Println("\nTesting LinkedIn connection with credentials...")
	fmt.Println("   Note: This will only work with valid LinkedIn credentials")
	fmt.Println("   You can modify the credentials below for testing")

	// Example credentials - replace with real ones for testing
	testCredentials := &service.ConnectLinkedInRequest{
		Provider: "LINKEDIN",
		Username: username, // Replace with real email
		Password: password, // Replace with real password
	}

	fmt.Printf("   Testing with username: %s\n", testCredentials.Username)

	resp, err := unipileClient.ConnectLinkedIn(testCredentials)
	if err != nil {
		fmt.Printf("❌ LinkedIn connection failed: %v\n", err)
		fmt.Println("   This is expected if credentials are invalid or if there are checkpoints")
	} else {
		fmt.Printf("✅ LinkedIn connection response received!\n")
		fmt.Printf("   Object: %s\n", resp.Object)
		fmt.Printf("   Account ID: %s\n", resp.AccountID)
		fmt.Printf("   Status: %v\n", resp.Status)

		if resp.Checkpoint != nil {
			fmt.Printf("   Checkpoint required: %s\n", resp.Checkpoint.Type)
			fmt.Println("   You can use the SolveCheckpoint method to handle this")
		}

		fmt.Printf("   Row body: %s\n", resp.RowBody)
	}
}

// TestLinkedInConnectionWithCookie tests the LinkedIn connection with cookie
func TestLinkedInConnectionWithCookie(unipileClient service.UnipileClient, accessToken, userAgent string) {
	fmt.Println("\nTesting LinkedIn connection with cookie...")
	fmt.Println("   Note: This will only work with a valid li_at cookie")
	fmt.Println("   You can modify the cookie below for testing")

	// Example cookie - replace with real one for testing
	testCookie := &service.ConnectLinkedInRequest{
		Provider:    "LINKEDIN",
		AccessToken: accessToken, // Replace with real li_at cookie
		UserAgent:   userAgent,
	}

	fmt.Printf("   Testing with access token: %s...\n", testCookie.AccessToken[:8])

	resp, err := unipileClient.ConnectLinkedIn(testCookie)
	if err != nil {
		fmt.Printf("❌ LinkedIn cookie connection failed: %v\n", err)
		fmt.Println("   This is expected if the cookie is invalid or expired")
	} else {
		fmt.Printf("✅ LinkedIn cookie connection response received!\n")
		fmt.Printf("   Object: %s\n", resp.Object)
		fmt.Printf("   Account ID: %s\n", resp.AccountID)
		fmt.Printf("   Status: %v\n", resp.Status)

		if resp.Checkpoint != nil {
			fmt.Printf("   Checkpoint required: %s\n", resp.Checkpoint.Type)
		}

		fmt.Printf("   Row body: %s\n", resp.RowBody)
	}
}

// TestCheckpointSolving tests the checkpoint solving functionality
func TestCheckpointSolving(unipileClient service.UnipileClient, accountID, code string) {
	fmt.Println("\nTesting checkpoint solving...")

	checkpointReq := &service.SolveCheckpointRequest{
		Provider:  "LINKEDIN",
		AccountID: accountID,
		Code:      code, // Example 2FA code
	}

	checkpointResp, err := unipileClient.SolveCheckpoint(checkpointReq)
	if err != nil {
		fmt.Printf("❌ Checkpoint solving failed: %v\n", err)
		fmt.Println("   This is expected if the code is invalid or the checkpoint expired")
	} else {
		fmt.Printf("✅ Checkpoint solving response received!\n")
		fmt.Printf("   Object: %s\n", checkpointResp.Object)
		fmt.Printf("   Account ID: %s\n", checkpointResp.AccountID)
	}
}

// TestAccountStatusCheck tests the account status check functionality
func TestAccountStatusCheck(unipileClient service.UnipileClient, accountID string) {
	fmt.Println("\nTesting account status check...")
	account, err := unipileClient.GetAccount(accountID)
	if err != nil {
		fmt.Printf("❌ Account status check failed: %v\n", err)
	} else {
		fmt.Printf("✅ Account status response received!\n")
		fmt.Printf("   Object: %s\n", account.Object)
		fmt.Printf("   Account ID: %s\n", account.ID)
		if len(account.Sources) > 0 {
			fmt.Printf("   Status: %s\n", account.Sources[0].Status)
		} else {
			fmt.Printf("   Status: %s\n", "No sources found")
		}
	}
}
