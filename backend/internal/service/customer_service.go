package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"mi-tech/internal/client/shopify"
	"mi-tech/internal/dto"
	"mi-tech/internal/entity"
	"mi-tech/internal/repository"
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

		phone := entity.NormalizePhone(phoneRaw)
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
				*target = entity.StrPtr(val)
			}
		}
	}

	if idx, ok := headerMap["First Name"]; ok && idx < len(record) {
		val := s.toTitleCase(strings.TrimSpace(record[idx]))
		if val != "" {
			c.FirstName = entity.StrPtr(val)
		}
	}
	if idx, ok := headerMap["Last Name"]; ok && idx < len(record) {
		val := s.toTitleCase(strings.TrimSpace(record[idx]))
		if val != "" {
			c.LastName = entity.StrPtr(val)
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
			c.State = entity.StrPtr(val)
		}
	} else if idx, ok := headerMap["Default Address State"]; ok && idx < len(record) {
		val := strings.TrimSpace(record[idx])
		if val != "" {
			c.State = entity.StrPtr(val)
		}
	}
}

func (s *CustomerService) UpdateFromOrder(ctx context.Context, order *entity.Order) error {
	phone := entity.NormalizePhone(entity.DerefStr(order.CustomerPhone))
	if phone == "" {
		return nil
	}

	customer := &entity.Customer{
		PhoneNumber: entity.NormalizePhone(phone),
		FirstName:   entity.StrPtr(s.toTitleCase(entity.DerefStr(order.CustomerFirstName))),
		LastName:    entity.StrPtr(s.toTitleCase(entity.DerefStr(order.CustomerLastName))),
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
		phone := entity.NormalizePhone(entity.DerefStr(orders[i].CustomerPhone))
		if phone == "" {
			continue
		}

		incoming := &entity.Customer{
			PhoneNumber: phone,
			FirstName:   entity.StrPtr(s.toTitleCase(entity.DerefStr(orders[i].CustomerFirstName))),
			LastName:    entity.StrPtr(s.toTitleCase(entity.DerefStr(orders[i].CustomerLastName))),
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

	cust.PhoneNumber = entity.NormalizePhone(cust.PhoneNumber)
	cust.FirstName = entity.StrPtr(s.toTitleCase(entity.DerefStr(cust.FirstName)))
	cust.LastName = entity.StrPtr(s.toTitleCase(entity.DerefStr(cust.LastName)))

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
		if parsed.LastNameEmpty {
			f.LastNameEmpty = true
		}
		if parsed.EmailEmpty {
			f.EmailEmpty = true
		}
		f.Search = parsed.Search
	}

	dbSortBy := "updated_at"
	switch f.SortBy {
	case "name":
		dbSortBy = "first_name"
	case "phone":
		dbSortBy = "phone_number"
	case "email":
		dbSortBy = "email"
	case "orders":
		dbSortBy = "total_orders"
	case "spent":
		dbSortBy = "total_spent"
	case "activity":
		dbSortBy = "updated_at"
	}

	return s.repo.List(ctx, f.Search, dbSortBy, f.SortOrder, f.SourceID, f.MinSpent, f.MaxSpent, f.MinOrders, f.City, f.State, f.FirstName, f.LastName, f.Email, f.FirstNameEmpty, f.LastNameEmpty, f.EmailEmpty, offset, f.PageSize)
}

func (s *CustomerService) parseSearchQuery(search string) CustomerFilter {
	f := CustomerFilter{}

	// Support "field = ''" or "field = \"\""
	// Performance: Use pre-compiled regex to avoid redundant allocations per request
	matches := customerSearchEmptyRegex.FindAllStringSubmatch(search, -1)
	for _, m := range matches {
		field := strings.ToLower(m[1])
		switch field {
		case "first_name":
			f.FirstNameEmpty = true
		case "last_name":
			f.LastNameEmpty = true
		case "email":
			f.EmailEmpty = true
		}
		search = strings.Replace(search, m[0], "", 1)
	}

	// Support "field > 1000" or "field < 5000"
	// Performance: Use pre-compiled regex to avoid redundant allocations per request
	matches = customerSearchRangeRegex.FindAllStringSubmatch(search, -1)
	for _, m := range matches {
		field := strings.ToLower(m[1])
		op := m[2]
		val, _ := strconv.ParseFloat(m[3], 64)
		switch field {
		case "spent":
			if op == ">" {
				f.MinSpent = val
			} else {
				f.MaxSpent = val
			}
		case "orders":
			if op == ">" {
				f.MinOrders = int(val)
			}
		}
		search = strings.Replace(search, m[0], "", 1)
	}

	// Support "field:value" or "field=value"
	// Performance: Use pre-compiled regex to avoid redundant allocations per request
	matches = customerSearchKVRegex.FindAllStringSubmatch(search, -1)
	for _, m := range matches {
		field := strings.ToLower(m[1])
		val := strings.Trim(m[2], `"'`)
		switch field {
		case "city":
			f.City = val
		case "state":
			f.State = val
		case "first_name":
			f.FirstName = val
		case "last_name":
			f.LastName = val
		case "email":
			f.Email = val
		case "source":
			f.SourceID = val
		}
		search = strings.Replace(search, m[0], "", 1)
	}

	f.Search = strings.TrimSpace(search)
	return f
}

func (s *CustomerService) DeleteAllCustomers(ctx context.Context) error {
	return s.repo.DeleteAll(ctx)
}

func (s *CustomerService) toTitleCase(str string) string {
	if str == "" {
		return ""
	}
	words := strings.Fields(strings.ToLower(str))
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

func (s *CustomerService) GetCustomersByIDs(ctx context.Context, ids []uint) ([]entity.Customer, error) {
	return s.repo.GetByIDs(ctx, ids)
}
func (s *CustomerService) CreateCustomer(ctx context.Context, cust *entity.Customer, syncToShopify bool) error {
	cust.PhoneNumber = entity.NormalizePhone(cust.PhoneNumber)
	cust.FirstName = entity.StrPtr(s.toTitleCase(entity.DerefStr(cust.FirstName)))
	cust.LastName = entity.StrPtr(s.toTitleCase(entity.DerefStr(cust.LastName)))
	// Ensure DeletedAt is reset to reactive the customer if it was previously soft-deleted
	cust.DeletedAt = gorm.DeletedAt{}

	// 1. Use UpsertByPhone to handle reactive and conflict
	err := s.repo.UpsertByPhone(ctx, cust)
	if err != nil {
		return err
	}

	// 2. Sync to Shopify if requested
	if syncToShopify && s.shopifyClient != nil {
		sc := dto.ShopifyRestCustomer{
			FirstName: entity.DerefStr(cust.FirstName),
			LastName:  entity.DerefStr(cust.LastName),
			Email:     entity.DerefStr(cust.Email),
			Phone:     cust.PhoneNumber,
			Addresses: []dto.ShopifyRestAddress{
				{
					Address1: entity.DerefStr(cust.Address1),
					Address2: entity.DerefStr(cust.Address2),
					City:     entity.DerefStr(cust.City),
					Province: entity.DerefStr(cust.State),
					Country: func() string {
						c := entity.DerefStr(cust.Country)
						if c == "" {
							return "India"
						}
						return c
					}(),
					Zip: entity.DerefStr(cust.ZipCode),
				},
			},
		}

		resp, err := s.shopifyClient.CreateCustomer(sc)
		if err == nil && resp != nil {
			// Update local customer with shopify ID
			extID := strconv.FormatInt(resp.ID, 10)
			cust.ExternalID = &extID
			cust.SourceID = "shopify"
			s.repo.Update(ctx, cust)
		} else if err != nil {
			log.Printf("Failed to sync new customer to Shopify: %v", err)
			// We continue because local creation succeeded
		}
	}

	return nil
}

func (s *CustomerService) UpdateCustomer(ctx context.Context, cust *entity.Customer, syncToShopify bool) error {
	// 1. Fetch existing customer to preserve statistics and metadata
	existing, err := s.repo.GetByID(ctx, cust.ID)
	if err != nil {
		return fmt.Errorf("customer not found: %w", err)
	}

	// 2. Patch only provided fields (preventing data loss of stats/metadata)
	if cust.PhoneNumber != "" {
		existing.PhoneNumber = entity.NormalizePhone(cust.PhoneNumber)
	}
	if cust.FirstName != nil {
		existing.FirstName = entity.StrPtr(s.toTitleCase(entity.DerefStr(cust.FirstName)))
	}
	if cust.LastName != nil {
		existing.LastName = entity.StrPtr(s.toTitleCase(entity.DerefStr(cust.LastName)))
	}
	if cust.Email != nil {
		existing.Email = cust.Email
	}
	if cust.Address1 != nil {
		existing.Address1 = cust.Address1
	}
	if cust.Address2 != nil {
		existing.Address2 = cust.Address2
	}
	if cust.City != nil {
		existing.City = cust.City
	}
	if cust.State != nil {
		existing.State = cust.State
	}
	if cust.Country != nil {
		existing.Country = cust.Country
	}
	if cust.ZipCode != nil {
		existing.ZipCode = cust.ZipCode
	}

	existing.UpdatedAt = time.Now()

	// 3. Update locally first
	err = s.repo.Update(ctx, existing)
	if err != nil {
		return err
	}

	// 4. Sync to Shopify if requested
	if syncToShopify && s.shopifyClient != nil {
		sc := dto.ShopifyRestCustomer{
			FirstName: entity.DerefStr(existing.FirstName),
			LastName:  entity.DerefStr(existing.LastName),
			Email:     entity.DerefStr(existing.Email),
			Phone:     existing.PhoneNumber,
		}

		if existing.ExternalID != nil && *existing.ExternalID != "" {
			// SYNC: Update existing Shopify customer
			extID, _ := strconv.ParseInt(*existing.ExternalID, 10, 64)
			if extID > 0 {
				// Try to find address ID to update existing address instead of creating new one
				var addressID int64
				remoteCust, err := s.shopifyClient.GetCustomer(extID)
				if err == nil && remoteCust != nil {
					for _, addr := range remoteCust.Addresses {
						if addr.Default {
							addressID = addr.ID
							break
						}
					}
					if addressID == 0 && len(remoteCust.Addresses) > 0 {
						addressID = remoteCust.Addresses[0].ID
					}
				}

				if existing.Address1 != nil {
					sc.Addresses = []dto.ShopifyRestAddress{
						{
							ID:       addressID,
							Address1: entity.DerefStr(existing.Address1),
							Address2: entity.DerefStr(existing.Address2),
							City:     entity.DerefStr(existing.City),
							Province: entity.DerefStr(existing.State),
							Country: func() string {
								c := entity.DerefStr(existing.Country)
								if c == "" {
									return "India"
								}
								return c
							}(),
							Zip:     entity.DerefStr(existing.ZipCode),
							Default: true,
						},
					}
				}
				_, err = s.shopifyClient.UpdateCustomer(extID, sc)
				if err != nil {
					log.Printf("Failed to sync customer update to Shopify: %v", err)
				}
			}
		} else {
			// LINK: Create customer on Shopify if it doesn't exist yet
			if existing.Address1 != nil {
				sc.Addresses = []dto.ShopifyRestAddress{
					{
						Address1: entity.DerefStr(existing.Address1),
						Address2: entity.DerefStr(existing.Address2),
						City:     entity.DerefStr(existing.City),
						Province: entity.DerefStr(existing.State),
						Country: func() string {
							c := entity.DerefStr(existing.Country)
							if c == "" {
								return "India"
							}
							return c
						}(),
						Zip:     entity.DerefStr(existing.ZipCode),
						Default: true,
					},
				}
			}
			resp, err := s.shopifyClient.CreateCustomer(sc)
			if err == nil && resp != nil {
				extID := strconv.FormatInt(resp.ID, 10)
				existing.ExternalID = &extID
				existing.SourceID = "shopify"
				s.repo.Update(ctx, existing)
			} else if err != nil {
				log.Printf("Failed to create new customer on Shopify during update: %v", err)
			}
		}
	}

	return nil
}

func (s *CustomerService) GetCustomerByID(ctx context.Context, id int64) (*entity.Customer, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CustomerService) DeleteCustomer(ctx context.Context, id int64) error {
	// 1. Get customer to check for Shopify ID
	cust, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 2. Sync deletion to Shopify if linked
	if s.shopifyClient != nil && cust.ExternalID != nil && *cust.ExternalID != "" {
		extID, _ := strconv.ParseInt(*cust.ExternalID, 10, 64)
		if extID > 0 {
			err := s.shopifyClient.DeleteCustomer(extID)
			if err != nil {
				log.Printf("Failed to sync customer deletion to Shopify: %v", err)
				// We still delete locally
			}
		}
	}

	// 3. Delete locally (Soft delete)
	return s.repo.Delete(ctx, id)
}

func (s *CustomerService) DeleteByExternalID(ctx context.Context, externalID string) error {
	cust, err := s.repo.GetByExternalID(ctx, externalID)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, cust.ID)
}

// BulkDeleteCustomers optimizes bulk customer deletion by fetching entities in one batch,
// parallelizing Shopify deletions, and performing a single batch database delete.
// Optimization: Reduces DB roundtrips from O(2N) to O(2) and handles Shopify deletions concurrently.
// Expected Impact: Significantly slashes latency for large batch deletions.
func (s *CustomerService) BulkDeleteCustomers(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	// 1. Fetch all customers in one batch to get ExternalIDs
	uids := make([]uint, len(ids))
	for i, id := range ids {
		uids[i] = uint(id)
	}

	customers, err := s.repo.GetByIDs(ctx, uids)
	if err != nil {
		return fmt.Errorf("BulkDelete: failed to fetch customers: %w", err)
	}

	// 2. Parallelize Shopify deletions using errgroup with bounded concurrency
	if s.shopifyClient != nil {
		g, _ := errgroup.WithContext(ctx)
		g.SetLimit(5) // Respect rate limits while gaining parallelism

		for _, cust := range customers {
			if cust.ExternalID != nil && *cust.ExternalID != "" {
				c := cust // shadow for closure
				g.Go(func() error {
					extID, _ := strconv.ParseInt(*c.ExternalID, 10, 64)
					if extID > 0 {
						if err := s.shopifyClient.DeleteCustomer(extID); err != nil {
							// Log but don't fail the whole group to maintain existing behavior
							log.Printf("BulkDelete: Failed to sync customer deletion to Shopify for %d: %v", c.ID, err)
						}
					}
					return nil
				})
			}
		}
		// Wait for all Shopify deletions to finish (they all return nil)
		_ = g.Wait()
	}

	// 3. Perform a single batch database delete
	if err := s.repo.BulkDelete(ctx, ids); err != nil {
		return fmt.Errorf("BulkDelete: failed to delete customers from database: %w", err)
	}

	return nil
}

func (s *CustomerService) ExportMetaCSV(ctx context.Context, boughtOnly bool) ([]byte, error) {
	minOrders := 0
	if boughtOnly {
		minOrders = 1
	}

	// Fetch all matching customers (limit -1 or large number)
	customers, _, err := s.repo.List(ctx, "", "updated_at", "DESC", "", 0, 0, minOrders, "", "", "", "", "", false, false, false, 0, 1000000)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch customers: %w", err)
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Meta-recognized headers: phone, email, fn, ln
	if err := writer.Write([]string{"phone", "email", "fn", "ln"}); err != nil {
		return nil, err
	}

	seenPhones := make(map[string]bool)

	for _, c := range customers {
		phone := s.cleanMetaPhone(c.PhoneNumber)
		// Deduplicate by phone
		if phone == "" || seenPhones[phone] {
			continue
		}
		seenPhones[phone] = true

		email := ""
		if c.Email != nil {
			email = strings.ToLower(strings.TrimSpace(*c.Email))
		}

		fn := ""
		if c.FirstName != nil {
			fn = strings.TrimSpace(*c.FirstName)
		}

		ln := ""
		if c.LastName != nil {
			ln = strings.TrimSpace(*c.LastName)
		}

		if err := writer.Write([]string{phone, email, fn, ln}); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), nil
}

func (s *CustomerService) cleanMetaPhone(phone string) string {
	// Replicate Python: re.sub(r"\D", "", phone)
	cleaned := nonDigitRegex.ReplaceAllString(phone, "")

	if len(cleaned) == 10 {
		return "91" + cleaned
	}
	if len(cleaned) == 12 && strings.HasPrefix(cleaned, "91") {
		return cleaned
	}
	return ""
}
