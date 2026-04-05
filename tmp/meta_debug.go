package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	// Try multiple paths for .env
	err := godotenv.Load(".env")
	if err != nil {
		err = godotenv.Load("backend/.env")
	}
	if err != nil {
		log.Printf("Warning: .env file not found, using system environment")
	}

	token := os.Getenv("META_ACCESS_TOKEN")
	pageID := os.Getenv("META_FACEBOOK_PAGE_ID")

	if token == "" || pageID == "" {
		log.Fatal("META_ACCESS_TOKEN or META_FACEBOOK_PAGE_ID not found in environment")
	}

	client := &http.Client{}
	
	fmt.Println("--- PHASE 1: TESTING PAGE INSIGHTS ---")
	testURL := fmt.Sprintf("https://graph.facebook.com/v22.0/%s/insights?metric=page_impressions,page_post_engagements&period=day&access_token=%s", pageID, token)
	
	req, _ := http.NewRequest("GET", testURL, nil)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %s\n", resp.Status)
	
	var prettyJSON map[string]interface{}
	json.Unmarshal(body, &prettyJSON)
	formatted, _ := json.MarshalIndent(prettyJSON, "", "  ")
	fmt.Println(string(formatted))

	if strings.Contains(string(body), "error") {
		fmt.Println("\n[!] ERROR DETECTED: Review the 'error' block above for permission issues.")
	} else {
		fmt.Println("\n[✓] SUCCESS: Page Insights reachable. If data is empty, check date range.")
	}

	fmt.Println("\n--- PHASE 2: TESTING PAGE TOKENS ---")
	// Check if this is a Page Access Token or a User Access Token
	tokenLimitURL := fmt.Sprintf("https://graph.facebook.com/v22.0/%s?fields=access_token,name&access_token=%s", pageID, token)
	reqT, _ := http.NewRequest("GET", tokenLimitURL, nil)
	respT, _ := client.Do(reqT)
	defer respT.Body.Close()
	bodyT, _ := io.ReadAll(respT.Body)
	fmt.Printf("Token Verification Status: %s\n", respT.Status)
	fmt.Println(string(bodyT))
}
