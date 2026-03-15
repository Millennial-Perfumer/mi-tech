package whatsapp

import (
	"database/sql"
	"fmt"
	"log"
)

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

func (r *TemplatesRepository) GetTemplates(storeID string, startDate, endDate string) ([]AutomationTemplate, error) {
	dateFilter := ""
	args := []interface{}{storeID}
	placeholderID := 2

	if startDate != "" {
		dateFilter += fmt.Sprintf(" AND sent_at >= $%d", placeholderID)
		args = append(args, startDate)
		placeholderID++
	}
	if endDate != "" {
		if len(endDate) == 10 {
			endDate += " 23:59:59"
		}
		dateFilter += fmt.Sprintf(" AND sent_at <= $%d", placeholderID)
		args = append(args, endDate)
		placeholderID++
	}

	query := fmt.Sprintf(`
		SELECT 
			t.id, t.store_id, t.template_name, t.language, t.category, t.body, t.header, t.footer, t.buttons, t.status, 
			COALESCE(t.meta_template_id, ''), t.created_at, t.updated_at,
			(SELECT COUNT(*) FROM automation_messages WHERE template_id = t.id AND status != 'failed' %s) as sent_count,
			(SELECT COUNT(*) FROM automation_messages WHERE template_id = t.id AND (status = 'delivered' OR status = 'read') %s) as delivered_count,
			(SELECT COUNT(*) FROM automation_messages WHERE template_id = t.id AND status = 'read' %s) as read_count
		FROM automation_templates t
		WHERE t.store_id = $1`, dateFilter, dateFilter, dateFilter)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []AutomationTemplate
	for rows.Next() {
		var t AutomationTemplate
		err := rows.Scan(
			&t.ID, &t.StoreID, &t.TemplateName, &t.Language, &t.Category, &t.Body, &t.Header, &t.Footer, &t.Buttons, &t.Status, 
			&t.MetaTemplateID, &t.CreatedAt, &t.UpdatedAt, &t.SentCount, &t.DeliveredCount, &t.ReadCount,
		)
		if err != nil {
			log.Printf("Error scanning template row: %v", err)
			return nil, err
		}
		templates = append(templates, t)
	}
	log.Printf("Repository: GetTemplates (filtered) returned %d rows", len(templates))
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
	query := `SELECT id, store_id, template_name, language, category, body, header, footer, buttons, status, COALESCE(meta_template_id, '') FROM automation_templates WHERE id = $1`
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
