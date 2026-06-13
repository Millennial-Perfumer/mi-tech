package dto

import (
	"mi-tech/internal/domain/communication/entity"
)

type CreateTemplateRequest struct {
	Name     string                  `json:"name"`
	Language string                  `json:"language"`
	Category string                  `json:"category"`
	Header   *entity.TemplateHeader  `json:"header,omitempty"`
	Body     string                  `json:"body"`
	Footer   string                  `json:"footer,omitempty"`
	Buttons  []entity.TemplateButton `json:"buttons,omitempty"`
	Examples string                  `json:"examples,omitempty"`
}
