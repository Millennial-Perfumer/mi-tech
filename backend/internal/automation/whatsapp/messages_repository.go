package whatsapp

import (
	"database/sql"
	"fmt"
	"mi-tech/internal/entity"
	"time"
)

type AutomationMessage struct {
	ID           int        `json:"id"`
	StoreID      string     `json:"store_id"`
	TemplateID   int        `json:"template_id"`
	TemplateName string     `json:"template_name"`
	OrderID      string     `json:"order_id"`
	PhoneNumber  string     `json:"phone_number"`
	MessageID    string     `json:"message_id"`
	Status       string     `json:"status"`
	SentAt       time.Time  `json:"sent_at"`
	DeliveredAt  *time.Time `json:"delivered_at"`
	ReadAt       *time.Time `json:"read_at"`
	ErrorMessage string     `json:"error_message"`
}

type MessagesRepository struct {
	db *sql.DB
}

func NewMessagesRepository(db *sql.DB) *MessagesRepository {
	return &MessagesRepository{db: db}
}

func (r *MessagesRepository) SaveMessage(m AutomationMessage) (int, error) {
	query := `
		INSERT INTO automation_messages (store_id, template_id, order_id, phone_number, message_id, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`
	var id int
	err := r.db.QueryRow(query, m.StoreID, m.TemplateID, m.OrderID, m.PhoneNumber, m.MessageID, m.Status).Scan(&id)
	return id, err
}

func (r *MessagesRepository) HasSentTemplate(orderID string, templateID int) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM automation_messages WHERE order_id = $1 AND template_id = $2 AND status != 'failed'", orderID, templateID).Scan(&count)
	return count > 0, err
}

func (r *MessagesRepository) UpdateMessageStatus(messageID, status string) error {
	var query string
	now := time.Now()
	switch status {
	case "delivered":
		query = `UPDATE automation_messages SET status = $1, delivered_at = $2 WHERE message_id = $3`
		_, err := r.db.Exec(query, status, now, messageID)
		return err
	case "read":
		query = `UPDATE automation_messages SET status = $1, read_at = $2 WHERE message_id = $3`
		_, err := r.db.Exec(query, status, now, messageID)
		return err
	case "failed":
		query = `UPDATE automation_messages SET status = $1 WHERE message_id = $2`
		_, err := r.db.Exec(query, status, messageID)
		return err
	default:
		query = `UPDATE automation_messages SET status = $1 WHERE message_id = $2`
		_, err := r.db.Exec(query, status, messageID)
		return err
	}
}

