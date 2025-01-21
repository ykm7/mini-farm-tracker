package main

import (
	"context"
	"log"
	"mini-farm-tracker-server/core"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

/*
To be run manually to populate the database with various mock data
*/
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
	// Ignore duplicate key errorls
	dontPanicOnMongoCode(11000, err)
	log.Printf("%v", inserted)

	// generates raw data - WORKING
	// From successfully parsed payload
	// {
	// 	"Bat": 3.427,
	// 	"Distance": "2321 mm",
	// 	"Interrupt_flag": 0,
	// 	"Sensor_flag": 1,
	// 	"TempC_DS18B20": "0.00"
	// }
	mockSensorData := []core.LDDS45RawData{
		{
			Distance:     "2321 mm",
			Battery:      3.427,
			InterruptPin: uint8(0),
			Temperature:  "0.00",
			SensorFlag:   uint8(0),
		},
	}
	timestamp := time.Now()
	for _, v := range mockSensorData {
		if _, err = core.GetRawDataCollection[core.LDDS45RawData](mongoDb).InsertOne(context.TODO(), core.RawData[core.LDDS45RawData]{
			Timestamp: primitive.NewDateTimeFromTime(timestamp),
			Sensor:    sensorName,
			Data:      v,
		}); err != nil {
			log.Panicf("%v", err)
		}

		timestamp = timestamp.Add(-1 + 24*time.Hour)
	}

	// WORKING
	results, err := core.GetRawDataCollection[core.LDDS45RawData](mongoDb).Find(context.TODO(), bson.M{"sensor": sensorName})
	if err != nil {
		// Handle error
		panic(err)
	}

	log.Printf("Raw data: %v", results)
}
