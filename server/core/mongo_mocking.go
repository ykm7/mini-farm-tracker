package core

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MockMongoDatabase struct {
	CollectionFn func(name string, opts ...*options.CollectionOptions) MongoCollection[any]
}

func (m *MockMongoDatabase) Collection(name string, opts ...*options.CollectionOptions) MongoCollection[any] {
	return m.CollectionFn(name, opts...)
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
