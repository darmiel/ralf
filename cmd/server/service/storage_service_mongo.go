package service

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// make sure MongoStorageService implements StorageService
var _ StorageService = (*MongoStorageService)(nil)

// MongoStorageService is an implementation of StorageService for MongoDB.
type MongoStorageService struct {
	client   *mongo.Client
	db       *mongo.Database
	flowColl *mongo.Collection
	histColl *mongo.Collection
}

// NewMongoStorageService creates a new MongoStorageService.
func NewMongoStorageService(uri, dbName string) *MongoStorageService {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	return &MongoStorageService{
		client:   client,
		db:       client.Database(dbName),
		flowColl: client.Database(dbName).Collection("flows"),
		histColl: client.Database(dbName).Collection("history"),
	}
}

// GetFlow retrieves a flow by its ID.
func (m *MongoStorageService) GetFlow(ctx context.Context, flowID string) (*SavedFlow, error) {
	var flow SavedFlow
	if err := m.flowColl.FindOne(ctx, bson.M{"flow_id": flowID}).Decode(&flow); err != nil {
		return nil, err
	}
	return &flow, nil
}

func (m *MongoStorageService) GetFlowJSON(ctx context.Context, flowID string) (interface{}, error) {
	var flow SavedFlow
	if err := m.flowColl.FindOne(ctx, bson.M{"flow_id": flowID}).Decode(&flow); err != nil {
		return nil, err
	}
	return flow, nil
}

func (m *MongoStorageService) GetFlowHistory(ctx context.Context, flowID string, limit int) ([]History, error) {
	cur, err := m.histColl.Find(ctx, bson.M{"flow_id": flowID}, options.Find().SetSort(bson.M{"timestamp": -1}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	var history []History
	if err := cur.All(ctx, &history); err != nil {
		return nil, err
	}
	return history, nil
}

func (m *MongoStorageService) DeleteFlow(ctx context.Context, flowID string) error {
	_, err := m.flowColl.DeleteOne(ctx, bson.M{"flow_id": flowID})
	return err
}

func (m *MongoStorageService) GetFlowsByUser(ctx context.Context, userID string) ([]SavedFlow, error) {
	cur, err := m.flowColl.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	var flows []SavedFlow
	if err := cur.All(ctx, &flows); err != nil {
		return nil, err
	}
	return flows, nil
}

func (m *MongoStorageService) SaveFlow(ctx context.Context, flow *SavedFlow) error {
	_, err := m.flowColl.UpdateOne(ctx, bson.M{"flow_id": flow.FlowID}, bson.M{"$set": flow}, options.Update().SetUpsert(true))
	return err
}

func (m *MongoStorageService) SaveHistory(ctx context.Context, history History) error {
	_, err := m.histColl.InsertOne(ctx, history)
	return err
}
