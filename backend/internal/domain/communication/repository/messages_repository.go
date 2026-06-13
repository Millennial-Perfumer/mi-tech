package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"mi-tech/internal/domain/communication/entity"
	"time"
)

type MessagesRepository interface {
	GetDB() *sql.DB
	SaveMessage(m entity.AutomationMessage) (int, error)
	HasSentTemplate(orderID int64, templateID int, since time.Time) (bool, error)
	UpdateMessageStatus(messageID, status string) error
	GetMessagesByOrderID(orderID int64) ([]entity.AutomationMessage, error)
	GetMessages(storeID string, startDate, endDate *time.Time, search string, templateName string, limit, offset int) ([]entity.AutomationMessage, error)
	GetActiveTemplateNamesForFilter(storeID string, startDate, endDate *time.Time, search string) ([]string, error)
	GetMessagesCount(storeID string, startDate, endDate *time.Time, search string, templateName string) (int, error)
	GetAutomationMetrics(storeID string, startDate, endDate *time.Time) (map[string]interface{}, error)
	GetTriggeredCount(storeID string, startDate, endDate *time.Time) (int, error)
	GetFailedCount(storeID string, startDate, endDate *time.Time) (int, error)
	GetConversations() ([]entity.Conversation, error)
	GetChatMessages(conversationID int, limit, offset int) ([]entity.ChatMessage, error)
	UpsertConversation(phoneNumber, contactName, lastMessage string) (int, error)
	UpdateConversationMode(id int, mode string) error
	SaveChatMessage(m entity.ChatMessage) (int, error)
	GetConversationByPhone(phoneNumber string) (*entity.Conversation, error)
	GetIssuesSince(since time.Time) ([]entity.ChatMessage, error)
}

type sqlMessagesRepository struct {
	db *sql.DB
}

func NewMessagesRepository(db *sql.DB) MessagesRepository {
	return &sqlMessagesRepository{db: db}
}

func (r *sqlMessagesRepository) GetDB() *sql.DB {
	return r.db
}

