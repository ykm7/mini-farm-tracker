package main

import (
	"context"
	"fmt"
	"log"
	"mini-farm-tracker-server/core"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OldRawDataType interface {
	OldLDDS45RawData | OldRandomRawData
}

type OldRawData[T OldRawDataType] struct {
	Timestamp primitive.DateTime `bson:"timestamp"`
	Sensor    *string            `bson:"sensor,omitempty"`
	Valid     bool               `bson:"valid,omitempty"`
	Data      T                  `bson:"data"`
}

type OldRawDataFns interface {
	DetermineValid() bool
}

/*
Raw result from within the TTN payload. [mini-farm-tracker-server] [2025-01-21 07:14:44] 2025/01/21 07:14:44 'Decoded' payload: map[Bat:3.402 Distance:1752 mm Interrupt_flag:0 Sensor_flag:1 TempC_DS18B20:0.00]

Uplink formatter added to within the TTN when selecting the device from the repository.
*/
type OldLDDS45RawData struct {
	Battery      float64 `json:"Bat"`      // units are 'mv'
	Distance     string  `json:"Distance"` // units are 'mm'
	InterruptPin uint8   `json:"Interrupt_flag"`
	Temperature  string  `json:"TempC_DS18B20"` // units are 'c'
	SensorFlag   uint8   `json:"Sensor_flag"`
}

func (lDDS45RawData *OldLDDS45RawData) DetermineValid() bool {
	distanceSplit := strings.Split(lDDS45RawData.Distance, " ")
	_, units := distanceSplit[0], distanceSplit[1]

	_, ok := core.StringToUnits(units)
	return ok == nil
}

// Just to test the generics
type OldRandomRawData struct {
}

func V1OfCalibratedDataToV2(database *mongo.Database) {
	// const batchSize = 1
	calibratedCollectionName := string(core.CALIBRATED_DATA_COLLECTION)
	tempcalibratedCollectionName := fmt.Sprintf("TEMP_%s", calibratedCollectionName)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "dataPoints", Value: bson.D{{Key: "$exists", Value: false}}}}}},
		// Sort by timestamp
		{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: 1}}}},
		// Add limit. Partly so this can run efficently within a loop, however useful to test it works too.
		// bson.D{{Key: "$limit", Value: batchSize}},
		{{Key: "$addFields", Value: bson.D{
			{Key: "dataPoints", Value: bson.D{
				{Key: "volume", Value: bson.D{
					{Key: "units", Value: "$units"},
					{Key: "data", Value: "$data"},
				},
				},
			}},
		}}},
		bson.D{{Key: "$unset", Value: bson.A{
			"units",
			"data",
		}}},
		{{Key: "$out", Value: tempcalibratedCollectionName}},
	}

	cursor, err := database.Collection(calibratedCollectionName).Aggregate(context.Background(), pipeline)

	if err != nil {
		panic(err)
	}

	var results []bson.M
	if err = cursor.All(context.Background(), &results); err != nil {
		log.Fatal(err)
	}

	batchCount := len(results)
	fmt.Printf("Processed documents: %d\n", batchCount)

	// Now, read from the temp collection and insert into the time series collection
	tempCursor, err := database.Collection(tempcalibratedCollectionName).Find(context.Background(), bson.M{})
	if err != nil {
		panic(err)
	}
	defer tempCursor.Close(context.Background())

	var documents []interface{}
	for tempCursor.Next(context.Background()) {
		var doc bson.M
		if err := tempCursor.Decode(&doc); err != nil {
			panic(err)
		}
		documents = append(documents, doc)
	}

	if len(documents) > 0 {
		_, err = database.Collection(calibratedCollectionName).InsertMany(context.Background(), documents)
		if err != nil {
			panic(err)
		}
	}
}

/*
*
Due to the raw_data collection being a timeseries collection, I need to write the data to a temp collection
and them write data from this collection back into the original collection.
*/
func V1OfRawDataToV2(database *mongo.Database) {
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()

	const batchSize = 1
	rawCollectionName := string(core.RAW_DATA_COLLECTION)
	tempRawCollectionName := fmt.Sprintf("TEMP_%s", rawCollectionName)

	pipeline := mongo.Pipeline{
		// Run on all this have not been processed yet
		{{Key: "$match", Value: bson.D{{Key: "data.LDDS45", Value: bson.D{{Key: "$exists", Value: false}}}}}},
		// Sort by timestamp
		{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: 1}}}},
		// Add limit. Partly so this can run efficently within a loop, however useful to test it works too.
		// bson.D{{Key: "$limit", Value: batchSize}},
		{{Key: "$addFields", Value: bson.D{
			{Key: "data.LDDS45", Value: bson.D{
				{Key: "battery", Value: "$data.battery"},
				{Key: "distance", Value: "$data.distance"},
				{Key: "interruptPin", Value: "$data.interruptpin"},
				{Key: "temperature", Value: "$data.temperature"},
				{Key: "sensorFlag", Value: "$data.sensorflag"},
			}},
		}}},
		bson.D{{Key: "$unset", Value: bson.A{
			"data.battery",
			"data.distance",
			"data.interruptpin",
			"data.temperature",
			"data.sensorflag",
		}}},
		{{Key: "$out", Value: tempRawCollectionName}},
	}

	cursor, err := database.Collection(rawCollectionName).Aggregate(context.Background(), pipeline)

	if err != nil {
		panic(err)
	}

	var results []bson.M
	if err = cursor.All(context.Background(), &results); err != nil {
		log.Fatal(err)
	}

	batchCount := len(results)
	fmt.Printf("Processed documents: %d\n", batchCount)

	// Now, read from the temp collection and insert into the time series collection
	tempCursor, err := database.Collection(tempRawCollectionName).Find(context.Background(), bson.M{})
	if err != nil {
		panic(err)
	}
	defer tempCursor.Close(context.Background())

	var documents []interface{}
	for tempCursor.Next(context.Background()) {
		var doc bson.M
		if err := tempCursor.Decode(&doc); err != nil {
			panic(err)
		}
		documents = append(documents, doc)
	}

	if len(documents) > 0 {
		_, err = database.Collection(rawCollectionName).InsertMany(context.Background(), documents)
		if err != nil {
			panic(err)
		}
	}

	// Once confirmed that all data has been moved, delete the original entries.
	// Will perform this manually via MongoDB Compass
}

func main() {
	envs := core.ReadEnvs()

	database, mongoDeferFn := core.SetupMongo(envs)
	// mongoDb := &core.MongoDatabaseImpl{Db: database}
	defer mongoDeferFn()

	V1OfCalibratedDataToV2(database)
}
