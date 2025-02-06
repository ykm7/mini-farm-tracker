package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func handleWebhook(c *gin.Context, server *Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	apiKey := c.GetHeader("X-Downlink-Apikey")
	if apiKey == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing X-Downlink-Apikey header"})
		return
	}

	// Verify API Sign
	if apiKey != server.Envs.Ttn_webhhook_api {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Webhook env is invalid"})
		return
	}

	var uplinkMessage UplinkMessage
	if err := c.ShouldBindJSON(&uplinkMessage); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sensor, exists := server.Sensors.Get(*uplinkMessage.EndDeviceIDs.DeviceID)
	if !exists {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"status": fmt.Sprintf("A gateway with the TTN deviceId of %s was not found", *uplinkMessage.EndDeviceIDs.DeviceID),
		})
		return
	}

	jsonData, err := json.Marshal(uplinkMessage.UplinkMessage.DecodedPayload)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status": fmt.Sprintf("Error parsing the decoded payload: %s", *uplinkMessage.EndDeviceIDs.DeviceID),
		})
		return
	}

	receivedAtTime, err := convertTimeStringToMongoTime(*uplinkMessage.ReceivedAt)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"status": fmt.Sprintf("Unable to parse timestamp: %s", *uplinkMessage.ReceivedAt),
		})
		return
	}

	// TODO: Store data point within Mongo
	switch sensor.Model {
	case LDDS45:
		// "LDDS45" is aware it is an volume related Sensor type.
		data := SensorData{
			LDDS45: &LDDS45RawData{},
		}
		err = json.Unmarshal(jsonData, &data.LDDS45)
		if err != nil || data.LDDS45 == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status": fmt.Sprintf("Error casting the decoded json: %v to expected data type for: %s", jsonData, LDDS45),
			})
			return
		}

		var valid bool
		if valid, err = data.DetermineValid(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status": fmt.Sprintf("On webhook for %s\n", err),
			})
			return
		}

		dataPayload := RawData{
			Timestamp: receivedAtTime,
			Sensor:    &sensor.Id,
			Data:      data,
			Valid:     valid,
		}
		_, err := GetRawDataCollection(server.MongoDb).InsertOne(ctx, dataPayload)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status": fmt.Sprintf("Error trying to insert raw data %s\n", err),
			})
			return
		}

		if valid {
			storeLDDS45CalibratedData(ctx, server.MongoDb, sensor.Id, data.LDDS45, receivedAtTime)
		}

	case S2120:
		// TODO:

		data := SensorData{
			S2120: &S2120RawData{},
		}

		err = data.S2120.Unmarshal(jsonData)

		if err != nil {
			fmt.Printf("For payload (as string) %s and expected sensor %s have error %v", string(jsonData), S2120, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status": fmt.Sprintf("Error casting the decoded json: %v (as string: %s) to expected data type for: %s", jsonData, string(jsonData), S2120),
			})
			return
		}

		// TODO: handle validity
		valid := true
		dataPayload := RawData{
			Timestamp: receivedAtTime,
			Sensor:    &sensor.Id,
			Data:      data,
			Valid:     valid,
		}
		_, err := GetRawDataCollection(server.MongoDb).InsertOne(ctx, dataPayload)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status": fmt.Sprintf("Error trying to insert raw data %s\n", err),
			})
			return
		}

		if valid {
			storeLDDS45CalibratedData(ctx, server.MongoDb, sensor.Id, data.LDDS45, receivedAtTime)
		}

	default:
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"status": fmt.Sprintf("For sensor: %s unknown model type to handle: %s\n", sensor.Id, sensor.Model),
		})
		return
	}

	// Respond with a success status
	c.JSON(http.StatusOK, gin.H{"message": "Webhook received successfully"})
}

func storeS2120CalibratedData() error {

	return nil
}

func storeLDDS45CalibratedData(
	ctx context.Context,
	mongoDb MongoDatabase,
	sensorId string,
	data *LDDS45RawData,
	receivedAtTime primitive.DateTime) error {
	/*
		TODO: Current version simply finds an existing configuration for a sensor. Need to find the "latest"
		and one which hasn't been completed.
	*/
	sensorConfig := SensorConfiguration{}
	if err := GetSensorConfigurationCollection(mongoDb).FindOne(ctx, bson.M{
		"sensor": sensorId,
	}, &sensorConfig); err != nil {
		// 404 is a successful return
		if err == mongo.ErrNoDocuments {
			return nil
		}

		return err
	}

	log.Printf("Sensor configuration: %s found\n", sensorConfig.Id)
	// find the asset attached.

	asset := Asset{}
	if err := GetAssetsCollection(mongoDb).FindOne(ctx, bson.M{
		"_id": sensorConfig.Asset,
	}, &asset); err != nil {
		// 404 is a successful return
		if err == mongo.ErrNoDocuments {
			return nil
		}

		return err
	}

	if asset.Metrics != nil {
		// handle volume
		if asset.Metrics.Volume != nil {
			offset := 0.0

			if sensorConfig.Offset != nil && sensorConfig.Offset.Distance != nil {
				offset = sensorConfig.Offset.Distance.Distance
				// For this sensor need the offset to be in metres
				switch sensorConfig.Offset.Distance.Units {
				case MM_METRE:
					offset = offset / 1000
				case CM_METRE:
					offset = offset / 100
				case METRES:
					// ignore
				default:
					return fmt.Errorf("Unexpected units for a distance measurement: %s\n", sensorConfig.Offset.Distance.Units)
				}
			}

			distanceSplit := strings.Split(data.Distance, " ")
			// we don't need to dynamically handle units - sensor type will always generate the same units
			distanceInMmsString, _ := distanceSplit[0], distanceSplit[1]

			distanceInMms, err := strconv.ParseFloat(distanceInMmsString, 64)
			if err != nil {
				fmt.Println("Error:", err)
				return fmt.Errorf("Error converting LDDS45RawData distance to float %w", err)
			}

			distanceInM := distanceInMms / 1000

			/*
				Depth of the cylinder
				TODO: need 'offset'. sensor is installed in the roof of the water tank, pointing down.
				Positive offset is how much ABOVE the overflow pipe the sensor is located
				Need to know the offset value from the "top" of the tank. Top in the case is the top of the overflow output pipeline.
			*/
			cylinderDepth := asset.Metrics.Volume.Height - distanceInM - offset

			volume := asset.Metrics.Volume.CalcVolume(cylinderDepth)

			// 1 m³ = 1,000 litres
			litres := math.Round((volume*1000)*100) / 100

			calibrated := CalibratedData{
				Timestamp: receivedAtTime,
				Sensor:    sensorId,
				DataPoints: CalibratedDataPoints{
					Volume: &CalibratedDataType{
						Data:  litres,
						Units: METRES_CUBE,
					},
				},
			}

			_, err = GetCalibratedDataCollection(mongoDb).InsertOne(ctx, calibrated)
			if err != nil {
				return fmt.Errorf("Error trying to insert calibrated data %w\n", err)
			}
		}
	}

	return nil
}
