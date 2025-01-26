package core

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MockMongoDatabase struct {
	collections map[string]MongoCollection[any]
}

func NewMockMongoDatabase() *MockMongoDatabase {
	return &MockMongoDatabase{
		collections: make(map[string]MongoCollection[any]),
	}
}

func (m *MockMongoDatabase) SetCollection(name string, collection MongoCollection[any]) {
	m.collections[name] = collection
}

func (m *MockMongoDatabase) Collection(name string, opts ...*options.CollectionOptions) MongoCollection[any] {
	return m.collections[name]
}

type MockMongoCollectionWrapper[T any] struct {
	col MongoCollection[any]
}

func (m *MockMongoCollectionWrapper[T]) InsertOne(ctx context.Context, document T) (*mongo.InsertOneResult, error) {
	return m.col.InsertOne(ctx, document)
}

func (m *MockMongoCollectionWrapper[T]) FindOne(ctx context.Context, filter interface{}, result *T) error {
	var anyResult any
	err := m.col.FindOne(ctx, filter, &anyResult)
	if err == nil {
		*result = anyResult.(T)
	}
	return err
}

func (m *MockMongoCollectionWrapper[T]) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]T, error) {
	anyResults, err := m.col.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	results := make([]T, len(anyResults))
	for i, anyResult := range anyResults {
		results[i] = anyResult.(T)
	}
	return results, nil
}

func (m *MockMongoCollectionWrapper[T]) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return m.col.UpdateOne(ctx, filter, update, opts...)
}

func (m *MockMongoCollectionWrapper[T]) Watch(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error) {
	return m.col.Watch(ctx, pipeline, opts...)
}

type MockMongoCollection[T any] struct {
	InsertOneFn func(ctx context.Context, document T) (*mongo.InsertOneResult, error)
	FindOneFn   func(ctx context.Context, filter interface{}, result *T) error
	FindFn      func(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]T, error)
	UpdateOneFn func(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	WatchFn     func(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error)
}

func (m *MockMongoCollection[T]) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]T, error) {
	return m.FindFn(ctx, filter, opts...)
}

func (m *MockMongoCollection[T]) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return m.UpdateOneFn(ctx, filter, update, opts...)
}

func (m *MockMongoCollection[T]) InsertOne(ctx context.Context, document T) (*mongo.InsertOneResult, error) {
	return m.InsertOneFn(ctx, document)
}

func (m *MockMongoCollection[T]) FindOne(ctx context.Context, filter interface{}, result *T) error {
	return m.FindOneFn(ctx, filter, result)
}

func (m *MockMongoCollection[T]) Watch(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error) {
	return m.WatchFn(ctx, pipeline, opts...)
}