func (r *sqlMessagesRepository) SaveMessage(m entity.AutomationMessage) (int, error) {
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

func (r *sqlMessagesRepository) HasSentTemplate(orderID int64, templateID int, since time.Time) (bool, error) {
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

func (r *sqlMessagesRepository) UpdateMessageStatus(messageID, status string) error {
	now := time.Now().UTC()

	// 1. Update automation_messages
	var queryAuto string
	switch status {
	case "delivered":
		queryAuto = `UPDATE automation_messages SET status = $1, delivered_at = $2 WHERE message_id = $3`
		_, _ = r.db.Exec(queryAuto, status, now, messageID)
	case "read":
		queryAuto = `UPDATE automation_messages SET status = $1, read_at = $2 WHERE message_id = $3`
		_, _ = r.db.Exec(queryAuto, status, now, messageID)
	default:
		queryAuto = `UPDATE automation_messages SET status = $1 WHERE message_id = $2`
		_, _ = r.db.Exec(queryAuto, status, messageID)
	}

	// 2. Update whatsapp_chat_messages (the one used in Chat Hub)
	queryChat := `UPDATE whatsapp_chat_messages SET status = $1 WHERE message_id = $2`
	_, err := r.db.Exec(queryChat, status, messageID)

	return err
}

func (r *sqlMessagesRepository) GetMessagesByOrderID(orderID int64) ([]entity.AutomationMessage, error) {
	query := `SELECT id, store_id, template_id, order_id, phone_number, message_id, status, sent_at FROM automation_messages WHERE order_id = $1`
	rows, err := r.db.Query(query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []entity.AutomationMessage
	for rows.Next() {
		var m entity.AutomationMessage
		err := rows.Scan(&m.ID, &m.StoreID, &m.TemplateID, &m.OrderID, &m.PhoneNumber, &m.MessageID, &m.Status, &m.SentAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func (r *sqlMessagesRepository) GetMessages(storeID string, startDate, endDate *time.Time, search string, templateName string, limit, offset int) ([]entity.AutomationMessage, error) {
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

	var messages []entity.AutomationMessage
	for rows.Next() {
		var m entity.AutomationMessage
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

func (r *sqlMessagesRepository) GetActiveTemplateNamesForFilter(storeID string, startDate, endDate *time.Time, search string) ([]string, error) {
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

func (r *sqlMessagesRepository) GetMessagesCount(storeID string, startDate, endDate *time.Time, search string, templateName string) (int, error) {
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

func (r *sqlMessagesRepository) GetAutomationMetrics(storeID string, startDate, endDate *time.Time) (map[string]interface{}, error) {
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

func (r *sqlMessagesRepository) GetTriggeredCount(storeID string, startDate, endDate *time.Time) (int, error) {
	metrics, err := r.GetAutomationMetrics(storeID, startDate, endDate)
	if err != nil {
		return 0, err
	}
	return metrics["triggered"].(int), nil
}

func (r *sqlMessagesRepository) GetFailedCount(storeID string, startDate, endDate *time.Time) (int, error) {
	metrics, err := r.GetAutomationMetrics(storeID, startDate, endDate)
	if err != nil {
		return 0, err
	}
	return metrics["failed"].(int), nil
}

func (r *sqlMessagesRepository) GetConversations() ([]entity.Conversation, error) {
	query := `SELECT id, phone_number, contact_name, last_message, last_message_at, mode, active_task_id, priority, created_at, updated_at 
	          FROM whatsapp_conversations ORDER BY last_message_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []entity.Conversation
	for rows.Next() {
		var c entity.Conversation
		var contactName, lastMessage sql.NullString
		err := rows.Scan(&c.ID, &c.PhoneNumber, &contactName, &lastMessage, &c.LastMessageAt, &c.Mode, &c.ActiveTaskID, &c.Priority, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		c.ContactName = contactName.String
		c.LastMessage = lastMessage.String
		conversations = append(conversations, c)
	}
	return conversations, nil
}

func (r *sqlMessagesRepository) GetChatMessages(conversationID int, limit, offset int) ([]entity.ChatMessage, error) {
	query := `SELECT id, conversation_id, message_id, text, type, direction, sender_role, status, is_issue, priority, sent_at, metadata 
	          FROM whatsapp_chat_messages WHERE conversation_id = $1 ORDER BY sent_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(query, conversationID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []entity.ChatMessage
	for rows.Next() {
		var m entity.ChatMessage
		var messageID, text, senderRole sql.NullString
		var metadata interface{}
		err := rows.Scan(&m.ID, &m.ConversationID, &messageID, &text, &m.Type, &m.Direction, &senderRole, &m.Status, &m.IsIssue, &m.Priority, &m.SentAt, &metadata)
		if err != nil {
			log.Printf("Scan error in GetChatMessages: %v", err)
			return nil, err
		}

		m.MessageID = messageID.String
		m.Text = text.String
		m.SenderRole = senderRole.String

		if metadata != nil {
			if bytes, ok := metadata.([]byte); ok {
				m.Metadata = json.RawMessage(bytes)
			}
		}

		messages = append(messages, m)
	}
	// Return in chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

func (r *sqlMessagesRepository) UpsertConversation(phoneNumber, contactName, lastMessage string) (int, error) {
	query := `
		INSERT INTO whatsapp_conversations (phone_number, contact_name, last_message, last_message_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (phone_number) DO UPDATE SET
			contact_name = CASE WHEN EXCLUDED.contact_name <> '' THEN EXCLUDED.contact_name ELSE whatsapp_conversations.contact_name END,
			last_message = EXCLUDED.last_message,
			last_message_at = EXCLUDED.last_message_at,
			updated_at = EXCLUDED.updated_at
		RETURNING id`
	var id int
	err := r.db.QueryRow(query, phoneNumber, contactName, lastMessage).Scan(&id)
	return id, err
}

func (r *sqlMessagesRepository) UpdateConversationMode(id int, mode string) error {
	query := `UPDATE whatsapp_conversations SET mode = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.Exec(query, mode, id)
	return err
}

func (r *sqlMessagesRepository) SaveChatMessage(m entity.ChatMessage) (int, error) {
	query := `
		INSERT INTO whatsapp_chat_messages (conversation_id, message_id, text, type, direction, sender_role, status, is_issue, priority, sent_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`
	var id int
	sentAt := m.SentAt
	if sentAt.IsZero() {
		sentAt = time.Now().UTC()
	}
	err := r.db.QueryRow(query, m.ConversationID, m.MessageID, m.Text, m.Type, m.Direction, m.SenderRole, m.Status, m.IsIssue, m.Priority, sentAt, m.Metadata).Scan(&id)
	return id, err
}

func (r *sqlMessagesRepository) GetConversationByPhone(phoneNumber string) (*entity.Conversation, error) {
	query := `SELECT id, phone_number, contact_name, last_message, last_message_at, mode, active_task_id, priority, created_at, updated_at 
	          FROM whatsapp_conversations WHERE phone_number = $1`
	var c entity.Conversation
	var contactName, lastMessage sql.NullString
	err := r.db.QueryRow(query, phoneNumber).Scan(&c.ID, &c.PhoneNumber, &contactName, &lastMessage, &c.LastMessageAt, &c.Mode, &c.ActiveTaskID, &c.Priority, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	c.ContactName = contactName.String
	c.LastMessage = lastMessage.String
	return &c, nil
}

func (r *sqlMessagesRepository) GetIssuesSince(since time.Time) ([]entity.ChatMessage, error) {
	query := `SELECT id, conversation_id, message_id, text, type, direction, sender_role, status, is_issue, priority, sent_at, metadata 
	          FROM whatsapp_chat_messages WHERE is_issue = true AND sent_at >= $1 ORDER BY sent_at DESC`
	rows, err := r.db.Query(query, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []entity.ChatMessage
	for rows.Next() {
		var m entity.ChatMessage
		var messageID, text, senderRole sql.NullString
		var metadata interface{}
		err := rows.Scan(&m.ID, &m.ConversationID, &messageID, &text, &m.Type, &m.Direction, &senderRole, &m.Status, &m.IsIssue, &m.Priority, &m.SentAt, &metadata)
		if err != nil {
			return nil, err
		}
		m.MessageID = messageID.String
		m.Text = text.String
		m.SenderRole = senderRole.String
		if metadata != nil {
			if bytes, ok := metadata.([]byte); ok {
				m.Metadata = json.RawMessage(bytes)
			}
		}
		messages = append(messages, m)
	}
	return messages, nil
}
