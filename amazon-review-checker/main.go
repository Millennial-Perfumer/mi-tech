package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

const (
	AmazonBaseURL = "https://www.amazon.in"
)

func main() {
	reportPath := flag.String("report", "Category+Listings+Report_06-18-2026.xlsm", "Path to listings Excel file")
	outputPath := flag.String("output", "product_ratings.json", "Path to output JSON file")
	concurrency := flag.Int("concurrency", 5, "Number of parallel requests allowed")
	limit := flag.Int("limit", 0, "Limit the number of products to scrape")
	flag.Parse()

	// Parse Listings (defined in listings.go)
	products, err := parseListingsReport(*reportPath)
	if err != nil {
		fmt.Printf("Error reading spreadsheet: %v\n", err)
		os.Exit(1)
	}

	if *limit > 0 && *limit < len(products) {
		products = products[:*limit]
		fmt.Printf("Limiting scrape run to first %d products.\n", *limit)
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, *concurrency)
	ratingsChan := make(chan ProductRating, len(products))

	fmt.Printf("Running parallel scrape with concurrency of %d...\n", *concurrency)

	for _, p := range products {
		wg.Add(1)
		go func(prod ProductInfo) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire token
			defer func() { <-sem }() // Release token

			fmt.Printf("Fetching ASIN: %s...\n", prod.ASIN)
			avg, total, err := fetchRatingSummary(context.Background(), prod.ASIN)
			if err != nil {
				fmt.Printf("  Error fetching ASIN %s: %v\n", prod.ASIN, err)
			} else {
				fmt.Printf("  ASIN: %s | Rating: %s | Count: %s\n", prod.ASIN, avg, total)
			}

			ratingsChan <- ProductRating{
				ASIN:          prod.ASIN,
				SKUs:          prod.SKUs,
				ProductTitle:  prod.Title,
				AverageRating: avg,
				TotalRatings:  total,
			}

			// Polite random jitter delay to prevent heavy load blocks
			time.Sleep(time.Duration(500+rand.Intn(1000)) * time.Millisecond)
		}(p)
	}

	// Wait in background and close channel
	go func() {
		wg.Wait()
		close(ratingsChan)
	}()

	var results []ProductRating
	for rating := range ratingsChan {
		results = append(results, rating)
	}

	// Write to JSON
	outFile, err := os.Create(*outputPath)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	encoder := json.NewEncoder(outFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		fmt.Printf("Error serializing JSON output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nPipeline finished successfully. Output written to %s\n", *outputPath)
}
