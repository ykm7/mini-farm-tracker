package main

import (
	"context"
	"log"
	"mini-farm-tracker-server/core"

	"go.mongodb.org/mongo-driver/mongo"
)

func main() {

	envs := core.ReadEnvs()

	mongoDb, mongoDeferFn := core.SetupMongo(envs)
	defer mongoDeferFn()

	// testing mongo - START
	var inserted *mongo.InsertOneResult
	var err error

	inserted, err = core.GetSensorCollection(mongoDb).InsertOne(context.TODO(), core.Sensor{Id: "Sensor 1"})
	if err != nil {
		log.Panicf("%v", err)
	}
	log.Printf("%v", inserted)

	inserted, err = core.GetSensorCollection(mongoDb).InsertOne(context.TODO(), core.Sensor{Id: "Sensor 2"})
	if err != nil {
		log.Panicf("%v", err)
	}

	log.Printf("%v", inserted)
}
