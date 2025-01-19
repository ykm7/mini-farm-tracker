package main

import (
	"context"
	"log"
	"mini-farm-tracker-server/core"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func dontPanicOnMongoCode(code int, err error) {
	if err != nil {
		ignore := false
		if writeException, ok := err.(mongo.WriteException); ok {
			for _, writeError := range writeException.WriteErrors {
				if writeError.Code == code {
					ignore = true
					break
				}
			}
		}
		if !ignore {
			log.Panicf("%v", err)
		}
	}
}

func main() {

	envs := core.ReadEnvs()

	mongoDb, mongoDeferFn := core.SetupMongo(envs)
	defer mongoDeferFn()

	// testing mongo - START
	var inserted *mongo.InsertOneResult
	var err error

	sensorName := "Sensor 1"
	inserted, err = core.GetSensorCollection(mongoDb).InsertOne(context.TODO(), core.Sensor{Id: sensorName})
	// Ignore duplicate key error
	dontPanicOnMongoCode(11000, err)
	log.Printf("%v", inserted)

	// inserted, err = core.GetSensorCollection(mongoDb).InsertOne(context.TODO(), core.Sensor{Id: "Sensor 2"})
	// // Ignore duplicate key error
	// dontPanicOnMongoCode(11000, err)
	// log.Printf("%v", inserted)

	// // generates raw data - WORKING
	// mockSensorData := []int64{35, 20, 15, 115, 80, 25}
	// timestamp := time.Now()
	// for _, v := range mockSensorData {
	// 	if _, err = core.GetRawDataCollection(mongoDb).InsertOne(context.TODO(), core.RawData{
	// 		Timestamp: primitive.NewDateTimeFromTime(timestamp),
	// 		Sensor:    sensorName,
	// 		Data:      v,
	// 	}); err != nil {
	// 		log.Panicf("%v", err)
	// 	}

	// 	timestamp = timestamp.Add(-1 + 24*time.Hour)
	// }

	// WORKING
	results, err := core.GetRawDataCollection(mongoDb).Find(context.TODO(), bson.M{"sensor": sensorName})
	if err != nil {
		// Handle error
	}
	// defer cursor.Close(context.TODO())

	// var results []core.RawData
	// if err = cursor.All(context.TODO(), &results); err != nil {
	// 	// Handle error
	// 	log.Panicf("%v", err)
	// }

	log.Printf("Raw data: %v", results)
}
