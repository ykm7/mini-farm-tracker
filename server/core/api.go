package core

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func handleWithoutSensorID(c *gin.Context, sensorCache map[string]Sensor) {
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	// results, err := GetSensorCollection(mongoDb).Find(ctx, nil)
	// if err != nil {
	// 	// Handle error
	// 	c.JSON(http.StatusInternalServerError, gin.H{
	// 		"status": "ok",
	// 	})
	// 	return
	// }

	c.JSON(http.StatusOK, mapToList(sensorCache))
}

func handleAssetsWithoutId(c *gin.Context, mongoDb MongoDatabase) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	results, err := GetAssetsCollection(mongoDb).Find(ctx, nil)
	if err != nil {
		// Handle error
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status": "ok",
		})
		return
	}

	c.JSON(http.StatusOK, results)
}

func handleWithSensorID(c *gin.Context) {
	// sensorID := c.Param("SENSOR_ID")

	c.AbortWithStatusJSON(http.StatusNotImplemented, nil)
}

/*
*
Last day:

	{
	  sensor: "a840418241843281",
	  timestamp: {
	    $gte: new Date(
	      new Date().setDate(new Date().getDate() - 1)
	    ),
	    $lt: new Date()
	  }
	}

Last hour:

		{
		  sensor: "a840418241843281",
		  timestamp: {
		    $gte: new Date(
		      new Date().setHours(
		        new Date().getHours() - 1
		      )
		    ),
		    $lt: new Date()
		  }
		}

		timestamp: {
	    $gte: ISODate("2025-01-28T23:57:58.763Z"),
	    $lt: ISODate("2025-01-29T00:57:58.763Z")
	}
*/
func getRawDataWithSensorId(c *gin.Context, mongoDb MongoDatabase, sensorCache map[string]Sensor) {
	sensorID := c.Param(SENSOR_ID_PARAM)

	now := time.Now()
	// default to 7 days
	startDate := c.DefaultQuery(START_DATE, now.AddDate(0, 0, -7).Format(time.RFC3339))
	// should be "now"
	endDate := c.DefaultQuery(END_DATE, now.Format(time.RFC3339))

	start, err := time.Parse(time.RFC3339, startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start time format"})
		return
	}

	end, err := time.Parse(time.RFC3339, endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end time format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sensor, exists := sensorCache[sensorID]
	if !exists {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"status": fmt.Sprintf("Unable to find sensor: %s", sensorID),
		})
		return
	}

	switch sensor.Model {
	case LDDS45:
		results, err := GetRawDataCollection[LDDS45RawData](mongoDb).Find(
			ctx,
			bson.D{
				{Key: "sensor", Value: sensor.Id},
				{Key: "timestamp", Value: bson.D{
					{Key: "$gte", Value: primitive.NewDateTimeFromTime(start)},
					{Key: "$lt", Value: primitive.NewDateTimeFromTime(end)},
				}},
			},
			options.Find().SetProjection(
				bson.D{
					{
						Key: "Id", Value: 0,
					},
					{
						Key: "Sensor", Value: 0,
					},
				},
			),
		)
		if err != nil {
			log.Printf("Error within raw data query: %v\n", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status": "ok",
			})
			return
		}

		c.JSON(http.StatusOK, results)
		return

	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status": fmt.Sprintf("Unexpected sensor model: %s for sensor id: %s\n", sensor.Model, sensor.Id),
		})
		return
	}
}

func getCalibratedDataWithSensorId(c *gin.Context, mongoDb MongoDatabase, sensorCache map[string]Sensor) {
	sensorID := c.Param(SENSOR_ID_PARAM)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sensor, exists := sensorCache[sensorID]
	if !exists {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"status": fmt.Sprintf("Unable to find sensor: %s", sensorID),
		})
		return
	}

	switch sensor.Model {
	case LDDS45:
		results, err := GetCalibratedDataCollection(mongoDb).Find(
			ctx,
			bson.D{{Key: "sensor", Value: sensor.Id}},
			options.Find().SetProjection(
				bson.D{
					{
						Key: "Id", Value: 0,
					},
					{
						Key: "Sensor", Value: 0,
					},
				},
			),
		)
		if err != nil {
			log.Printf("Error within raw data query: %v\n", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status": "ok",
			})
			return
		}

		c.JSON(http.StatusOK, results)
		return

	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status": fmt.Sprintf("Unexpected sensor model: %s for sensor id: %s\n", sensor.Model, sensor.Id),
		})
		return
	}
}
