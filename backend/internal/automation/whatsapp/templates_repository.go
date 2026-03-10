package whatsapp

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"
)

type AutomationTemplate struct {
	ID             int              `json:"id"`
	StoreID        string           `json:"store_id"`
	TemplateName   string           `json:"template_name"`
	Language       string           `json:"language"`
	Category       string           `json:"category"`
	Body           string           `json:"body"`
	Header         *json.RawMessage `json:"header,omitempty"`
	Footer         *string          `json:"footer,omitempty"`
	Buttons        *json.RawMessage `json:"buttons,omitempty"`
	Status         string           `json:"status"`
	MetaTemplateID string           `json:"meta_template_id"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

type Trigger struct {
	ID           int       `json:"id"`
	StoreID      string    `json:"store_id"`
	WebhookTopic string    `json:"webhook_topic"`
	TemplateID   int       `json:"template_id"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
}

type TemplatesRepository struct {
	db *sql.DB
}

func NewTemplatesRepository(db *sql.DB) *TemplatesRepository {
	return &TemplatesRepository{db: db}
}

func (r *TemplatesRepository) SaveTemplate(t AutomationTemplate) (int, error) {
	query := `
		INSERT INTO automation_templates (store_id, template_name, language, category, body, header, footer, buttons, status, meta_template_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`
	var id int
	err := r.db.QueryRow(query, t.StoreID, t.TemplateName, t.Language, t.Category, t.Body, t.Header, t.Footer, t.Buttons, t.Status, t.MetaTemplateID).Scan(&id)
	if err != nil {
		log.Printf("Repository: Error in SaveTemplate Query: %v", err)
	} else {
		log.Printf("Repository: SaveTemplate successful, id: %d", id)
	}
	return id, err
}

func (r *TemplatesRepository) GetTemplates(storeID string) ([]AutomationTemplate, error) {
	query := `SELECT id, store_id, template_name, language, category, body, header, footer, buttons, status, meta_template_id, created_at, updated_at 
	          FROM automation_templates WHERE store_id = $1`
	rows, err := r.db.Query(query, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []AutomationTemplate
	for rows.Next() {
		var t AutomationTemplate
		err := rows.Scan(&t.ID, &t.StoreID, &t.TemplateName, &t.Language, &t.Category, &t.Body, &t.Header, &t.Footer, &t.Buttons, &t.Status, &t.MetaTemplateID, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning template row: %v", err)
			return nil, err
		}
		templates = append(templates, t)
	}
	log.Printf("Repository: GetTemplates returned %d rows", len(templates))
	return templates, nil
}

func (r *TemplatesRepository) UpdateStatus(templateName, status string) error {
	query := `UPDATE automation_templates SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE template_name = $2`
	_, err := r.db.Exec(query, status, templateName)
	return err
}

func (r *TemplatesRepository) SaveTrigger(tr Trigger) error {
	query := `
		INSERT INTO automation_triggers (store_id, webhook_topic, template_id, enabled)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING`
	_, err := r.db.Exec(query, tr.StoreID, tr.WebhookTopic, tr.TemplateID, tr.Enabled)
	return err
}

func (r *TemplatesRepository) GetTriggerByTopic(storeID, topic string) (*Trigger, error) {
	query := `SELECT id, store_id, webhook_topic, template_id, enabled FROM automation_triggers 
	          WHERE store_id = $1 AND webhook_topic = $2 AND enabled = true`
	var t Trigger
	err := r.db.QueryRow(query, storeID, topic).Scan(&t.ID, &t.StoreID, &t.WebhookTopic, &t.TemplateID, &t.Enabled)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

func (r *TemplatesRepository) GetTriggers(storeID string) ([]Trigger, error) {
	query := `SELECT id, store_id, webhook_topic, template_id, enabled, created_at FROM automation_triggers WHERE store_id = $1`
	rows, err := r.db.Query(query, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var triggers []Trigger
	for rows.Next() {
		var tr Trigger
		err := rows.Scan(&tr.ID, &tr.StoreID, &tr.WebhookTopic, &tr.TemplateID, &tr.Enabled, &tr.CreatedAt)
		if err != nil {
			return nil, err
		}
		triggers = append(triggers, tr)
	}
	return triggers, nil
}

func (r *TemplatesRepository) GetTemplateByID(id int) (*AutomationTemplate, error) {
	query := `SELECT id, store_id, template_name, language, category, body, header, footer, buttons, status, meta_template_id FROM automation_templates WHERE id = $1`
	var t AutomationTemplate
	err := r.db.QueryRow(query, id).Scan(&t.ID, &t.StoreID, &t.TemplateName, &t.Language, &t.Category, &t.Body, &t.Header, &t.Footer, &t.Buttons, &t.Status, &t.MetaTemplateID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

func (r *TemplatesRepository) UpdateTemplate(t AutomationTemplate) error {
	query := `UPDATE automation_templates SET template_name = $1, language = $2, category = $3, body = $4, header = $5, footer = $6, buttons = $7, status = $8 WHERE id = $9 AND store_id = $10`
	_, err := r.db.Exec(query, t.TemplateName, t.Language, t.Category, t.Body, t.Header, t.Footer, t.Buttons, t.Status, t.ID, t.StoreID)
	return err
}

func (r *TemplatesRepository) DeleteTemplate(id int, storeID string) error {
	query := `DELETE FROM automation_templates WHERE id = $1 AND store_id = $2`
	_, err := r.db.Exec(query, id, storeID)
	return err
}

func (r *TemplatesRepository) UpdateTrigger(id int, storeID string, enabled bool) error {
	query := `UPDATE automation_triggers SET enabled = $1 WHERE id = $2 AND store_id = $3`
	_, err := r.db.Exec(query, enabled, id, storeID)
	return err
}

func (r *TemplatesRepository) DeleteTrigger(id int, storeID string) error {
	query := `DELETE FROM automation_triggers WHERE id = $1 AND store_id = $2`
	_, err := r.db.Exec(query, id, storeID)
	return err
}

func (r *TemplatesRepository) DeleteTriggersByTemplateID(templateID int, storeID string) error {
	query := `DELETE FROM automation_triggers WHERE template_id = $1 AND store_id = $2`
	_, err := r.db.Exec(query, templateID, storeID)
	return err
}
