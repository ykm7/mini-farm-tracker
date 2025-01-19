package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	log.Println("Starting up...")

	// values for Mongo and TTN
	envs := readEnvs()

	mongoDb, mongoDeferFn := setupMongo(envs)
	defer mongoDeferFn()

	// testing mongo - START
	var inserted *mongo.InsertOneResult
	var err error

	inserted, err = GetSensorCollection(mongoDb).InsertOne(context.TODO(), Sensor{Id: primitive.NewObjectIDFromTimestamp(time.Now())})
	if err != nil {
		log.Panicf("%v", err)
	}

	log.Printf("%v", inserted)
	// testing mongo - END

	r := setupRouter(envs, mongoDb)

	log.Println("Server listening...")
	// port defaults 8080 but for clarify, declaring
	log.Fatal(r.Run(":8080"))
}
