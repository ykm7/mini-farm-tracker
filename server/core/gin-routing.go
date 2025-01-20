package core

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const HEALTH_ENDPOINT = "/health"

func CustomLogger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{HEALTH_ENDPOINT},
	})
}

func SetupRouter(envs *environmentVariables, db MongoDatabase) *gin.Engine {
	r := gin.New()
	r.Use(CustomLogger())
	r.Use(gin.Recovery())

	config := cors.DefaultConfig()

	if isProduction() {
		config.AllowOrigins = []string{"https://mini-farm-tracker.io", "https://www.mini-farm-tracker.io"}
	} else {
		// vue development
		config.AllowOrigins = []string{"http://localhost:5173"}
	}

	r.Use(cors.New(config))
	api := r.Group("/api")
	{
		api.GET("/sensors", func(c *gin.Context) {
			handleWithoutSensorID(c, db)
		})
		api.GET("/sensors/:SENSOR_ID", handleWithSensorID)
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

	return r
}
