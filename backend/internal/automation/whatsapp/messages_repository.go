package whatsapp

import (
	"database/sql"
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
func (r *MessagesRepository) GetMessages(storeID string) ([]AutomationMessage, error) {
	query := `
		SELECT m.id, m.store_id, m.template_id, t.template_name, m.order_id, m.phone_number, m.message_id, m.status, m.sent_at, m.delivered_at, m.read_at, m.error_message 
		FROM automation_messages m
		LEFT JOIN automation_templates t ON m.template_id = t.id
		WHERE m.store_id = $1 
		ORDER BY m.sent_at DESC`
	rows, err := r.db.Query(query, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []AutomationMessage
	for rows.Next() {
		var m AutomationMessage
		err := rows.Scan(&m.ID, &m.StoreID, &m.TemplateID, &m.TemplateName, &m.OrderID, &m.PhoneNumber, &m.MessageID, &m.Status, &m.SentAt, &m.DeliveredAt, &m.ReadAt, &m.ErrorMessage)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func (r *MessagesRepository) GetAutomationMetrics(storeID string) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	var sent, delivered, read, failed int
	err := r.db.QueryRow(`SELECT 
		COUNT(*) FILTER (WHERE status != 'failed'),
		COUNT(*) FILTER (WHERE status = 'delivered' OR status = 'read'),
		COUNT(*) FILTER (WHERE status = 'read'),
		COUNT(*) FILTER (WHERE status = 'failed')
		FROM automation_messages WHERE store_id = $1`, storeID).Scan(&sent, &delivered, &read, &failed)
	if err != nil {
		return nil, err
	}

	metrics["sent"] = sent
	metrics["delivered"] = delivered
	metrics["read"] = read
	metrics["failed"] = failed

	var triggered int
	err = r.db.QueryRow(`SELECT COUNT(*) FROM automation_messages WHERE store_id = $1`, storeID).Scan(&triggered)
	if err != nil {
		return nil, err
	}
	metrics["triggered"] = triggered

	readRate := 0.0
	if sent > 0 {
		readRate = (float64(read) / float64(sent)) * 100
	}
	metrics["read_rate"] = readRate

	return metrics, nil
}
