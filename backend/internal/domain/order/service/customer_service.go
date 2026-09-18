package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"mi-tech/internal/domain/order/entity"
	"mi-tech/internal/domain/order/repository"
	"mi-tech/internal/shared/extclient/shopify"
	"mi-tech/internal/shared/util"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

var (
	customerSearchEmptyRegex = regexp.MustCompile(`(\w+)\s*=\s*['"]{2}`)
	customerSearchRangeRegex = regexp.MustCompile(`(\w+)\s*([><])\s*(\d+)`)
	customerSearchKVRegex    = regexp.MustCompile(`(\w+)[:=]\s*([^ ]+)`)
	nonDigitRegex            = regexp.MustCompile(`\D`)
)

type CustomerService struct {
	repo          *repository.CustomerRepository
	orderRepo     repository.OrderRepository
	shopifyClient *shopify.Client
}

type CustomerFilter struct {
	Search    string
	SortBy    string
	SortOrder string
	SourceID  string
	MinSpent  float64
	MaxSpent  float64
	MinOrders int
	City      string
	State     string
	// New fields for Query Style Search
	FirstName      string
	LastName       string
	Email          string
	FirstNameEmpty bool
	LastNameEmpty  bool
	EmailEmpty     bool
	Page           int
	PageSize       int
}

func NewCustomerService(repo *repository.CustomerRepository, orderRepo repository.OrderRepository, shopifyClient *shopify.Client) *CustomerService {
	return &CustomerService{
		repo:          repo,
		orderRepo:     orderRepo,
		shopifyClient: shopifyClient,
	}
}

// ImportFromCSV parses a Shopify customer export CSV and syncs it to the database.
func (s *CustomerService) ImportFromCSV(ctx context.Context, r io.Reader, sourceID string) error {
	reader := csv.NewReader(r)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return err
	}

	headerMap := make(map[string]int)
	for i, h := range header {
		headerMap[h] = i
	}

	// Validate essential headers
	required := []string{"Total Orders", "Total Spent", "First Name"}
	for _, req := range required {
		if _, ok := headerMap[req]; !ok {
			return fmt.Errorf("missing required column: %s", req)
		}
	}

	hasPhone := false
	if _, ok := headerMap["Phone"]; ok {
		hasPhone = true
	}
	if _, ok := headerMap["Default Address Phone"]; ok {
		hasPhone = true
	}
	if !hasPhone {
		return fmt.Errorf("missing phone column (Phone or Default Address Phone)")
	}

	customersByPhone := make(map[string]*entity.Customer)
	now := time.Now()

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if len(record) < len(header) {
			continue
		}

		// Try to get phone number
		phoneRaw := ""
		if idx, ok := headerMap["Default Address Phone"]; ok && idx < len(record) {
			phoneRaw = record[idx]
		}
		if phoneRaw == "" || strings.TrimSpace(phoneRaw) == "" {
			if idx, ok := headerMap["Phone"]; ok && idx < len(record) {
				phoneRaw = record[idx]
			}
		}

		phone := util.NormalizePhone(phoneRaw)
		if phone == "" {
			continue
		}

		// Parse stats
		spent, _ := strconv.ParseFloat(record[headerMap["Total Spent"]], 64)
		orders, _ := strconv.Atoi(record[headerMap["Total Orders"]])

		customer, ok := customersByPhone[phone]
		if !ok {
			customer = &entity.Customer{
				PhoneNumber: phone,
				TotalSpent:  0,
				TotalOrders: 0,
				SourceID:    sourceID,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			customersByPhone[phone] = customer
		}

		customer.SourceID = sourceID

		customer.TotalSpent += spent
		customer.TotalOrders += orders

		// Update metadata if this record has more orders or if fields are currently empty
		if !ok || orders > 0 {
			s.updateMetadata(customer, record, headerMap)
		}
	}

	if len(customersByPhone) == 0 {
		return fmt.Errorf("no valid customer records found in CSV")
	}

	var batch []entity.Customer
	for _, c := range customersByPhone {
		batch = append(batch, *c)
	}

	return s.repo.UpsertBatch(ctx, batch)
}

