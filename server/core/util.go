package core

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Server struct {
	Envs        *environmentVariables
	MongoDb     MongoDatabase
	Sensors     SyncCache[string, Sensor]
	ExitContext context.Context
	ExitChan    chan struct{}
}

func NewSyncCache[K comparable, V any]() *syncCacheImpl[K, V] {
	return &syncCacheImpl[K, V]{
		cache: make(map[K]V),
	}
}

type SyncCache[K comparable, V any] interface {
	Get(key K) (V, bool)
	ToList() []V
	Update(key K, v V)
	Delete(key K)
}

type syncCacheImpl[K comparable, V any] struct {
	cache map[K]V
	mu    sync.RWMutex
}

func (s *syncCacheImpl[K, V]) Get(key K) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.cache[key]
	return v, ok
}

func (s *syncCacheImpl[K, V]) ToList() []V {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return mapToList(s.cache)
}

func (s *syncCacheImpl[K, V]) Update(key K, v V) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cache == nil {
		s.cache = make(map[K]V)
	}
	s.cache[key] = v
}

func (s *syncCacheImpl[K, V]) Delete(key K) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cache == nil {
		delete(s.cache, key)
	}
}

type environmentVariables struct {
	Ttn_webhhook_api string
	Mongo_conn       string
	Open_weather_api string
}

func ContextWithQuitChannel(parent context.Context, quit <-chan struct{}) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	go func() {
		select {
		case <-quit:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, cancel
}

func ReadEnvs() *environmentVariables {
	if !isProduction() {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	return &environmentVariables{
		Ttn_webhhook_api: os.Getenv("TTN_WEBHOOK_API"),
		Mongo_conn:       os.Getenv("MONGO_CONN"),
		Open_weather_api: os.Getenv("OPEN_WEATHER_API"),
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
Fair bit of TODOs here. propagate cancellation context. Examine retry... I believe it retries once automatically
*/
func ListenToSensors(server *Server) {
	results, err := GetSensorCollection(server.MongoDb).Find(server.ExitContext, nil)
	for _, r := range results {
		server.Sensors.Update(r.Id, r)
	}
	if err != nil {
		panic(err)
	}

	// 'UpdateLookup' Do to include the full document on insert, update and replace.
	//
	opts := options.ChangeStream().SetFullDocument(options.UpdateLookup).SetFullDocumentBeforeChange(options.Required)
	sensorStream, err := server.MongoDb.Collection(string(SENSORS_COLLECTION)).Watch(server.ExitContext, mongo.Pipeline{}, opts)
	if err != nil {
		panic(err)
	}

	go func(routineCtx context.Context, stream *mongo.ChangeStream) {
		defer stream.Close(routineCtx)

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
					server.Sensors.Update(sensor.Id, sensor)
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
					server.Sensors.Delete(sensor.Id)
				}
			}
		}

		if err := sensorStream.Err(); err != nil {
			log.Printf("Stream error: %v", err)
			server.ExitChan <- struct{}{}
		}
	}(server.ExitContext, sensorStream)
}

/*
Alternative

import "golang.org/x/exp/maps"
values := maps.Values(myMap)
*/
func mapToList[K comparable, V any](m map[K]V) []V {
	result := make([]V, 0, len(m))
	for _, value := range m {
		result = append(result, value)
	}
	return result
}
