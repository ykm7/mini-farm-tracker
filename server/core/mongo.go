package core

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
TODO: Add cancellation context
*/
func SetupMongo(envs *environmentVariables) (db *mongo.Database, deferFn func()) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(envs.mongo_conn).SetServerAPIOptions(serverAPI)
	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	// Send a ping to confirm a successful connection
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	db = client.Database(DATABASE_NAME)
	deferFn = func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}

	return
}

// MongoDatabase interface remains non-generic
type MongoDatabase interface {
	Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection
}

// Generic interface for collection operations
type MongoCollection[T any] interface {
	InsertOne(ctx context.Context, document T) (*mongo.InsertOneResult, error)
	FindOne(ctx context.Context, filter interface{}, result *T) error
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]T, error)
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
}

// Wrapper struct implementing MongoCollection
type MongoCollectionWrapper[T any] struct {
	col *mongo.Collection
}

func newMongoCollection[T any](col *mongo.Collection) MongoCollection[T] {
	return &MongoCollectionWrapper[T]{col: col}
}

func GetSensorCollection(mongoDb MongoDatabase) MongoCollection[Sensor] {
	return newMongoCollection[Sensor](mongoDb.Collection(string(SENSORS_COLLECTION)))
}

func GetRawDataCollection[T RawDataType](mongoDb MongoDatabase) MongoCollection[RawData[T]] {
	return newMongoCollection[RawData[T]](mongoDb.Collection(string(RAW_DATA_COLLECTION)))
}

func (m *MongoCollectionWrapper[T]) InsertOne(ctx context.Context, document T) (*mongo.InsertOneResult, error) {
	return m.col.InsertOne(ctx, document)
}

func (m *MongoCollectionWrapper[T]) FindOne(ctx context.Context, filter interface{}, result *T) error {
	return m.col.FindOne(ctx, filter).Decode(result)
}

func (m *MongoCollectionWrapper[T]) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]T, error) {
	if filter == nil {
		filter = bson.D{}
	}
	cursor, err := m.col.Find(ctx, filter, opts...)

	if err != nil {
		// Handle error
		return nil, err
	}
	defer cursor.Close(ctx)
	var results []T
	if err = cursor.All(ctx, &results); err != nil {
		// Handle error
		return nil, err
	}

	if results == nil {
		results = make([]T, 0)
	}

	return results, nil
}

func (m *MongoCollectionWrapper[T]) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return m.col.UpdateOne(ctx, filter, update, opts...)
}
