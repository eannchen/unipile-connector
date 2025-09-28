package main

import (
	"fmt"

	"unipile-connector/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	fmt.Printf("Server Port: %s\n", cfg.Server.Port)
	fmt.Printf("Server Host: %s\n", cfg.Server.Host)
	fmt.Printf("Database Host: %s\n", cfg.Database.Host)
	fmt.Printf("Database Port: %d\n", cfg.Database.Port)
	fmt.Printf("Unipile Base URL: %s\n", cfg.Unipile.BaseURL)
}
