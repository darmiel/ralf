package service

import (
	"context"
	"github.com/darmiel/ralf/pkg/model"
	"time"
)

// StorageService defines the interface for storage services.
type StorageService interface {
	GetFlow(ctx context.Context, flowID string) (*SavedFlow, error)
	GetFlowHistory(ctx context.Context, flowID string, limit int) ([]History, error)
	DeleteFlow(ctx context.Context, flowID string) error
	GetFlowsByUser(ctx context.Context, userID string) ([]SavedFlow, error)
	SaveFlow(ctx context.Context, flow *SavedFlow) error
	SaveHistory(ctx context.Context, history History) error
}

// SavedFlow represents a flow model stored in the database.
type SavedFlow struct {
	FlowID        string           `json:"flow_id" bson:"flow_id"`
	UserID        string           `json:"user_id" bson:"user_id"`
	Name          string           `json:"name" bson:"name"`
	Source        model.SomeSource `json:"source" bson:"source"`
	CacheDuration model.Duration   `json:"cache_duration" bson:"cache_duration"`
	Flows         []model.Flow     `json:"flows" bson:"flows"`
}

// History represents the history of actions performed on a flow.
type History struct {
	FlowID    string    `json:"flow_id" bson:"flow_id"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	Address   string    `json:"address" bson:"address"`
	Success   bool      `json:"success" bson:"success"`
	Debug     []string  `json:"debug" bson:"debug"`
	Action    string    `json:"action" bson:"action"`
}
