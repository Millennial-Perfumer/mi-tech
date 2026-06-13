package repository

import (
	"database/sql"
	"fmt"
	"log"
	"mi-tech/internal/domain/communication/entity"
	"time"
)

type TemplatesRepository interface {
	SaveTemplate(t entity.AutomationTemplate) (int, error)
	GetTemplates(storeID string, startDate, endDate *time.Time) ([]entity.AutomationTemplate, error)
	UpdateStatus(templateName, status string) error
	SaveTrigger(tr entity.Trigger) error
	GetTriggerByTopic(storeID, topic string) (*entity.Trigger, error)
	GetTemplateByName(storeID, name string) (*entity.AutomationTemplate, error)
	GetTriggers(storeID string) ([]entity.Trigger, error)
	GetTemplateByID(id int) (*entity.AutomationTemplate, error)
	UpdateTemplate(t entity.AutomationTemplate) error
	DeleteTemplate(id int, storeID string) error
	UpdateTrigger(id int, storeID string, enabled bool) error
	DeleteTrigger(id int, storeID string) error
	DeleteTriggersByTemplateID(templateID int, storeID string) error
	UpsertMetaTemplate(t entity.AutomationTemplate) (int, error)
	GetEvents() ([]entity.AutomationEvent, error)
	SaveEvent(e entity.AutomationEvent) error
	DeleteEvent(id int) error
}

type sqlTemplatesRepository struct {
	db *sql.DB
}

func NewTemplatesRepository(db *sql.DB) TemplatesRepository {
	return &sqlTemplatesRepository{db: db}
}

func (r *sqlTemplatesRepository) SaveTemplate(t entity.AutomationTemplate) (int, error) {
	query := `
		INSERT INTO automation_templates (store_id, template_name, language, category, body, header, footer, buttons, status, meta_template_id, variable_mappings)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`
	var id int
	err := r.db.QueryRow(query, t.StoreID, t.TemplateName, t.Language, t.Category, t.Body, t.Header, t.Footer, t.Buttons, t.Status, t.MetaTemplateID, t.VariableMappings).Scan(&id)
	if err != nil {
		log.Printf("Repository: Error in SaveTemplate Query: %v", err)
	} else {
		log.Printf("Repository: SaveTemplate successful, id: %d", id)
	}
	return id, err
}

