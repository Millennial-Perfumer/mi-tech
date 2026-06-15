package service

import (
	"errors"
	"fmt"
	"strings"

	"mi-tech/internal/domain/b2b/entity"
	"mi-tech/internal/domain/b2b/helper"
)

// Customers CRUD
func (s *B2BService) ListCustomers(search string) ([]entity.B2BCustomer, error) {
	return s.repo.ListCustomers(search)
}

func (s *B2BService) GetCustomerByID(id int64) (entity.B2BCustomer, error) {
	return s.repo.GetCustomerByID(id)
}

func (s *B2BService) CreateCustomer(cust *entity.B2BCustomer) error {
	if err := s.validateCustomer(cust); err != nil {
		return err
	}
	return s.repo.CreateCustomer(cust)
}

func (s *B2BService) UpdateCustomer(cust *entity.B2BCustomer) error {
	if cust.ID <= 0 {
		return errors.New("invalid customer ID for update")
	}
	if err := s.validateCustomer(cust); err != nil {
		return err
	}
	return s.repo.UpdateCustomer(cust)
}

func (s *B2BService) DeleteCustomer(id int64) error {
	return s.repo.DeleteCustomer(id)
}

func (s *B2BService) validateCustomer(cust *entity.B2BCustomer) error {
	cust.GSTIN = strings.TrimSpace(strings.ToUpper(cust.GSTIN))
	if !helper.IsValidGSTIN(cust.GSTIN) {
		return fmt.Errorf("invalid GSTIN format: %s", cust.GSTIN)
	}

	// Extract PAN (chars 3 to 12)
	pan := cust.GSTIN[2:12]
	cust.PAN = &pan

	// Resolve State & State Code
	stateCode := cust.GSTIN[0:2]
	cust.StateCode = stateCode

	stateName, err := s.getStateNameByCode(stateCode)
	if err != nil {
		return fmt.Errorf("failed to resolve state for code %s: %w", stateCode, err)
	}
	cust.State = stateName

	if cust.LegalName == "" {
		return errors.New("legal name is required")
	}
	if cust.BillingAddress == "" {
		return errors.New("billing address is required")
	}

	return nil
}

func (s *B2BService) getStateNameByCode(code string) (string, error) {
	var name string
	err := s.db.Table("gst_state_codes").Where("code = ?", code).Pluck("name", &name).Error
	if err != nil {
		return "", err
	}
	if name == "" {
		return "", fmt.Errorf("state code %s not found", code)
	}
	return name, nil
}
