package main

import (
	"fmt"
	"log"
	"os"

	"unipile-connector/config"
	"unipile-connector/internal/infrastructure/client"
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
	TestListAccounts(unipileClient)
	// TestLinkedInConnectionWithCredentials(unipileClient)
	// TestLinkedInConnectionWithCookie(unipileClient)
	// TestCheckpointSolving(unipileClient, nil)
	// TestAccountStatusCheck(unipileClient, nil)

	fmt.Println("\n🎉 Unipile client test completed!")
}

// TestConnection tests the connection to Unipile API by calling ListAccounts
func TestConnection(unipileClient *client.UnipileClient) {
	fmt.Println("\nTesting connection...")
	if err := unipileClient.TestConnection(); err != nil {
		fmt.Printf("❌ Connection test failed: %v\n", err)
		fmt.Println("   This might be normal if you don't have any accounts yet")
	} else {
		fmt.Println("✅ Connection test successful!")
	}
}

// TestListAccounts tests the list accounts functionality
func TestListAccounts(unipileClient *client.UnipileClient) {
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

// TestLinkedInConnectionWithCredentials tests the LinkedIn connection with credentials
func TestLinkedInConnectionWithCredentials(unipileClient *client.UnipileClient) {
	fmt.Println("\n2️⃣ Testing LinkedIn connection with credentials...")
	fmt.Println("   Note: This will only work with valid LinkedIn credentials")
	fmt.Println("   You can modify the credentials below for testing")

	// Example credentials - replace with real ones for testing
	testCredentials := &client.ConnectLinkedInRequest{
		Provider: "LINKEDIN",
		Username: "your_linkedin_email@example.com", // Replace with real email
		Password: "your_linkedin_password",          // Replace with real password
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
	}
}

// TestLinkedInConnectionWithCookie tests the LinkedIn connection with cookie
func TestLinkedInConnectionWithCookie(unipileClient *client.UnipileClient) {
	fmt.Println("\n3️⃣ Testing LinkedIn connection with cookie...")
	fmt.Println("   Note: This will only work with a valid li_at cookie")
	fmt.Println("   You can modify the cookie below for testing")

	// Example cookie - replace with real one for testing
	testCookie := &client.ConnectLinkedInRequest{
		Provider:    "LINKEDIN",
		AccessToken: "your_li_at_cookie_here", // Replace with real li_at cookie
		UserAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
	}

	fmt.Printf("   Testing with access token: %s...\n", testCookie.AccessToken[:8])

	resp2, err := unipileClient.ConnectLinkedIn(testCookie)
	if err != nil {
		fmt.Printf("❌ LinkedIn cookie connection failed: %v\n", err)
		fmt.Println("   This is expected if the cookie is invalid or expired")
	} else {
		fmt.Printf("✅ LinkedIn cookie connection response received!\n")
		fmt.Printf("   Object: %s\n", resp2.Object)
		fmt.Printf("   Account ID: %s\n", resp2.AccountID)
		fmt.Printf("   Status: %v\n", resp2.Status)

		if resp2.Checkpoint != nil {
			fmt.Printf("   Checkpoint required: %s\n", resp2.Checkpoint.Type)
		}
	}
}

// TestCheckpointSolving tests the checkpoint solving functionality
func TestCheckpointSolving(unipileClient *client.UnipileClient, resp *client.ConnectLinkedInResponse) {
	fmt.Println("\n4️⃣ Testing checkpoint solving...")
	if resp != nil && resp.AccountID != "" {
		fmt.Printf("   Using account ID from previous test: %s\n", resp.AccountID)

		checkpointReq := &client.SolveCheckpointRequest{
			Provider:  "LINKEDIN",
			AccountID: resp.AccountID,
			Code:      "123456", // Example 2FA code
		}

		checkpointResp, err := unipileClient.SolveCheckpoint(checkpointReq)
		if err != nil {
			fmt.Printf("❌ Checkpoint solving failed: %v\n", err)
			fmt.Println("   This is expected if the code is invalid or the checkpoint expired")
		} else {
			fmt.Printf("✅ Checkpoint solving response received!\n")
			fmt.Printf("   Object: %s\n", checkpointResp.Object)
			fmt.Printf("   Account ID: %s\n", checkpointResp.AccountID)

			if checkpointResp.Checkpoint != nil {
				fmt.Printf("   Another checkpoint required: %s\n", checkpointResp.Checkpoint.Type)
			}
		}
	} else {
		fmt.Println("   Skipping checkpoint test - no account ID available")
	}
}

// TestAccountStatusCheck tests the account status check functionality
func TestAccountStatusCheck(unipileClient *client.UnipileClient, resp *client.ConnectLinkedInResponse) {
	fmt.Println("\n5️⃣ Testing account status check...")
	if resp != nil && resp.AccountID != "" {
		fmt.Printf("   Checking status for account ID: %s\n", resp.AccountID)

		statusResp, err := unipileClient.GetAccountStatus(resp.AccountID)
		if err != nil {
			fmt.Printf("❌ Account status check failed: %v\n", err)
		} else {
			fmt.Printf("✅ Account status response received!\n")
			fmt.Printf("   Object: %s\n", statusResp.Object)
			fmt.Printf("   Account ID: %s\n", statusResp.AccountID)
			fmt.Printf("   Status: %s\n", statusResp.Status)
		}
	} else {
		fmt.Println("   Skipping status check - no account ID available")
	}
}
