package core

import (
	"context"
	"fmt"
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
		return
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

	// TODO: Revisit this. Having the query the database again simply to pull the sensor model isn't ideal.
	// Want to investigate using the mongo listener to have a cached version of all the available sensors.
	sensor := &Sensor{}
	var err error
	err = GetSensorCollection(mongoDb).FindOne(ctx, bson.D{{Key: "_id", Value: sensorID}}, sensor)
	if err != nil {
		// Handle error
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "ok",
		})
		return
	}

	// var results []interface{}
	switch sensor.Model {
	case LDDS45:
		results, err := GetRawDataCollection[LDDS45RawData](mongoDb).Find(ctx, bson.D{{Key: "sensor", Value: sensor.Id}})
		if err != nil {
			log.Printf("Error within raw data query: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "ok",
			})
			return
		}

		c.JSON(http.StatusOK, results)
		return

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": fmt.Sprintf("Unexpected sensor model: %s for sensor id: %s\n", sensor.Model, sensor.Id),
		})
		return
	}
}

func getCalibratedDataWithSensorId(c *gin.Context) {
	sensorID := c.Param("SENSOR_ID")
	log.Printf("%s\n", sensorID)
	c.JSON(http.StatusNotImplemented, nil)
}
