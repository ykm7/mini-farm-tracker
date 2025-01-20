package core

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func handleWithoutSensorID(c *gin.Context, mongoDb MongoDatabase) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	results, err := GetSensorCollection(mongoDb).Find(ctx, nil)
	if err != nil {
		// Handle error
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "ok",
		})
	}

	c.JSON(http.StatusOK, results)
}

func handleWithSensorID(c *gin.Context) {
	sensorID := c.Param("SENSOR_ID")

	c.JSON(http.StatusOK, gin.H{
		"message": "Fetching data for sensor " + sensorID,
	})
}

func getRawDataWithSensorId(c *gin.Context, mongoDb MongoDatabase) {
	sensorID := c.Param("SENSOR_ID")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	results, err := GetRawDataCollection(mongoDb).Find(ctx, bson.D{{Key: "sensor", Value: sensorID}})
	if err != nil {
		// Handle error
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "ok",
		})
	}

	c.JSON(http.StatusOK, results)
}

func getCalibratedDataWithSensorId(c *gin.Context) {
	sensorID := c.Param("SENSOR_ID")
	log.Printf("%s\n", sensorID)
	c.JSON(http.StatusNotImplemented, nil)
}
