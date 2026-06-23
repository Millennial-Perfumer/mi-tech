package main

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os/exec"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var userAgents = []string{
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
}

type ProductRating struct {
	ASIN          string   `json:"asin"`
	SKUs          []string `json:"skus"`
	ProductTitle  string   `json:"product_title"`
	AverageRating string   `json:"average_rating"`
	TotalRatings  string   `json:"total_ratings"`
}

func cleanCount(s string) string {
	s = strings.TrimSpace(s)
	var res strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			res.WriteRune(r)
		}
	}
	if res.Len() == 0 {
		return "0"
	}
	return res.String()
}

func cleanRating(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.Index(s, "out of"); idx != -1 {
		s = strings.TrimSpace(s[:idx])
	} else if idx := strings.Index(s, "of 5"); idx != -1 {
		s = strings.TrimSpace(s[:idx])
	}
	return s
}

func fetchRatingSummary(ctx context.Context, asin string) (string, string, error) {
	url := fmt.Sprintf("%s/dp/%s", AmazonBaseURL, asin)
	userAgent := userAgents[rand.Intn(len(userAgents))]

	cmd := exec.CommandContext(ctx, "curl", "-s", "--compressed",
		"-H", "User-Agent: "+userAgent,
		"-H", "Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		"-H", "Accept-Language: en-US,en;q=0.9",
		url,
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		return "0", "0", err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(stdout.Bytes()))
	if err != nil {
		return "0", "0", err
	}

	avgRating := "0"
	totalRatings := "0"

	// 1. Try average rating inside primary product review elements only
	doc.Find("#averageCustomerReviews .a-icon-alt").First().Each(func(i int, s *goquery.Selection) {
		avgRating = strings.TrimSpace(s.Text())
	})
	if avgRating == "0" || avgRating == "" {
		doc.Find("#acrPopover .a-icon-alt").First().Each(func(i int, s *goquery.Selection) {
			avgRating = strings.TrimSpace(s.Text())
		})
	}
	if avgRating == "0" || avgRating == "" {
		doc.Find("[data-hook='average-star-rating']").First().Each(func(i int, s *goquery.Selection) {
			avgRating = strings.TrimSpace(s.Text())
		})
	}

	// 2. Try total ratings count inside primary product review elements only
	doc.Find("#acrCustomerReviewText").First().Each(func(i int, s *goquery.Selection) {
		totalRatings = strings.TrimSpace(s.Text())
	})
	if totalRatings == "0" || totalRatings == "" {
		doc.Find("[data-hook='total-review-count']").First().Each(func(i int, s *goquery.Selection) {
			totalRatings = strings.TrimSpace(s.Text())
		})
	}

	return cleanRating(avgRating), cleanCount(totalRatings), nil
}
