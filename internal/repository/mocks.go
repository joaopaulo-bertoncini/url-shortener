package repository

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ===== REDIS MOCKS =====

type MockStatusCmd struct {
	mock.Mock
	err error
}

func (m *MockStatusCmd) Err() error {
	return m.err
}

type MockStringCmd struct {
	mock.Mock
	val string
	err error
}

func (m *MockStringCmd) Result() (string, error) {
	return m.val, m.err
}

type MockIntCmd struct {
	mock.Mock
	val int64
	err error
}

func (m *MockIntCmd) Result() (int64, error) {
	return m.val, m.err
}

type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *MockStatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*MockStatusCmd)
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *MockStringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*MockStringCmd)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *MockIntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*MockIntCmd)
}

// ===== MONGO MOCKS =====

type MockMongoCollection struct {
	mock.Mock
}

func (m *MockMongoCollection) InsertOne(ctx context.Context, document interface{},
	opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (m *MockMongoCollection) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{},
	opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult {
	args := m.Called(ctx, filter, update)
	return args.Get(0).(*mongo.SingleResult)
}

func (m *MockMongoCollection) FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.SingleResult)
}

func (m *MockMongoCollection) DeleteOne(ctx context.Context, filter interface{},
	opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

type MockMongoClient struct {
	db map[string]*MockMongoCollection
}

func NewMockMongoClient(col *MockMongoCollection) *MockMongoClient {
	return &MockMongoClient{
		db: map[string]*MockMongoCollection{"urls": col},
	}
}

func (m *MockMongoClient) Database(name string, opts ...*options.DatabaseOptions) *MockMongoDatabase {
	return &MockMongoDatabase{colls: m.db}
}

type MockMongoDatabase struct {
	colls map[string]*MockMongoCollection
}

func (db *MockMongoDatabase) Collection(name string, opts ...*options.CollectionOptions) *MockMongoCollection {
	return db.colls[name]
}