func (r *sqlTemplatesRepository) GetTemplates(storeID string, startDate, endDate *time.Time) ([]entity.AutomationTemplate, error) {
	args := []interface{}{storeID}
	placeholderID := 2

	dateFilter := ""
	if startDate != nil {
		dateFilter += fmt.Sprintf(" AND sent_at >= $%d", placeholderID)
		args = append(args, *startDate)
		placeholderID++
	}
	if endDate != nil {
		dateFilter += fmt.Sprintf(" AND sent_at <= $%d", placeholderID)
		args = append(args, *endDate)
		placeholderID++
	}

	query := fmt.Sprintf(`
		SELECT 
			t.id, t.store_id, t.template_name, t.language, t.category, t.body, t.header, t.footer, t.buttons, t.status, 
			COALESCE(t.meta_template_id, ''), t.variable_mappings, t.created_at, t.updated_at,
			COALESCE(m.sent_count, 0),
			COALESCE(m.delivered_count, 0),
			COALESCE(m.read_count, 0)
		FROM automation_templates t
		LEFT JOIN (
			SELECT 
				template_id,
				COUNT(*) FILTER (WHERE status != 'failed') as sent_count,
				COUNT(*) FILTER (WHERE status = 'delivered' OR status = 'read') as delivered_count,
				COUNT(*) FILTER (WHERE status = 'read') as read_count
			FROM automation_messages
			WHERE 1=1 %s
			GROUP BY template_id
		) m ON t.id = m.template_id
		WHERE t.store_id = $1 AND t.status != 'ARCHIVED'`, dateFilter)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []entity.AutomationTemplate
	for rows.Next() {
		var t entity.AutomationTemplate
		err := rows.Scan(
			&t.ID, &t.StoreID, &t.TemplateName, &t.Language, &t.Category, &t.Body, &t.Header, &t.Footer, &t.Buttons, &t.Status,
			&t.MetaTemplateID, &t.VariableMappings, &t.CreatedAt, &t.UpdatedAt, &t.SentCount, &t.DeliveredCount, &t.ReadCount,
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

func (r *sqlTemplatesRepository) UpdateStatus(templateName, status string) error {
	query := `UPDATE automation_templates SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE template_name = $2`
	_, err := r.db.Exec(query, status, templateName)
	return err
}

func (r *sqlTemplatesRepository) SaveTrigger(tr entity.Trigger) error {
	query := `
		INSERT INTO automation_triggers (store_id, webhook_topic, template_id, enabled)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING`
	_, err := r.db.Exec(query, tr.StoreID, tr.WebhookTopic, tr.TemplateID, tr.Enabled)
	return err
}

func (r *sqlTemplatesRepository) GetTriggerByTopic(storeID, topic string) (*entity.Trigger, error) {
	query := `SELECT id, store_id, webhook_topic, template_id, enabled FROM automation_triggers 
	          WHERE store_id = $1 AND webhook_topic = $2 AND enabled = true`
	var t entity.Trigger
	err := r.db.QueryRow(query, storeID, topic).Scan(&t.ID, &t.StoreID, &t.WebhookTopic, &t.TemplateID, &t.Enabled)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

func (r *sqlTemplatesRepository) GetTemplateByName(storeID, name string) (*entity.AutomationTemplate, error) {
	query := `SELECT id, store_id, template_name, language, category, body, header, footer, buttons, status, COALESCE(meta_template_id, ''), variable_mappings 
	          FROM automation_templates 
	          WHERE store_id = $1 AND template_name = $2`
	var t entity.AutomationTemplate
	err := r.db.QueryRow(query, storeID, name).Scan(&t.ID, &t.StoreID, &t.TemplateName, &t.Language, &t.Category, &t.Body, &t.Header, &t.Footer, &t.Buttons, &t.Status, &t.MetaTemplateID, &t.VariableMappings)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

func (r *sqlTemplatesRepository) GetTriggers(storeID string) ([]entity.Trigger, error) {
	query := `
		SELECT 
			tr.id, tr.store_id, tr.webhook_topic, tr.template_id, tr.enabled, tr.created_at,
			t.template_name, t.body, t.status
		FROM automation_triggers tr
		JOIN automation_templates t ON tr.template_id = t.id
		WHERE tr.store_id = $1`
	rows, err := r.db.Query(query, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var triggers []entity.Trigger
	for rows.Next() {
		var tr entity.Trigger
		err := rows.Scan(&tr.ID, &tr.StoreID, &tr.WebhookTopic, &tr.TemplateID, &tr.Enabled, &tr.CreatedAt, &tr.TemplateName, &tr.TemplateBody, &tr.TemplateStatus)
		if err != nil {
			return nil, err
		}
		triggers = append(triggers, tr)
	}
	return triggers, nil
}

func (r *sqlTemplatesRepository) GetTemplateByID(id int) (*entity.AutomationTemplate, error) {
	query := `SELECT id, store_id, template_name, language, category, body, header, footer, buttons, status, COALESCE(meta_template_id, ''), variable_mappings FROM automation_templates WHERE id = $1`
	var t entity.AutomationTemplate
	err := r.db.QueryRow(query, id).Scan(&t.ID, &t.StoreID, &t.TemplateName, &t.Language, &t.Category, &t.Body, &t.Header, &t.Footer, &t.Buttons, &t.Status, &t.MetaTemplateID, &t.VariableMappings)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

func (r *sqlTemplatesRepository) UpdateTemplate(t entity.AutomationTemplate) error {
	query := `UPDATE automation_templates SET template_name = $1, language = $2, category = $3, body = $4, header = $5, footer = $6, buttons = $7, status = $8, variable_mappings = $9 WHERE id = $10 AND store_id = $11`
	_, err := r.db.Exec(query, t.TemplateName, t.Language, t.Category, t.Body, t.Header, t.Footer, t.Buttons, t.Status, t.VariableMappings, t.ID, t.StoreID)
	return err
}

func (r *sqlTemplatesRepository) DeleteTemplate(id int, storeID string) error {
	query := `DELETE FROM automation_templates WHERE id = $1 AND store_id = $2`
	_, err := r.db.Exec(query, id, storeID)
	return err
}

func (r *sqlTemplatesRepository) UpdateTrigger(id int, storeID string, enabled bool) error {
	query := `UPDATE automation_triggers SET enabled = $1 WHERE id = $2 AND store_id = $3`
	_, err := r.db.Exec(query, enabled, id, storeID)
	return err
}

func (r *sqlTemplatesRepository) DeleteTrigger(id int, storeID string) error {
	query := `DELETE FROM automation_triggers WHERE id = $1 AND store_id = $2`
	_, err := r.db.Exec(query, id, storeID)
	return err
}

func (r *sqlTemplatesRepository) DeleteTriggersByTemplateID(templateID int, storeID string) error {
	query := `DELETE FROM automation_triggers WHERE template_id = $1 AND store_id = $2`
	_, err := r.db.Exec(query, templateID, storeID)
	return err
}

func (r *sqlTemplatesRepository) UpsertMetaTemplate(t entity.AutomationTemplate) (int, error) {
	// First check if it exists
	var existingID int
	checkQuery := `SELECT id FROM automation_templates WHERE store_id = $1 AND template_name = $2`
	err := r.db.QueryRow(checkQuery, t.StoreID, t.TemplateName).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Insert
		insertQuery := `
			INSERT INTO automation_templates (store_id, template_name, language, category, body, header, footer, buttons, status, meta_template_id, variable_mappings)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING id`
		err = r.db.QueryRow(insertQuery, t.StoreID, t.TemplateName, t.Language, t.Category, t.Body, t.Header, t.Footer, t.Buttons, t.Status, t.MetaTemplateID, t.VariableMappings).Scan(&existingID)
		if err != nil {
			return 0, fmt.Errorf("failed to insert meta template: %w", err)
		}
		return existingID, nil
	} else if err != nil {
		return 0, fmt.Errorf("failed to check existing template: %w", err)
	}

	// Update existing
	updateQuery := `
		UPDATE automation_templates 
		SET language = $1, category = $2, body = $3, header = $4, footer = $5, buttons = $6, status = $7, meta_template_id = $8, updated_at = CURRENT_TIMESTAMP
		WHERE id = $9 AND store_id = $10`
	_, err = r.db.Exec(updateQuery, t.Language, t.Category, t.Body, t.Header, t.Footer, t.Buttons, t.Status, t.MetaTemplateID, existingID, t.StoreID)
	if err != nil {
		return 0, fmt.Errorf("failed to update meta template: %w", err)
	}

	return existingID, nil
}

func (r *sqlTemplatesRepository) GetEvents() ([]entity.AutomationEvent, error) {
	query := `SELECT id, name, topic, description, created_at FROM automation_events ORDER BY created_at ASC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []entity.AutomationEvent
	for rows.Next() {
		var e entity.AutomationEvent
		err := rows.Scan(&e.ID, &e.Name, &e.Topic, &e.Description, &e.CreatedAt)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (r *sqlTemplatesRepository) SaveEvent(e entity.AutomationEvent) error {
	query := `INSERT INTO automation_events (name, topic, description) VALUES ($1, $2, $3) ON CONFLICT (topic) DO UPDATE SET name = $1, description = $3`
	_, err := r.db.Exec(query, e.Name, e.Topic, e.Description)
	if err != nil {
		log.Printf("Repository ERROR: SaveEvent failed for topic '%s': %v", e.Topic, err)
	} else {
		log.Printf("Repository SUCCESS: Saved event topic '%s'", e.Topic)
	}
	return err
}

func (r *sqlTemplatesRepository) DeleteEvent(id int) error {
	query := `DELETE FROM automation_events WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
