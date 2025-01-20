package core

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const HEALTH_ENDPOINT = "/health"
const SENSOR_ID_PARAM = "SENSOR_ID"

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
		sensorApi := api.Group("/sensors")
		{
			sensorApi.GET("", func(c *gin.Context) {
				handleWithoutSensorID(c, db)
			})
			sensorApi.GET(fmt.Sprintf(":%s", SENSOR_ID_PARAM), handleWithSensorID)

			sensorDataApi := sensorApi.Group(fmt.Sprintf(":%s/data", SENSOR_ID_PARAM))
			{
				sensorDataApi.GET("/raw_data", func(c *gin.Context) {
					getRawDataWithSensorId(c, db)
				})
				sensorDataApi.GET("/calibrated_data", getCalibratedDataWithSensorId)
			}
		}
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
