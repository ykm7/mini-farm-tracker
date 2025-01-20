package core

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

func dataFn(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
