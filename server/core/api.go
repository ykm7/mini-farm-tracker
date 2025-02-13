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

const DATA_API_LIMIT_HEADER = "X-Max-Data-Limit"
const DATA_API_LIMIT = 500

func getStartStopTimes(c *gin.Context) (start time.Time, end time.Time, err error) {
	now := time.Now()
	// default to 7 days
	startDate := c.DefaultQuery(START_DATE, now.AddDate(0, 0, -7).Format(time.RFC3339))
	// should be "now"
	endDate := c.DefaultQuery(END_DATE, now.Format(time.RFC3339))

	// var err error
	start, err = time.Parse(time.RFC3339, startDate)
	if err != nil {
		err = fmt.Errorf("Invalid start time format")
		return
	}

	end, err = time.Parse(time.RFC3339, endDate)
	if err != nil {
		err = fmt.Errorf("Invalid end time format")
		return
	}

	return
}

func handleWithoutSensorID(c *gin.Context, server *Server) {
	c.JSON(http.StatusOK, server.Sensors.ToList())
}

func handleAssetsWithoutId(c *gin.Context, server *Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	results, err := GetAssetsCollection(server.MongoDb).Find(ctx, nil)
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
	c.AbortWithStatusJSON(http.StatusNotImplemented, nil)
}

func getRawDataWithSensorId(c *gin.Context, server *Server) {
	sharedDataPullFunctionality(c, server, GetRawDataCollection)
}

func getCalibratedDataWithSensorId(c *gin.Context, server *Server) {
	sharedDataPullFunctionality(c, server, GetCalibratedDataCollection)
}

func sharedDataPullFunctionality[T QueryData](c *gin.Context, server *Server, dataPullFn func(db MongoDatabase) MongoCollection[T]) {
	sensorID := c.Param(SENSOR_ID_PARAM)

	start, end, err := getStartStopTimes(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sensor, exists := server.Sensors.Get(sensorID)
	if !exists {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"status": fmt.Sprintf("Unable to find sensor: %s", sensorID),
		})
		return
	}

	switch sensor.Model {
	case LDDS45, S2120:

		options := options.Find().SetSort(
			bson.D{
				{
					Key: "timestamp", Value: 1,
				},
			},
		).SetProjection(
			bson.D{
				{
					Key: "Id", Value: 0,
				},
				{
					Key: "Sensor", Value: 0,
				},
			},
		).SetLimit(DATA_API_LIMIT)

		results, err := dataPullFn(server.MongoDb).Find(
			ctx,
			bson.D{
				{Key: "sensor", Value: sensor.Id},
				{Key: "timestamp", Value: bson.D{
					{Key: "$gte", Value: primitive.NewDateTimeFromTime(start)},
					{Key: "$lt", Value: primitive.NewDateTimeFromTime(end)},
				}},
			},
			options,
		)
		if err != nil {
			log.Printf("Error within %T data query: %v\n", results, err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status": "ok",
			})
			return
		}

		c.Header(DATA_API_LIMIT_HEADER, fmt.Sprint(DATA_API_LIMIT))
		c.JSON(http.StatusOK, results)
		return

	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status": fmt.Sprintf("Unexpected sensor model: %s for sensor id: %s\n", sensor.Model, sensor.Id),
		})
		return
	}
}