func (s *CustomerService) updateMetadata(c *entity.Customer, record []string, headerMap map[string]int) {
	setIfNotEmpty := func(target **string, key string) {
		if idx, ok := headerMap[key]; ok && idx < len(record) {
			val := strings.TrimSpace(record[idx])
			if val != "" {
				*target = util.StrPtr(val)
			}
		}
	}

	if idx, ok := headerMap["First Name"]; ok && idx < len(record) {
		val := s.toTitleCase(strings.TrimSpace(record[idx]))
		if val != "" {
			c.FirstName = util.StrPtr(val)
		}
	}
	if idx, ok := headerMap["Last Name"]; ok && idx < len(record) {
		val := s.toTitleCase(strings.TrimSpace(record[idx]))
		if val != "" {
			c.LastName = util.StrPtr(val)
		}
	}
	setIfNotEmpty(&c.Email, "Email")
	setIfNotEmpty(&c.Address1, "Default Address Address1")
	setIfNotEmpty(&c.Address2, "Default Address Address2")
	setIfNotEmpty(&c.City, "Default Address City")
	setIfNotEmpty(&c.ZipCode, "Default Address Zip")
	setIfNotEmpty(&c.Country, "Default Address Country Code")

	// State/Province
	if idx, ok := headerMap["Default Address Province Code"]; ok && idx < len(record) {
		val := strings.TrimSpace(record[idx])
		if val != "" {
			c.State = util.StrPtr(val)
		}
	} else if idx, ok := headerMap["Default Address State"]; ok && idx < len(record) {
		val := strings.TrimSpace(record[idx])
		if val != "" {
			c.State = util.StrPtr(val)
		}
	}
}

func (s *CustomerService) UpdateFromOrder(ctx context.Context, order *entity.Order) error {
	phone := util.NormalizePhone(util.DerefStr(order.CustomerPhone))
	if phone == "" {
		return nil
	}

	customer := &entity.Customer{
		PhoneNumber: util.NormalizePhone(phone),
		FirstName:   util.StrPtr(s.toTitleCase(util.DerefStr(order.CustomerFirstName))),
		LastName:    util.StrPtr(s.toTitleCase(util.DerefStr(order.CustomerLastName))),
		Email:       order.CustomerEmail,
		Address1:    order.CustomerAddress1,
		Address2:    order.CustomerAddress2,
		City:        order.CustomerCity,
		State:       order.CustomerState,
		Country:     order.CustomerCountry,
		ZipCode:     order.CustomerZip,
		UpdatedAt:   time.Now(),
	}

	existing, err := s.repo.GetByPhone(ctx, phone)
	if err == nil && existing != nil {
		safeMerge(customer, existing)
		// Inherit created at
		customer.CreatedAt = existing.CreatedAt
	} else {
		customer.CreatedAt = time.Now()
	}

	// Recalculate absolute totals from the orders table
	if s.orderRepo != nil {
		totalOrders, totalSpent, err := s.orderRepo.GetCustomerStats(phone)
		if err == nil {
			customer.TotalOrders = totalOrders
			customer.TotalSpent = totalSpent
		} else {
			log.Printf("Warning: Failed to recalculate stats for customer %s: %v", phone, err)
		}
	}

	if order.ID > 0 {
		return s.repo.UpsertByPhoneForOrder(ctx, customer, order.ID)
	}
	return s.repo.UpsertByPhone(ctx, customer)
}

// UpdateFromOrdersBatch aggregates customer data from a batch of orders and performs a batch upsert.
// Optimization: Reduces database roundtrips from O(N) to O(1) by fetching existing customers
// in a single query and performing a batch upsert.
// Expected Impact: For a sync of 250 orders, this reduces database calls from ~500 to 2.
func (s *CustomerService) UpdateFromOrdersBatch(ctx context.Context, orders []entity.Order) error {
	if len(orders) == 0 {
		return nil
	}

	// 1. Group orders by normalized phone number and aggregate PII
	phoneToCustomer := make(map[string]*entity.Customer)
	var phones []string
	now := time.Now()

	for i := range orders {
		phone := util.NormalizePhone(util.DerefStr(orders[i].CustomerPhone))
		if phone == "" {
			continue
		}

		incoming := &entity.Customer{
			PhoneNumber: phone,
			FirstName:   util.StrPtr(s.toTitleCase(util.DerefStr(orders[i].CustomerFirstName))),
			LastName:    util.StrPtr(s.toTitleCase(util.DerefStr(orders[i].CustomerLastName))),
			Email:       orders[i].CustomerEmail,
			Address1:    orders[i].CustomerAddress1,
			Address2:    orders[i].CustomerAddress2,
			City:        orders[i].CustomerCity,
			State:       orders[i].CustomerState,
			Country:     orders[i].CustomerCountry,
			ZipCode:     orders[i].CustomerZip,
			TotalOrders: 1,
			TotalSpent:  orders[i].TotalPrice,
			UpdatedAt:   now,
		}

		if existing, exists := phoneToCustomer[phone]; exists {
			incoming.TotalOrders += existing.TotalOrders
			incoming.TotalSpent += existing.TotalSpent
			safeMerge(incoming, existing)
			phoneToCustomer[phone] = incoming
		} else {
			phones = append(phones, phone)
			phoneToCustomer[phone] = incoming
		}
	}

	if len(phones) == 0 {
		return nil
	}

	// 2. Fetch existing customers in batch to preserve totals and created_at
	existingCustomers, err := s.repo.GetByPhones(ctx, phones)
	if err != nil {
		return fmt.Errorf("UpdateFromOrdersBatch: failed to fetch existing customers: %w", err)
	}

	existingMap := make(map[string]entity.Customer)
	for _, c := range existingCustomers {
		existingMap[c.PhoneNumber] = c
	}

	// 3. Fetch absolute totals from orders table for all phones in batch
	var statsMap map[string]struct {
		Count int
		Sum   float64
	}
	if s.orderRepo != nil {
		statsMap, err = s.orderRepo.GetCustomersStats(phones)
		if err != nil {
			log.Printf("Warning: Failed to fetch bulk customer stats: %v", err)
		}
	}

	var customersToUpsert []entity.Customer
	for _, phone := range phones {
		customer := phoneToCustomer[phone]
		if existing, found := existingMap[phone]; found {
			customer.CreatedAt = existing.CreatedAt
			safeMerge(customer, &existing)
		} else {
			customer.CreatedAt = now
		}

		// Apply absolute totals if available
		if statsMap != nil {
			if stats, ok := statsMap[phone]; ok {
				customer.TotalOrders = stats.Count
				customer.TotalSpent = stats.Sum
			}
		}

		customersToUpsert = append(customersToUpsert, *customer)
	}

	// 4. Batch Upsert
	return s.repo.UpsertBatch(ctx, customersToUpsert)
}

