package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	log.Println("Starting up...")

	// values for Mongo and TTN
	envs := readEnvs()

	// Adding custom logger to prevent logs from being filled with /health endpoints
	r := gin.New()
	r.Use(CustomLogger())
	r.Use(gin.Recovery())

	// connect to mongo - start
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(envs.mongo_conn).SetServerAPIOptions(serverAPI)
	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	// Send a ping to confirm a successful connection
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
	// connect to mongo - end

	config := cors.DefaultConfig()

	// config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	// config.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	// config.AllowCredentials = true

	if isProduction() {
		config.AllowOrigins = []string{"https://mini-farm-tracker.io", "https://www.mini-farm-tracker.io"}
	} else {
		// vue development
		config.AllowOrigins = []string{"http://localhost:5173"}
	}

	r.Use(cors.New(config))
	// r.SetTrustedProxies([]string{"mini-farm-tracker.io"})
	// r.ForwardedByClientIP = true

	api := r.Group("/api")
	{
		api.GET("/data", dataFn)
	}

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/webhook", func(c *gin.Context) {
		handleWebhook(c, envs)
	})

	log.Printf("Endpoint: %s not logged\n", HEALTH_ENDPOINT)
	r.GET(HEALTH_ENDPOINT, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// if gin.Mode() == "release" {
	// 	log.Println("Starting in release mode")
	// 	// log.Fatal(autotls.Run(r, "api.mini-farm-tracker.io"))
	// 	log.Fatal(autotls.Run(r, "mini-farm-tracker.io"))
	// } else {
	// 	log.Println("Starting in local dev mode")
	// 	log.Fatal(r.Run())
	// }

	log.Println("Server listening...")
	// port defaults 8080 but for clarify, declaring
	log.Fatal(r.Run(":8080"))
}
