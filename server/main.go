package main

import (
	"context"
	"log"
	"mini-farm-tracker-server/core"

	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	log.Println("Starting up...")

	// values for Mongo and TTN
	envs := core.ReadEnvs()

	mongoDb, mongoDeferFn := core.SetupMongo(envs)
	defer mongoDeferFn()

	// testing mongo - START
	var inserted *mongo.InsertOneResult
	var err error

	inserted, err = core.GetSensorCollection(mongoDb).InsertOne(context.TODO(), core.Sensor{Id: "Sensor X"})
	if err != nil {
		log.Panicf("%v", err)
	}

	log.Printf("%v", inserted)
	// testing mongo - END

	r := core.SetupRouter(envs, mongoDb)

	log.Println("Server listening...")
	// port defaults 8080 but for clarify, declaring
	log.Fatal(r.Run(":8080"))
}
