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

type DataWithTimes struct {
	SensorId  string
	StartTime time.Time
	EndTime   *time.Time
}

type QuerySensorData struct {
	SensorId string `uri:"sensor_id" binding:"required"`
}

type QueryTimeData struct {
	StartTime *time.Time `form:"start,omitempty" time_format:"2006-01-02T15:04:05Z07:00" time_utc:"1"` // RFC3339
	EndTime   *time.Time `form:"end,omitempty" time_format:"2006-01-02T15:04:05Z07:00" time_utc:"1"`   // RFC3339, optional
}

type QueryAggregationData struct {
	DataType  CalibratedDataNames `form:"dataType" binding:"required,oneof=volume airTemperature lightIntensity uVIndex windSpeed windDirection rainfallHourly barometricPressure"`
	StartTime *time.Time          `form:"start,omitempty" time_format:"2006-01-02T15:04:05Z07:00" time_utc:"1"` // RFC3339
	EndTime   *time.Time          `form:"end,omitempty" time_format:"2006-01-02T15:04:05Z07:00" time_utc:"1"`   // RFC3339, optional
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

func getAggregationData(c *gin.Context, server *Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var querySensorData QuerySensorData
	var queryTimeData QueryAggregationData
	if err := c.ShouldBindUri(&querySensorData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.ShouldBind(&queryTimeData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if queryTimeData.StartTime == nil {
		log.Println("No start time available, using the default of a 7 day period")
		v := time.Now().AddDate(0, 0, -7)
		queryTimeData.StartTime = &v
	}

	if queryTimeData.EndTime == nil {
		log.Println("No end time available, using now")
		v := time.Now()
		queryTimeData.EndTime = &v
	}

	sensor, exists := server.Sensors.Get(querySensorData.SensorId)
	if !exists {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"status": fmt.Sprintf("Unable to find sensor: %s", querySensorData.SensorId),
		})
		return
	}

	options := options.Find().SetSort(
		bson.D{
			{
				Key: "date", Value: 1,
			},
		},
	).SetProjection(
		bson.D{
			{
				Key: "_id", Value: 0,
			},
			{
				Key: "metadata.sensor", Value: 0,
			},
			{
				Key: "metadata.dataType", Value: 0,
			},
		},
	).SetLimit(DATA_API_LIMIT)

	filter := bson.D{
		{Key: "metadata.sensor", Value: sensor.Id},
		{Key: "metadata.dataType", Value: queryTimeData.DataType},
		{Key: "date",
			Value: bson.D{
				{Key: "$gte", Value: primitive.NewDateTimeFromTime(*queryTimeData.StartTime)},
				{Key: "$lt", Value: primitive.NewDateTimeFromTime(*queryTimeData.EndTime)},
			},
		},
	}

	results, err := GetAggregatedDataCollection(server.MongoDb).Find(ctx, filter, options)
	if err != nil {
		log.Printf("Error within %T data query: %v\n", results, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status": err,
		})
		return
	}

	c.Header(DATA_API_LIMIT_HEADER, fmt.Sprint(DATA_API_LIMIT))
	c.JSON(http.StatusOK, results)
	return
}

func getRawDataWithSensorId(c *gin.Context, server *Server) {
	sharedDataPullFunctionality(c, server, GetRawDataCollection)
}

func getCalibratedDataWithSensorId(c *gin.Context, server *Server) {
	sharedDataPullFunctionality(c, server, GetCalibratedDataCollection)
}

func sharedDataPullFunctionality[T QueryData](c *gin.Context, server *Server, dataPullFn func(db MongoDatabase) MongoCollection[T]) {
	var querySensorData QuerySensorData
	var queryTimeData QueryTimeData
	if err := c.ShouldBindUri(&querySensorData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.ShouldBind(&queryTimeData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if queryTimeData.StartTime == nil {
		log.Println("No start time available, using the default of a 7 day period")
		v := time.Now().AddDate(0, 0, -7)
		queryTimeData.StartTime = &v
	}

	if queryTimeData.EndTime == nil {
		log.Println("No end time available, using now")
		v := time.Now()
		queryTimeData.EndTime = &v
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sensor, exists := server.Sensors.Get(querySensorData.SensorId)
	if !exists {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"status": fmt.Sprintf("Unable to find sensor: %s", querySensorData.SensorId),
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
					{Key: "$gte", Value: primitive.NewDateTimeFromTime(*queryTimeData.StartTime)},
					{Key: "$lt", Value: primitive.NewDateTimeFromTime(*queryTimeData.EndTime)},
				}},
			},
			options,
		)
		if err != nil {
			log.Printf("Error within %T data query: %v\n", results, err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status": err,
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
