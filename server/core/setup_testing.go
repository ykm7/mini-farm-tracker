package core

import (
	"context"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const TEST_DATABASE_NAME string = "test_db"

/*
https://golang.testcontainers.org/modules/mongodb/#connectionstring
*/
func MockSetupMongo(ctx context.Context) (db *mongo.Database, deferFn func()) {
	mongoDBContainer, err := mongodb.Run(ctx, "mongo:8")

	deferFn = func() {
		if err := testcontainers.TerminateContainer(mongoDBContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}

	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	endpoint, err := mongoDBContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(endpoint))
	if err != nil {
		log.Printf("failed to connect to MongoDB: %s", err)
		return
	}

	db = mongoClient.Database(TEST_DATABASE_NAME)

	return
}
