package main

import (
	"context"
	"fmt"
	"log"
	"mini-farm-tracker-server/core"
	"time"

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

func insertSensorConfig(mongoDb core.MongoDatabase, sensorConfig *core.SensorConfiguration) (primitive.ObjectID, error) {
	result, err := core.GetSensorConfigurationCollection(mongoDb).InsertOne(context.TODO(), *sensorConfig)

	if err != nil {
		log.Printf("Error inserting sensor configuration: %v\n", err)
		log.Printf("Result is %v\n", result)
		return primitive.NilObjectID, err
	}

	if cast, ok := result.InsertedID.(primitive.ObjectID); !ok {
		return primitive.NilObjectID, fmt.Errorf("Expected ID type to be of ObjectId")
	} else {
		return cast, nil
	}
}

func insertAssert(mongoDb core.MongoDatabase, asset *core.Asset) (primitive.ObjectID, error) {
	result, err := core.GetAssetsCollection(mongoDb).InsertOne(context.TODO(), *asset)

	if err != nil {
		log.Printf("Error inserting asset: %v\n", err)
		log.Printf("Result is %v\n", result)
		return primitive.NilObjectID, err
	}

	if cast, ok := result.InsertedID.(primitive.ObjectID); !ok {
		return primitive.NilObjectID, fmt.Errorf("Expected ID type to be of ObjectId")
	} else {
		return cast, nil
	}
}

/*
To be run manually to populate the database with various mock data

We have two tanks
"first one"
172 000 litres - Rhino

	Height: 2.2 m
	Radius: 4.99 m ===  5m

181 000 litres - Rhino

	Height: 2.2 m
	Radius: 5.12 m
	Diameter: 10.24 m
*/
func main() {
	envs := core.ReadEnvs()

	database, mongoDeferFn := core.SetupMongo(envs)
	mongoDb := &core.MongoDatabaseImpl{Db: database}
	defer mongoDeferFn()

	// testing mongo - START
	// var inserted *mongo.InsertOneResult
	// var err error

	// sensorName := "Sensor 1"
	// inserted, err = core.GetSensorCollection(mongoDb).InsertOne(context.TODO(), core.Sensor{Id: sensorName})
	// // Ignore duplicate key error
	// dontPanicOnMongoCode(11000, err)
	// log.Printf("%v", inserted)

	// // inserted, err = core.GetSensorCollection(mongoDb).InsertOne(context.TODO(), core.Sensor{Id: "Sensor 2"})
	// // Ignore duplicate key errorls
	// dontPanicOnMongoCode(11000, err)
	// log.Printf("%v", inserted)

	sensorId := "a840414f118397f3"

	asset := &core.Asset{
		Name: "Mock Initial Asset",
		Sensors: &[]string{
			sensorId,
		},
		Metrics: &core.AssetMetrics{
			Volume: &core.AssetMetricsCylinderVolume{
				Volume: float64(172000),
				Radius: float64(5),
				Height: float64(2.2),
			},
		},
	}

	assetId, err := insertAssert(mongoDb, asset)
	if err != nil {
		panic(err)
	}

	sensorConfig := &core.SensorConfiguration{
		// not setting id - auto gen
		Sensor:  sensorId,
		Asset:   assetId,
		Applied: primitive.NewDateTimeFromTime(time.Now()),
		Offset: &struct {
			Distance *struct {
				Distance float64    "bson:\"distance\""
				Units    core.UNITS "bson:\"units\""
			} "bson:\"distance\""
		}{
			Distance: &struct {
				Distance float64    "bson:\"distance\""
				Units    core.UNITS "bson:\"units\""
			}{
				Distance: float64(0),
				Units:    core.METRES,
			},
		},
	}

	_, err = insertSensorConfig(mongoDb, sensorConfig)
	if err != nil {
		panic(err)
	}

	// generates raw data - WORKING
	// From successfully parsed payload
	// {
	// 	"Bat": 3.427,
	// 	"Distance": "2321 mm",
	// 	"Interrupt_flag": 0,
	// 	"Sensor_flag": 1,
	// 	"TempC_DS18B20": "0.00"
	// }
	// mockSensorData := []core.LDDS45RawData{
	// 	{
	// 		Distance:     "2321 mm",
	// 		Battery:      3.427,
	// 		InterruptPin: uint8(0),
	// 		Temperature:  "0.00",
	// 		SensorFlag:   uint8(0),
	// 	},
	// }
	// timestamp := time.Now()
	// for _, v := range mockSensorData {
	// 	if _, err = core.GetRawDataCollection[core.LDDS45RawData](mongoDb).InsertOne(context.TODO(), core.RawData[core.LDDS45RawData]{
	// 		Timestamp: primitive.NewDateTimeFromTime(timestamp),
	// 		Sensor:    &sensorName,
	// 		Data:      v,
	// 	}); err != nil {
	// 		log.Panicf("%v", err)
	// 	}

	// 	timestamp = timestamp.Add(-1 + 24*time.Hour)
	// }

	// // WORKING
	// results, err := core.GetRawDataCollection[core.LDDS45RawData](mongoDb).Find(context.TODO(), bson.M{"sensor": sensorName})
	// if err != nil {
	// 	// Handle error
	// 	panic(err)
	// }

	// log.Printf("Raw data: %v", results)
}
