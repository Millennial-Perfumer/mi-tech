package repository

import (
	"encoding/json"
	"mi-tech/internal/domain/ai/entity"
)

// AIConversationRepository defines operations for user AI chats.
type AIConversationRepository interface {
	CreateConversation(userID int64, title string) (*entity.AIConversation, error)
	ListConversations(userID int64, limit int) ([]entity.AIConversation, error)
	GetConversation(id, userID int64) (*entity.AIConversation, error)
	DeleteConversation(id, userID int64) error
	UpdateTitle(id int64, title string) error

	AddMessage(conversationID int64, role, content string, metadata *json.RawMessage) error
	GetMessages(conversationID int64) ([]entity.AIMessage, error)
}

// AIMemoryRepository defines data access for AI persistent memory.
type AIMemoryRepository interface {
	Upsert(memory *entity.AIMemory) error
	List(category string) ([]entity.AIMemory, error)
	GetByKey(key string) (*entity.AIMemory, error)
	Delete(key string) error
}

// AIReadRepository defines the contract for aggregate data retrieval for AI analysis.
type AIReadRepository interface {
	GetRevenueSummary(startDate, endDate string) (entity.AIRevenueSummary, error)
	GetRevenueByChannel(startDate, endDate string) ([]entity.AIChannelRevenue, error)
	GetRevenueByState(startDate, endDate string) ([]entity.AIStateRevenue, error)
	GetDailyRevenueTrend(startDate, endDate string) ([]entity.AIDailyRevenue, error)
	GetTopProducts(startDate, endDate string, limit int) ([]entity.AIProductRank, error)
	GetProductPerformance(sku string) (entity.AIProductStats, error)
	GetCustomerSegmentation() (entity.AICustomerSegments, error)
	GetTopCustomers(limit int) ([]entity.AITopCustomer, error)
	GetInventorySnapshot() ([]entity.AIInventoryStatus, error)
	GetBusinessSnapshot() (entity.AIBusinessSnapshot, error)
	ExecuteRawQuery(query string) ([]map[string]interface{}, error)
	ListTables() ([]string, error)
	DescribeTable(tableName string) ([]map[string]interface{}, error)
}
