package service

import (
	"errors"
	"strings"

	"mi-tech/internal/domain/b2b/entity"
)

// Payment Terms functions
func (s *B2BService) ListPaymentTerms() ([]entity.B2BPaymentTerm, error) {
	return s.repo.ListPaymentTerms()
}

func (s *B2BService) CreatePaymentTerm(term *entity.B2BPaymentTerm) error {
	term.Name = strings.TrimSpace(term.Name)
	if term.Name == "" {
		return errors.New("term name is required")
	}
	if term.DueDays < 0 {
		return errors.New("due days cannot be negative")
	}
	return s.repo.CreatePaymentTerm(term)
}
