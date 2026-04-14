package main

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	filePath := flag.String("file", "", "Path to the webhook dump file")
	webhookURL := flag.String("url", "http://localhost:8080/api/webhooks/shopify", "Target webhook URL")
	secret := flag.String("secret", "ff8a8413409945bf1ba1cf700e2f6644e1cd3944c9f5f36a45d415c5977afe92", "Shopify Webhook Secret")
	flag.Parse()

	if *filePath == "" {
		log.Fatal("Error: -file flag is required")
	}

	file, err := os.Open(*filePath)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Find the start of JSON payload
		jsonStart := strings.Index(line, "{")
		if jsonStart == -1 {
			continue
		}

		metadataStr := strings.TrimSpace(line[:jsonStart])
		payloadStr := strings.TrimSpace(line[jsonStart:])
		
		// Split metadata: [RowID] [StoreID] [Topic] [ExternalID] [DeliveryUUID]
		metadata := strings.Fields(metadataStr)
		if len(metadata) < 5 {
			log.Printf("Skip: Invalid metadata line: %s", metadataStr)
			continue
		}

		topic := metadata[2]
		deliveryID := metadata[4]

		err := sendWebhook(*webhookURL, *secret, topic, deliveryID, payloadStr)
		if err != nil {
			log.Printf("Error sending %s: %v", topic, err)
		}

		// Small delay to prevent race conditions in background processing
		// Increased to 5s for definitive showcase reliability
		time.Sleep(5 * time.Second)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}
}

func sendWebhook(url, secret, topic, deliveryID, payload string) error {
	// 1. Compute HMAC (Base64)
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write([]byte(payload))
	signature := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	// 2. Prepare Request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Topic", topic)
	req.Header.Set("X-Shopify-Hmac-Sha256", signature)
	req.Header.Set("X-Shopify-Webhook-Id", deliveryID)
	req.Header.Set("X-Shopify-Shop-Domain", "millennial-perfumer.myshopify.com")

	// 3. Send
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Replayed %-20s | Status: %d | Response: %s\n", topic, resp.StatusCode, string(body))

	return nil
}
