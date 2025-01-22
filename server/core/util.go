package core

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type environmentVariables struct {
	ttn_webhhook_api string
	mongo_conn       string
}

func ReadEnvs() *environmentVariables {
	if !isProduction() {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	return &environmentVariables{
		ttn_webhhook_api: os.Getenv("TTN_WEBHOOK_API"),
		mongo_conn:       os.Getenv("MONGO_CONN"),
	}
}

/*
Gins mode is set to "release" if the
environment variable GIN_MODE == "release"
*/
func isProduction() bool {
	return gin.Mode() == "release"
}

func convertTimeStringToMongoTime(timeStr string) (primitive.DateTime, error) {
	t, err := time.Parse(time.RFC3339Nano, timeStr)
	if err != nil {
		// Handle error
		return 0, err
	}
	return primitive.NewDateTimeFromTime(t), nil
}

/*
TODO: Move the "Watch" function to within the wrapper functionality to be the same as .Find etc
Fair bit of TODOs here. propagate cancellation context. Examine retry... I believe it retries one automatically
*/
func ListenToSensors(ctx context.Context, mongoDb *mongo.Database, sensorCache map[string]Sensor, exitChan chan struct{}) {
	results, err := GetSensorCollection(mongoDb).Find(ctx, nil)
	for _, r := range results {
		sensorCache[r.Id] = r
	}
	if err != nil {
		panic(err)
	}

	// 'UpdateLookup' Do to include the full document on insert, update and replace.
	//
	opts := options.ChangeStream().SetFullDocument(options.UpdateLookup).SetFullDocumentBeforeChange(options.Required)
	sensorStream, err := mongoDb.Collection(string(SENSORS_COLLECTION)).Watch(ctx, mongo.Pipeline{}, opts)
	if err != nil {
		panic(err)
	}

	go func(routineCtx context.Context, stream *mongo.ChangeStream) {
		defer stream.Close(routineCtx)
		// defer waitGroup.Done()

		for stream.Next(routineCtx) {
			fmt.Println("Stream listener on the 'sensors' collection started...")

			var changeEvent bson.M
			if err := sensorStream.Decode(&changeEvent); err != nil {
				log.Printf("Error decoding change event: %v", err)
				continue
			}

			var sensor Sensor
			// Handle different operation types
			switch changeEvent["operationType"] {
			case "insert", "update", "replace":

				if fullDoc, ok := changeEvent["fullDocument"].(bson.M); ok {
					bsonData, err := bson.Marshal(fullDoc)
					if err != nil {
						log.Printf("Error marshaling full document: %v", err)
						continue
					}
					if err := bson.Unmarshal(bsonData, &sensor); err != nil {
						log.Printf("Error unmarshaling full document: %v", err)
						continue
					}
					sensorCache[sensor.Id] = sensor
					fmt.Printf("Sensor 'inserted', 'updated' or 'replaced': %+v\n", sensor)
				}

			case "delete":
				if fullDocBefore, ok := changeEvent["fullDocumentBeforeChange"].(bson.M); ok {
					bsonData, err := bson.Marshal(fullDocBefore)
					if err != nil {
						log.Printf("Error marshaling full document before change: %v", err)
						continue
					}
					if err := bson.Unmarshal(bsonData, &sensor); err != nil {
						log.Printf("Error unmarshaling full document before change: %v", err)
						continue
					}
					fmt.Printf("Sensor deleted: %+v\n", sensor)
					delete(sensorCache, sensor.Id)
				}
			}
		}

		if err := sensorStream.Err(); err != nil {
			log.Printf("Stream error: %v", err)
			exitChan <- struct{}{}
		}
	}(ctx, sensorStream)
}

/*
Alternative

import "golang.org/x/exp/maps"
values := maps.Values(myMap)
*/
func mapToList[T any](m map[string]T) []T {
	result := make([]T, 0, len(m))
	for _, value := range m {
		result = append(result, value)
	}
	return result
}
