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
	OrderID      int64      `json:"order_id"`
	OrderNumber  string     `json:"order_number"`
	CustomerName string     `json:"customer_name"`
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
	if m.SentAt.IsZero() {
		m.SentAt = time.Now().UTC()
	} else {
		m.SentAt = m.SentAt.UTC()
	}
	query := `
		INSERT INTO automation_messages (store_id, template_id, order_id, phone_number, message_id, status, sent_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
	var id int
	err := r.db.QueryRow(query, m.StoreID, m.TemplateID, m.OrderID, m.PhoneNumber, m.MessageID, m.Status, m.SentAt).Scan(&id)
	return id, err
}

func (r *MessagesRepository) HasSentTemplate(orderID int64, templateID int, since time.Time) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM automation_messages WHERE order_id = $1 AND template_id = $2 AND status != 'failed'"
	args := []interface{}{orderID, templateID}

	if !since.IsZero() {
		query += " AND sent_at > $3"
		args = append(args, since)
	}

	err := r.db.QueryRow(query, args...).Scan(&count)
	return count > 0, err
}

func (r *MessagesRepository) UpdateMessageStatus(messageID, status string) error {
	var query string
	now := time.Now().UTC()
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

func (r *MessagesRepository) GetMessagesByOrderID(orderID int64) ([]AutomationMessage, error) {
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

func (r *MessagesRepository) GetMessages(storeID string, startDate, endDate *time.Time, search string, templateName string, limit, offset int) ([]AutomationMessage, error) {
	query := `
		SELECT 
			m.id, m.store_id, m.template_id, t.template_name, 
			m.order_id, o.order_number, o.customer_name,
			m.phone_number, m.message_id, m.status, m.sent_at, m.delivered_at, m.read_at, m.error_message 
		FROM automation_messages m
		LEFT JOIN automation_templates t ON m.template_id = t.id
		LEFT JOIN orders o ON m.order_id = o.id
		WHERE m.store_id = $1`

	args := []interface{}{storeID}
	placeholderID := 2

	if startDate != nil {
		query += fmt.Sprintf(" AND m.sent_at >= $%d", placeholderID)
		args = append(args, *startDate)
		placeholderID++
	}
	if endDate != nil {
		query += fmt.Sprintf(" AND m.sent_at <= $%d", placeholderID)
		args = append(args, *endDate)
		placeholderID++
	}
	if search != "" {
		searchTerm := "%" + search + "%"
		query += fmt.Sprintf(" AND (m.order_id::TEXT ILIKE $%d OR o.order_number ILIKE $%d OR m.phone_number ILIKE $%d)", placeholderID, placeholderID, placeholderID)
		args = append(args, searchTerm)
		placeholderID++
	}
	if templateName != "" {
		query += fmt.Sprintf(" AND t.template_name = $%d", placeholderID)
		args = append(args, templateName)
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
		var templateNameVal, orderNumber, customerName, errorMsg sql.NullString
		err := rows.Scan(
			&m.ID, &m.StoreID, &m.TemplateID, &templateNameVal,
			&m.OrderID, &orderNumber, &customerName,
			&m.PhoneNumber, &m.MessageID, &m.Status, &m.SentAt,
			&m.DeliveredAt, &m.ReadAt, &errorMsg,
		)
		if err != nil {
			return nil, err
		}
		m.TemplateName = templateNameVal.String
		m.OrderNumber = orderNumber.String
		m.CustomerName = customerName.String
		m.ErrorMessage = errorMsg.String
		messages = append(messages, m)
	}
	return messages, nil
}

func (r *MessagesRepository) GetActiveTemplateNamesForFilter(storeID string, startDate, endDate *time.Time, search string) ([]string, error) {
	query := `
		SELECT DISTINCT t.template_name
		FROM automation_messages m
		JOIN automation_templates t ON m.template_id = t.id
		LEFT JOIN orders o ON m.order_id = o.id
		WHERE m.store_id = $1`

	args := []interface{}{storeID}
	placeholderID := 2

	if startDate != nil {
		query += fmt.Sprintf(" AND m.sent_at >= $%d", placeholderID)
		args = append(args, *startDate)
		placeholderID++
	}
	if endDate != nil {
		query += fmt.Sprintf(" AND m.sent_at <= $%d", placeholderID)
		args = append(args, *endDate)
		placeholderID++
	}
	if search != "" {
		searchTerm := "%" + search + "%"
		query += fmt.Sprintf(" AND (m.order_id::TEXT ILIKE $%d OR o.order_number ILIKE $%d OR m.phone_number ILIKE $%d)", placeholderID, placeholderID, placeholderID)
		args = append(args, searchTerm)
		placeholderID++
	}

	query += " ORDER BY t.template_name"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

func (r *MessagesRepository) GetMessagesCount(storeID string, startDate, endDate *time.Time, search string, templateName string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM automation_messages m
		LEFT JOIN orders o ON m.order_id = o.id
		LEFT JOIN automation_templates t ON t.id = m.template_id
		WHERE m.store_id = $1`
	args := []interface{}{storeID}
	placeholderID := 2

	if startDate != nil {
		query += fmt.Sprintf(" AND m.sent_at >= $%d", placeholderID)
		args = append(args, *startDate)
		placeholderID++
	}
	if endDate != nil {
		query += fmt.Sprintf(" AND m.sent_at <= $%d", placeholderID)
		args = append(args, *endDate)
		placeholderID++
	}
	if search != "" {
		searchTerm := "%" + search + "%"
		query += fmt.Sprintf(" AND (m.order_id::TEXT ILIKE $%d OR o.order_number ILIKE $%d OR m.phone_number ILIKE $%d)", placeholderID, placeholderID, placeholderID)
		args = append(args, searchTerm)
		placeholderID++
	}
	if templateName != "" {
		query += fmt.Sprintf(" AND t.template_name = $%d", placeholderID)
		args = append(args, templateName)
		placeholderID++
	}

	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	return count, err
}

func (r *MessagesRepository) GetAutomationMetrics(storeID string, startDate, endDate *time.Time) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	whereClause := "WHERE 1=1"
	args := []interface{}{}
	placeholderID := 1

	if storeID != "" {
		whereClause += fmt.Sprintf(" AND store_id = $%d", placeholderID)
		args = append(args, storeID)
		placeholderID++
	}

	if startDate != nil {
		whereClause += fmt.Sprintf(" AND sent_at >= $%d", placeholderID)
		args = append(args, *startDate)
		placeholderID++
	}
	if endDate != nil {
		whereClause += fmt.Sprintf(" AND sent_at <= $%d", placeholderID)
		args = append(args, *endDate)
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

func (r *MessagesRepository) GetTriggeredCount(storeID string, startDate, endDate *time.Time) (int, error) {
	metrics, err := r.GetAutomationMetrics(storeID, startDate, endDate)
	if err != nil {
		return 0, err
	}
	return metrics["triggered"].(int), nil
}

func (r *MessagesRepository) GetFailedCount(storeID string, startDate, endDate *time.Time) (int, error) {
	metrics, err := r.GetAutomationMetrics(storeID, startDate, endDate)
	if err != nil {
		return 0, err
	}
	return metrics["failed"].(int), nil
}

func (r *MessagesRepository) GetOrderLineItems(orderID int64) ([]entity.LineItem, error) {
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
