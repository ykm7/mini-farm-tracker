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
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
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
			if err := storeLDDS45CalibratedData(ctx, server.MongoDb, sensor.Id, data.LDDS45, receivedAtTime); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"status": fmt.Sprintf("Error trying to process calibrated data for %s %s\n", sensor.Model, err),
				})
				return
			}
		}

	case S2120:
		data := SensorData{
			S2120: &S2120RawData{},
		}

		err = json.Unmarshal(jsonData, &data.S2120)

		// err = data.S2120.Unmarshal(jsonData)

		if err != nil || data.S2120 == nil {
			log.Printf("For payload (as string) %s and expected sensor %s have error %v", string(jsonData), S2120, err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
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
			if err := storeS2120CalibratedData(ctx, server.MongoDb, sensor.Id, data.S2120, receivedAtTime); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"status": fmt.Sprintf("Error trying to process calibrated data for %s %s\n", sensor.Model, err),
				})
				return
			}
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

func storeS2120CalibratedData(
	ctx context.Context,
	mongoDb MongoDatabase,
	sensorId string,
	data *S2120RawData,
	receivedAtTime primitive.DateTime) error {
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

	/**
	Uniquely (atleast, in comparision to the LDDS45 sensor, there is no modification prior to storing the values)
	However, they are to be "flattened" to be accessible via a single flat struct instead of within the raw messages slices

	TODO: Not entirely true, a configuration must indicate a reporting rate of 10mins.
	From the below it is more clear that is less of a dynamic configuration and more of a requirement for these values to be set correctly.

	https://files.seeedstudio.com/products/SenseCAP/101990961_SenseCAP%20S2120/SenseCAP%20S2120%20LoRaWAN%208-in-1%20Weather%20Station%20User%20Guide.pdf
	13.3 How to obtain the cumulative rainfall from the past ten minutes?
	1) Connect S2120 to SenseCAP Mate App and set the uplink interval to 10 min.
	2) Divide the uploaded rainfall intensity by 6 to get the cumulative rainfall from
	the past ten minutes.
	The
	uploaded rainfall intensity data (mm/h) of S2120 is derived by multiplying the
	cumulative rainfall (mm) from the past ten minutes by 6. Therefore, setting the
	interval to 10 min will provide the actual cumulative rainfall value.

	https://forum.seeedstudio.com/t/sensecap-weather-sensor-rain-data/270688
	Every minute the device calculates the cumulative rainfall of the past 10 minutes,
	which is then multiplied by 6 to derive the mm/hr value representing the rainfall intensity.
	So to obtain the cumulative rainfall, you can connect the device to the SenseCAP Mate app and set
	the uplink interval to 10 minutes. The device will upload the rain intensity data calculated at the
	last minute. By dividing the rainfall intensity data by 6, you can obtain the cumulative rainfall of
	the past ten minutes. Adding the cumulative rainfall every ten minutes will give you the total rainfall of an hour/ a day.

	If you want to check the cumulative rainfall directly, you can also use the following two devices: S700+S2100
	*/
	parsingErr := fmt.Errorf("errors:\n")
	additionalErrors := fmt.Errorf("")
	dataPoint := CalibratedDataPoints{}
	for _, msg := range data.Messages {
		switch t := msg.(type) {
		case *S2120RawDataMeasurement:
			value := t.MeasurementValue

			switch v := t.Type; v {
			case AirTemperature:
				if v, ok := value.(float64); ok {
					dataPoint.AirTemperature = &CalibratedDataType{
						Data:  v,
						Units: DEGREE_C,
					}
				} else {
					additionalErrors = fmt.Errorf("%w for %s cannot parse value %d as the expected type (float64)\n", additionalErrors, string(AirTemperature), value)
				}

			case AirHumidity:
				if v, ok := value.(float64); ok {
					dataPoint.AirHumidity = &CalibratedDataType{
						Data:  v,
						Units: AIR_HUMIDITY,
					}
				} else {
					additionalErrors = fmt.Errorf("%w for %s cannot parse value %d as the expected type (float64)\n", additionalErrors, string(AirHumidity), value)
				}

			case LightIntensity:
				if v, ok := value.(float64); ok {
					dataPoint.LightIntensity = &CalibratedDataType{
						Data:  v,
						Units: LUX,
					}

				} else {
					additionalErrors = fmt.Errorf("%w for %s cannot parse value %d as the expected type (float64)\n", additionalErrors, string(LightIntensity), value)
				}

			case UVIndex:
				if v, ok := value.(float64); ok {
					dataPoint.UVIndex = &CalibratedDataType{
						Data:  v,
						Units: UV_INDEX,
					}
				} else {
					additionalErrors = fmt.Errorf("%w for %s cannot parse value %d as the expected type (float64)\n", additionalErrors, string(UVIndex), value)
				}
			case WindSpeed:
				if v, ok := value.(float64); ok {
					dataPoint.WindSpeed = &CalibratedDataType{
						Data:  v,
						Units: M_PER_SEC,
					}
				} else {
					additionalErrors = fmt.Errorf("%w for %s cannot parse value %d as the expected type (float64)\n", additionalErrors, string(WindSpeed), value)
				}

			case WindDirectionSensor:
				if v, ok := value.(float64); ok {
					dataPoint.WindDirection = &CalibratedDataType{
						Data:  v,
						Units: DEGREE,
					}
				} else {
					additionalErrors = fmt.Errorf("%w for %s cannot parse value %d as the expected type (float64)\n", additionalErrors, string(WindDirectionSensor), value)
				}

			case RainGauge:
				if v, ok := value.(float64); ok {
					dataPoint.RainGauge = &CalibratedDataType{
						Data:  v,
						Units: MM_PER_HOUR,
					}

					/**
					1) Connect S2120 to SenseCAP Mate App and set the uplink interval to 10 min.
					2) Divide the uploaded rainfall intensity by 6 to get the cumulative rainfall from the past ten minutes
					*/
					cumulatedRainfallOver10Mins := v / 6
					dataPoint.RainAccumulation = &CalibratedDataType{
						Data:  cumulatedRainfallOver10Mins,
						Units: MM_METRE,
					}

				} else {
					additionalErrors = fmt.Errorf("%w for %s cannot parse value %d as the expected type (float64)\n", additionalErrors, string(RainGauge), value)
				}

			case BarometricPressure:
				if v, ok := value.(float64); ok {
					dataPoint.BarometricPressure = &CalibratedDataType{
						Data:  v,
						Units: PRESSURE,
					}
				} else {
					additionalErrors = fmt.Errorf("%w for %s cannot parse value %d as the expected type (float64)\n", additionalErrors, string(BarometricPressure), value)
				}

			case PeakWindGust:
				if v, ok := value.(float64); ok {
					dataPoint.PeakWindGust = &CalibratedDataType{
						Data:  v,
						Units: M_PER_SEC,
					}
				} else {
					additionalErrors = fmt.Errorf("%w for %s cannot parse value %d as the expected type (float64)\n", additionalErrors, string(BarometricPressure), value)
				}

				// This is a continuously incrementing value. I prefer to tally the values in small interval and use the scheduled logic to gather all.
				// case RainAccumulation:
				// 	if v, ok := value.(float64); ok {
				// 		dataPoint.RainAccumulation = &CalibratedDataType{
				// 			Data:  v,
				// 			Units: MM_METRE,
				// 		}
				// 	} else {
				// 		additionalErrors = fmt.Errorf("%w for %s cannot parse value %d as the expected type (float64)\n", additionalErrors, string(BarometricPressure), value)
				// 	}
			}
		}
	}

	if additionalErrors.Error() != "" {
		return fmt.Errorf("%w %w", parsingErr, additionalErrors)
	}

	calibrated := CalibratedData{
		Timestamp:  receivedAtTime,
		Sensor:     sensorId,
		DataPoints: dataPoint,
	}

	_, err := GetCalibratedDataCollection(mongoDb).InsertOne(ctx, calibrated)
	if err != nil {
		return fmt.Errorf("Error trying to insert calibrated data %w\n", err)
	}

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
				log.Println("Error:", err)
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