// UpsertFromWebhook securely updates or creates a customer entirely from a customer webhook payload.
func (s *CustomerService) UpsertFromWebhook(ctx context.Context, cust *entity.Customer) error {
	if cust.PhoneNumber == "" {
		return fmt.Errorf("phone number is required for customer webhook upsert")
	}

	cust.PhoneNumber = util.NormalizePhone(cust.PhoneNumber)
	cust.FirstName = util.StrPtr(s.toTitleCase(util.DerefStr(cust.FirstName)))
	cust.LastName = util.StrPtr(s.toTitleCase(util.DerefStr(cust.LastName)))

	existing, err := s.repo.GetByPhone(ctx, cust.PhoneNumber)
	if err == nil && existing != nil {
		safeMerge(cust, existing)
		cust.CreatedAt = existing.CreatedAt
		if cust.TotalOrders == 0 && existing.TotalOrders > 0 {
			cust.TotalOrders = existing.TotalOrders
		}
		if cust.TotalSpent == 0 && existing.TotalSpent > 0 {
			cust.TotalSpent = existing.TotalSpent
		}
	} else {
		if cust.CreatedAt.IsZero() {
			cust.CreatedAt = time.Now()
		}
	}

	if cust.UpdatedAt.IsZero() {
		cust.UpdatedAt = time.Now()
	}

	return s.repo.UpsertByPhone(ctx, cust)
}

func safeMerge(newCust, oldCust *entity.Customer) {
	if newCust.FirstName == nil {
		newCust.FirstName = oldCust.FirstName
	}
	if newCust.LastName == nil {
		newCust.LastName = oldCust.LastName
	}
	if newCust.Email == nil {
		newCust.Email = oldCust.Email
	}
	if newCust.Address1 == nil {
		newCust.Address1 = oldCust.Address1
	}
	if newCust.Address2 == nil {
		newCust.Address2 = oldCust.Address2
	}
	if newCust.City == nil {
		newCust.City = oldCust.City
	}
	if newCust.State == nil {
		newCust.State = oldCust.State
	}
	if newCust.Country == nil {
		newCust.Country = oldCust.Country
	}
	if newCust.ZipCode == nil {
		newCust.ZipCode = oldCust.ZipCode
	}
}

func (s *CustomerService) ListCustomers(ctx context.Context, f CustomerFilter) ([]entity.Customer, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 20
	}
	offset := (f.Page - 1) * f.PageSize

	if f.Search != "" {
		parsed := s.parseSearchQuery(f.Search)
		if parsed.MinSpent > 0 {
			f.MinSpent = parsed.MinSpent
		}
		if parsed.MaxSpent > 0 {
			f.MaxSpent = parsed.MaxSpent
		}
		if parsed.MinOrders > 0 {
			f.MinOrders = parsed.MinOrders
		}
		if parsed.City != "" {
			f.City = parsed.City
		}
		if parsed.State != "" {
			f.State = parsed.State
		}
		if parsed.SourceID != "" {
			f.SourceID = parsed.SourceID
		}
		if parsed.FirstName != "" {
			f.FirstName = parsed.FirstName
		}
		if parsed.LastName != "" {
			f.LastName = parsed.LastName
		}
		if parsed.Email != "" {
			f.Email = parsed.Email
		}
		if parsed.FirstNameEmpty {
			f.FirstNameEmpty = true
		}
		if parsed.Las