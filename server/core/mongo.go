package core

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TODO: Leaving this as the default value until I understand better how to use this
const MAX_POOL_SIZE = 100

/*
TODO: Add cancellation context
*/
func SetupMongo(envs *environmentVariables) (db *mongo.Database, deferFn func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(envs.Mongo_conn).SetServerAPIOptions(serverAPI).SetMaxPoolSize(MAX_POOL_SIZE)
	// Create a new client and connect to the server
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		panic(err)
	}

	// Send a ping to confirm a successful connection
	if err := client.Database("admin").RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		panic(err)
	}
	log.Println("Pinged your deployment. You successfully connected to MongoDB!")

	db = client.Database(DATABASE_NAME)

	//// Leaving this here for documentation. These is required to be set BUT the established user does not have permission to modify.
	// result := db.RunCommand(ctx, bson.D{
	// 	{Key: "collMod", Value: SENSORS_COLLECTION},
	// 	{Key: "changeStreamPreAndPostImages", Value: bson.D{{Key: "enabled", Value: true}}},
	// })

	// var raw bson.Raw
	// err = result.Decode(&raw)
	// if err != nil {
	// 	log.Panicf("Failed to enable changeStreamPreAndPostImages: %v", err)
	// }

	deferFn = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}

	return
}

//go:generate mockgen -destination=../mocks/mock_MongoDatabase.go -package=mocks MongoDatabase
type MongoDatabase interface {
	Collection(name string, opts ...*options.CollectionOptions) MongoCollection[any]
}

type MongoCollection[T any] interface {
	InsertOne(ctx context.Context, document T) (*mongo.InsertOneResult, error)
	FindOne(ctx context.Context, filter interface{}, result *T) error
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]T, error)
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	Watch(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error)
	DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
	Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) ([]T, error)
}

type MongoDatabaseImpl struct {
	Db *mongo.Database
}

// Wrapper struct implementing MongoCollection
type MongoCollectionWrapper[T any] struct {
	col *mongo.Collection
}

func (m *MongoDatabaseImpl) Collection(name string, opts ...*options.CollectionOptions) MongoCollection[any] {
	return &MongoCollectionWrapper[any]{col: m.Db.Collection(name, opts...)}
}

/*
Not ideal as mocking logic is now within the core code path. TODO: revisit - however functional and allows tests
*/
func getTypedCollection[T any](mongoDb MongoDatabase, collectionName string) MongoCollection[T] {
	anyCollection := mongoDb.Collection(collectionName)

	switch c := anyCollection.(type) {
	case MongoCollection[T]:
		return c
	case *MongoCollectionWrapper[any]:
		return &MongoCollectionWrapper[T]{col: c.col}
	// case MongoCollection[any]:
	// 	return &MockMongoCollectionWrapper[T]{col: c}
	default:
		panic(fmt.Sprintf("Unexpected collection type: %T", anyCollection))
	}
}

func GetSensorCollection(mongoDb MongoDatabase) MongoCollection[Sensor] {
	return getTypedCollection[Sensor](mongoDb, string(SENSORS_COLLECTION))
}

func GetRawDataCollection(mongoDb MongoDatabase) MongoCollection[RawData] {
	return getTypedCollection[RawData](mongoDb, string(RAW_DATA_COLLECTION))
}

func GetSensorConfigurationCollection(mongoDb MongoDatabase) MongoCollection[SensorConfiguration] {
	return getTypedCollection[SensorConfiguration](mongoDb, string(SENSOR_CONFIGURATIONS_COLLECTION))
}

func GetCalibratedDataCollection(mongoDb MongoDatabase) MongoCollection[CalibratedData] {
	return getTypedCollection[CalibratedData](mongoDb, string(CALIBRATED_DATA_COLLECTION))
}

func GetAssetsCollection(mongoDb MongoDatabase) MongoCollection[Asset] {
	return getTypedCollection[Asset](mongoDb, string(ASSETS_COLLECTION))
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

func (m *MongoCollectionWrapper[T]) Watch(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error) {
	return m.col.Watch(ctx, pipeline, opts...)
}

func (m *MongoCollectionWrapper[T]) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return m.col.DeleteMany(ctx, filter, opts...)
}

func (m *MongoCollectionWrapper[T]) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) ([]T, error) {
	cursor, err := m.col.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []T
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

/*
*
dataType - rainfallHourly
*/
func getSumAggregation(dataType string) mongo.Pipeline {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: fmt.Sprintf("dataPoints.%s", dataType), Value: bson.D{{Key: "$exists", Value: true}}},
		}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "date", Value: bson.D{{Key: "$dateToString", Value: bson.D{{Key: "format", Value: "%Y-%m-%d"}, {Key: "date", Value: "$timestamp"}}}}},
				{Key: "sensor", Value: "$sensor"},
			}},
			{Key: "totalRainfall", Value: bson.D{{Key: "$sum", Value: fmt.Sprintf("dataPoints.%s.data", dataType)}}},
			{Key: "unit", Value: bson.D{{Key: "$first", Value: fmt.Sprintf("dataPoints.%s.units", dataType)}}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "date", Value: bson.D{{Key: "$dateFromString", Value: bson.D{{Key: "dateString", Value: "$_id.date"}}}}},
			{Key: "metadata", Value: bson.D{
				{Key: "sensor", Value: "$_id.sensor"},
				{Key: "type", Value: "daily"},
			}},
			{Key: "totalRainfall", Value: bson.D{
				{Key: "value", Value: "$totalRainfall"},
				{Key: "unit", Value: "$unit"},
			}},
		}}},
		{{Key: "$merge", Value: bson.D{
			{Key: "into", Value: "daily_aggregates"},
			{Key: "on", Value: bson.A{"date", "metadata.sensor"}},
			{Key: "whenMatched", Value: "replace"},
			{Key: "whenNotMatched", Value: "insert"},
		}}},
	}

	return pipeline
}

func createAggregationPipeline(dataType string, aggregationType string, groupByFormat string) mongo.Pipeline {
	return mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: fmt.Sprintf("dataPoints.%s", dataType), Value: bson.D{{Key: "$exists", Value: true}}},
		}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "date", Value: bson.D{{Key: "$dateToString", Value: bson.D{{Key: "format", Value: groupByFormat}, {Key: "date", Value: "$timestamp"}}}}},
				{Key: "sensor", Value: "$sensor"},
			}},
			{Key: "totalValue", Value: bson.D{{Key: "$sum", Value: fmt.Sprintf("$dataPoints.%s.data", dataType)}}},
			{Key: "unit", Value: bson.D{{Key: "$first", Value: fmt.Sprintf("$dataPoints.%s.units", dataType)}}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "date", Value: bson.D{{Key: "$dateFromString", Value: bson.D{{Key: "dateString", Value: "$_id.date"}}}}},
			{Key: "metadata", Value: bson.D{
				{Key: "sensor", Value: "$_id.sensor"},
				{Key: "type", Value: aggregationType},
				{Key: "dataType", Value: dataType},
			}},
			{Key: "totalValue", Value: bson.D{
				{Key: "value", Value: "$totalValue"},
				{Key: "unit", Value: "$unit"},
			}},
		}}},
		{{Key: "$merge", Value: bson.D{
			{Key: "into", Value: "aggregated_data"},
			{Key: "on", Value: bson.A{"date", "metadata.sensor", "metadata.type", "metadata.dataType"}},
			{Key: "whenMatched", Value: "replace"},
			{Key: "whenNotMatched", Value: "insert"},
		}}},
	}
}
