package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"mi-tech/internal/entity"
	"mi-tech/internal/repository"
	"strconv"
	"strings"
	"time"
)

type CustomerService struct {
	repo      *repository.CustomerRepository
	orderRepo repository.OrderRepository
}

func NewCustomerService(repo *repository.CustomerRepository, orderRepo repository.OrderRepository) *CustomerService {
	return &CustomerService{
		repo:      repo,
		orderRepo: orderRepo,
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
	if _, ok := headerMap["Phone"]; ok { hasPhone = true }
	if _, ok := headerMap["Default Address Phone"]; ok { hasPhone = true }
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

		phone := s.normalizePhone(phoneRaw)
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
				TotalSpent:   0,
				TotalOrders:  0,
				SourceID:     sourceID,
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
	phone := s.normalizePhone(entity.DerefStr(order.CustomerPhone))
	if phone == "" {
		return nil
	}

	customer := &entity.Customer{
		PhoneNumber: phone,
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
		customer.TotalOrders = existing.TotalOrders
		customer.TotalSpent = existing.TotalSpent
		safeMerge(customer, existing)
		// Inherit created at
		customer.CreatedAt = existing.CreatedAt
	} else {
		customer.CreatedAt = time.Now()
	}

	// For now just upsert. In a more advanced version, we'd recalculate totals from orders table.
	return s.repo.UpsertByPhone(ctx, customer)
}

// UpsertFromWebhook securely updates or creates a customer entirely from a customer webhook payload.
func (s *CustomerService) UpsertFromWebhook(ctx context.Context, customer entity.Customer) error {
	if customer.PhoneNumber == "" {
		return fmt.Errorf("phone number is required for customer webhook upsert")
	}

	if customer.FirstName != nil {
		customer.FirstName = entity.StrPtr(s.toTitleCase(entity.DerefStr(customer.FirstName)))
	}
	if customer.LastName != nil {
		customer.LastName = entity.StrPtr(s.toTitleCase(entity.DerefStr(customer.LastName)))
	}

	existing, err := s.repo.GetByPhone(ctx, customer.PhoneNumber)
	if err == nil && existing != nil {
		safeMerge(&customer, existing)
		customer.CreatedAt = existing.CreatedAt
		if customer.TotalOrders == 0 && existing.TotalOrders > 0 { customer.TotalOrders = existing.TotalOrders }
		if customer.TotalSpent == 0 && existing.TotalSpent > 0 { customer.TotalSpent = existing.TotalSpent }
	} else {
		if customer.CreatedAt.IsZero() {
			customer.CreatedAt = time.Now()
		}
	}

	if customer.UpdatedAt.IsZero() {
		customer.UpdatedAt = time.Now()
	}

	return s.repo.UpsertByPhone(ctx, &customer)
}

func safeMerge(newCust, oldCust *entity.Customer) {
	if newCust.FirstName == nil { newCust.FirstName = oldCust.FirstName }
	if newCust.LastName == nil { newCust.LastName = oldCust.LastName }
	if newCust.Email == nil { newCust.Email = oldCust.Email }
	if newCust.Address1 == nil { newCust.Address1 = oldCust.Address1 }
	if newCust.Address2 == nil { newCust.Address2 = oldCust.Address2 }
	if newCust.City == nil { newCust.City = oldCust.City }
	if newCust.State == nil { newCust.State = oldCust.State }
	if newCust.Country == nil { newCust.Country = oldCust.Country }
	if newCust.ZipCode == nil { newCust.ZipCode = oldCust.ZipCode }
}

func (s *CustomerService) ListCustomers(ctx context.Context, search, sortBy, sortOrder string, page, pageSize int) ([]entity.Customer, int64, error) {
	if page < 1 { page = 1 }
	if pageSize < 1 { pageSize = 20 }
	offset := (page - 1) * pageSize

	dbSortBy := "updated_at"
	switch sortBy {
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

	return s.repo.List(ctx, search, dbSortBy, sortOrder, offset, pageSize)
}

func (s *CustomerService) DeleteAllCustomers(ctx context.Context) error {
	return s.repo.DeleteAll(ctx)
}

func (s *CustomerService) normalizePhone(p string) string {
	p = strings.TrimPrefix(p, "'")
	var sb strings.Builder
	for _, r := range p {
		if r >= '0' && r <= '9' {
			sb.WriteRune(r)
		}
	}
	digits := sb.String()
	if digits == "" {
		return ""
	}

	// Handle local prefix 0 (e.g., 06383173716 -> 6383173716)
	if strings.HasPrefix(digits, "0") && len(digits) == 11 {
		digits = digits[1:]
	}

	// Handle country code 91 (e.g., 916383173716 -> 6383173716)
	if strings.HasPrefix(digits, "91") && len(digits) == 12 {
		digits = digits[2:]
	}

	// If it's a 10-digit number, assume it's an Indian mobile number and add +91
	if len(digits) == 10 {
		return "+91" + digits
	}

	// Otherwise, return with a + prefix as it might already include a country code
	return "+" + digits
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
