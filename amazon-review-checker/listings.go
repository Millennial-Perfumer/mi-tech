package main

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

type ProductInfo struct {
	ASIN  string   `json:"asin"`
	Title string   `json:"product_title"`
	SKUs  []string `json:"skus"`
}

func parseListingsReport(filePath string) ([]ProductInfo, error) {
	fmt.Printf("Loading listings report from: %s\n", filePath)
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := f.GetRows("Template")
	if err != nil {
		return nil, err
	}

	if len(rows) < 6 {
		return nil, fmt.Errorf("sheet has insufficient rows (less than 6)")
	}

	// Row 4 has headers (0-indexed: row idx 3)
	headers := rows[3]
	colMap := make(map[string]int)
	for idx, h := range headers {
		h = strings.TrimSpace(h)
		if h != "" {
			colMap[h] = idx
		}
	}

	skuIdx, hasSku := colMap["SKU"]
	asinIdx, hasAsin := colMap["Product Id"]
	idTypeIdx, hasIdType := colMap["Product Id Type"]
	titleIdx, hasTitle := colMap["Title"]
	parentageIdx, hasParentage := colMap["Parentage Level"]

	if !hasSku || !hasAsin {
		return nil, fmt.Errorf("could not find SKU or Product Id columns in row headers")
	}

	uniqueProducts := make(map[string]*ProductInfo)

	// Data rows start at 6 (idx 5)
	for i := 5; i < len(rows); i++ {
		row := rows[i]
		if len(row) <= skuIdx || len(row) <= asinIdx {
			continue
		}

		sku := strings.TrimSpace(row[skuIdx])
		if sku == "" {
			continue
		}

		asinVal := strings.TrimSpace(row[asinIdx])
		idType := ""
		if hasIdType && len(row) > idTypeIdx {
			idType = strings.TrimSpace(row[idTypeIdx])
		}

		title := ""
		if hasTitle && len(row) > titleIdx {
			title = strings.TrimSpace(row[titleIdx])
		}

		parentage := ""
		if hasParentage && len(row) > parentageIdx {
			parentage = strings.TrimSpace(row[parentageIdx])
		}

		// Filter for child listings
		if strings.ToLower(parentage) == "parent" {
			continue
		}

		var asin string
		if idType == "ASIN" && asinVal != "" {
			asin = asinVal
		} else if strings.HasPrefix(asinVal, "B0") && len(asinVal) == 10 {
			asin = asinVal
		}

		if asin == "" {
			continue
		}

		if prod, exists := uniqueProducts[asin]; exists {
			prod.SKUs = append(prod.SKUs, sku)
			if len(title) > len(prod.Title) {
				prod.Title = title
			}
		} else {
			uniqueProducts[asin] = &ProductInfo{
				ASIN:  asin,
				Title: title,
				SKUs:  []string{sku},
			}
		}
	}

	result := make([]ProductInfo, 0, len(uniqueProducts))
	for _, p := range uniqueProducts {
		result = append(result, *p)
	}

	fmt.Printf("Parsed %d unique child ASINs from report.\n", len(result))
	return result, nil
}
