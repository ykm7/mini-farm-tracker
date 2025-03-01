package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"mini-farm-tracker-server/core"

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

func pullAllRainfallAggregations(mongodb core.MongoDatabase) {

	loc, err := time.LoadLocation("Australia/Perth")
	if err != nil {
		log.Fatalf("Could not load timezone: %v", err)
	}

	// The purpose here to to grab and aggregate ALL the missing data
	// easily covers EPOCH of "project"
	timeRange := time.Now().In(loc).AddDate(-2, 0, 0)

	source := core.GetCalibratedDataCollection(mongodb)
	target := core.GetAggregatedDataCollection(mongodb)
	metricType := core.RAIN_ACCUMULATION_DATA_NAMES

	tasks := make(chan core.TaskJob)

	goroutineCount := runtime.NumCPU() * 4

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		// Refactor this to remove the timeout and to be tied to all the tasks queued
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		core.Debounce(ctx, time.Second*1, 100, tasks, core.TaskHandler, goroutineCount)

		log.Println("Debounce tasks completed")
	}()

	// daily
	if task, err := core.GenerateAggregationTask(source, target, metricType, core.DAILY_TYPE, timeRange, nil); err != nil {
		log.Fatalf("Error while generating daily rainfall accumulation %v\n", err)
	} else {
		tasks <- &task
	}

	// weekly
	if task, err := core.GenerateAggregationTask(source, target, metricType, core.WEEKLY_TYPE, timeRange, nil); err != nil {
		log.Fatalf("Error while generating daily rainfall accumulation %v\n", err)
	} else {
		tasks <- &task
	}

	// monthly
	if task, err := core.GenerateAggregationTask(source, target, metricType, core.MONTHLY_TYPE, timeRange, nil); err != nil {
		log.Fatalf("Error while generating daily rainfall accumulation %v\n", err)
	} else {
		tasks <- &task
	}

	// yearly
	if task, err := core.GenerateAggregationTask(source, target, metricType, core.YEARLY_TYPE, timeRange, nil); err != nil {
		log.Fatalf("Error while generating daily rainfall accumulation %v\n", err)
	} else {
		tasks <- &task
	}

	wg.Wait()
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

	// redis, redisDeferFn := core.GetRedisClient(envs)
	// defer redisDeferFn()

	pullAllRainfallAggregations(mongoDb)
}