func (r *MessagesRepository) GetMessagesByOrderID(orderID string) ([]AutomationMessage, error) {
	query := `SELECT id, store_id, template_id, order_id, phone_number, message_id, status, sent_at FROM automation_messages WHERE order_id = $1`
	rows, err := r.db.Query(query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []AutomationMessage
	for rows.Next() {
		var m AutomationMessage
		err := rows.Scan(&m.ID, &m.StoreID, &m.TemplateID, &m.OrderID, &m.PhoneNumber, &m.MessageID, &m.Status, &m.SentAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}
func (r *MessagesRepository) GetMessages(storeID string, startDate, endDate string, limit, offset int) ([]AutomationMessage, error) {
	query := `
		SELECT m.id, m.store_id, m.template_id, t.template_name, m.order_id, m.phone_number, m.message_id, m.status, m.sent_at, m.delivered_at, m.read_at, m.error_message 
		FROM automation_messages m
		LEFT JOIN automation_templates t ON m.template_id = t.id
		WHERE m.store_id = $1`
	
	args := []interface{}{storeID}
	placeholderID := 2

	if startDate != "" {
		query += fmt.Sprintf(" AND m.sent_at >= $%d", placeholderID)
		args = append(args, startDate)
		placeholderID++
	}
	if endDate != "" {
		// Append 23:59:59 to endDate if it's just a date
		if len(endDate) == 10 {
			endDate += " 23:59:59"
		}
		query += fmt.Sprintf(" AND m.sent_at <= $%d", placeholderID)
		args = append(args, endDate)
		placeholderID++
	}

	query += " ORDER BY m.sent_at DESC"
	
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", placeholderID, placeholderID+1)
		args = append(args, limit, offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []AutomationMessage
	for rows.Next() {
		var m AutomationMessage
		var templateName, errorMsg sql.NullString
		err := rows.Scan(&m.ID, &m.StoreID, &m.TemplateID, &templateName, &m.OrderID, &m.PhoneNumber, &m.MessageID, &m.Status, &m.SentAt, &m.DeliveredAt, &m.ReadAt, &errorMsg)
		if err != nil {
			return nil, err
		}
		m.TemplateName = templateName.String
		m.ErrorMessage = errorMsg.String
		messages = append(messages, m)
	}
	return messages, nil
}

func (r *MessagesRepository) GetMessagesCount(storeID string, startDate, endDate string) (int, error) {
	query := "SELECT COUNT(*) FROM automation_messages WHERE store_id = $1"
	args := []interface{}{storeID}
	placeholderID := 2

	if startDate != "" {
		query += fmt.Sprintf(" AND sent_at >= $%d", placeholderID)
		args = append(args, startDate)
		placeholderID++
	}
	if endDate != "" {
		if len(endDate) == 10 {
			endDate += " 23:59:59"
		}
		query += fmt.Sprintf(" AND sent_at <= $%d", placeholderID)
		args = append(args, endDate)
	}

	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	return count, err
}

func (r *MessagesRepository) GetAutomationMetrics(storeID string, startDate, endDate string) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	whereClause := "WHERE 1=1"
	args := []interface{}{}
	placeholderID := 1

	if storeID != "" {
		whereClause += fmt.Sprintf(" AND store_id = $%d", placeholderID)
		args = append(args, storeID)
		placeholderID++
	}

	if startDate != "" {
		whereClause += fmt.Sprintf(" AND sent_at >= $%d", placeholderID)
		args = append(args, startDate)
		placeholderID++
	}
	if endDate != "" {
		if len(endDate) == 10 {
			endDate += " 23:59:59"
		}
		whereClause += fmt.Sprintf(" AND sent_at <= $%d", placeholderID)
		args = append(args, endDate)
		placeholderID++
	}

	query := fmt.Sprintf(`SELECT 
		COUNT(*) FILTER (WHERE status != 'failed'),
		COUNT(*) FILTER (WHERE status = 'delivered' OR status = 'read'),
		COUNT(*) FILTER (WHERE status = 'read'),
		COUNT(*) FILTER (WHERE status = 'failed'),
		COUNT(*)
		FROM automation_messages %s`, whereClause)

	var sent, delivered, read, failed, triggered int
	err := r.db.QueryRow(query, args...).Scan(&sent, &delivered, &read, &failed, &triggered)
	if err != nil {
		return nil, err
	}
	
	metrics["sent"] = sent
	metrics["delivered"] = delivered
	metrics["read"] = read
	metrics["failed"] = failed
	metrics["triggered"] = triggered

	readRate := 0.0
	if sent > 0 {
		readRate = (float64(read) / float64(sent)) * 100
	}
	metrics["read_rate"] = readRate

	return metrics, nil
}

func (r *MessagesRepository) GetTriggeredCount(storeID string, startDate, endDate string) (int, error) {
	metrics, err := r.GetAutomationMetrics(storeID, startDate, endDate)
	if err != nil {
		return 0, err
	}
	return metrics["triggered"].(int), nil
}

func (r *MessagesRepository) GetFailedCount(storeID string, startDate, endDate string) (int, error) {
	metrics, err := r.GetAutomationMetrics(storeID, startDate, endDate)
	if err != nil {
		return 0, err
	}
	return metrics["failed"].(int), nil
}
func (r *MessagesRepository) GetOrderLineItems(orderID string) ([]entity.LineItem, error) {
	query := `SELECT id, order_id, product_id, variant_id, title, sku, hs_code, quantity, price, discount FROM order_line_items WHERE order_id = $1`
	rows, err := r.db.Query(query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []entity.LineItem
	for rows.Next() {
		var li entity.LineItem
		err := rows.Scan(&li.ID, &li.OrderID, &li.ProductID, &li.VariantID, &li.Title, &li.SKU, &li.HSCode, &li.Quantity, &li.Price, &li.Discount)
		if err != nil {
			return nil, err
		}
		items = append(items, li)
	}
	return items, nil
}
