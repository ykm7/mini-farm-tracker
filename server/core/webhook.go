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

/*
https://www.thethingsindustries.com/docs/the-things-stack/concepts/data-formats/#uplink-messages
*/
// type UplinkMessage struct {
// 	EndDeviceIDs struct {
// 		DeviceID       string `json:"device_id"`
// 		ApplicationIDs struct {
// 			ApplicationID string `json:"application_id"`
// 		} `json:"application_ids"`
// 		// // DevEUI of the end device (eg: 0004A30B001C0530)
// 		DevEui string `json:"dev_eui"`
// 	} `json:"end_device_ids"`
// 	// // ISO 8601 UTC timestamp at which the message has been received by the Application Server (eg: "2020-02-12T15:15...")
// 	ReceivedAt    string `json:"received_at"`
// 	UplinkMessage struct {
// 		FPort uint32 `json:"f_port"`
// 		// // Frame payload (Base64)
// 		FrmPayload []byte `json:"frm_payload"`
// 		// Decoded payload object, decoded by the device payload formatter
// 		DecodedPayload map[string]interface{} `json:"decoded_payload"`
// 		RxMetadata     []struct {
// 			GatewayIDs struct {
// 				GatewayID string `json:"gateway_id"`
// 				EUI       string `json:"eui"`
// 			} `json:"gateway_ids"`
// 			// ISO 8601 UTC timestamp at which the uplink has been received by the gateway (et: "2020-02-12T15:15:45.787Z")
// 			Time         string  `json:"time"`
// 			Timestamp    int64   `json:"timestamp"`
// 			RSSI         int     `json:"rssi"`
// 			ChannelRSSI  int     `json:"channel_rssi"`
// 			SNR          float64 `json:"snr"`
// 			UplinkToken  string  `json:"uplink_token"`
// 			ChannelIndex int     `json:"channel_index"`
// 			Location     struct {
// 				Latitude  float64 `json:"latitude"`
// 				Longitude float64 `json:"longitude"`
// 				Altitude  int     `json:"altitude"`
// 				Source    string  `json:"source"`
// 			} `json:"location"`
// 		} `json:"rx_metadata"`
// 	} `json:"uplink_message"`
// }

type UplinkMessage struct {
	EndDeviceIDs struct {
		DeviceID       *string `json:"device_id,omitempty"`
		ApplicationIDs struct {
			ApplicationID *string `json:"application_id,omitempty"`
		} `json:"application_ids"`
		DevEUI  *string `json:"dev_eui,omitempty"`
		JoinEUI *string `json:"join_eui,omitempty"`
		DevAddr *string `json:"dev_addr,omitempty"`
	} `json:"end_device_ids"`
	CorrelationIDs *[]string `json:"correlation_ids,omitempty"`
	ReceivedAt     *string   `json:"received_at,omitempty"`
	UplinkMessage  struct {
		SessionKeyID   *string                `json:"session_key_id,omitempty"`
		FCount         *int                   `json:"f_cnt,omitempty"`
		FPort          *int                   `json:"f_port,omitempty"`
		FrmPayload     *string                `json:"frm_payload,omitempty"`
		DecodedPayload map[string]interface{} `json:"decoded_payload,omitempty"`
		RxMetadata     []struct {
			GatewayIDs struct {
				GatewayID *string `json:"gateway_id,omitempty"`
				EUI       *string `json:"eui,omitempty"`
			} `json:"gateway_ids"`
			Time         *string  `json:"time,omitempty"`
			Timestamp    *int64   `json:"timestamp,omitempty"`
			RSSI         *int     `json:"rssi,omitempty"`
			ChannelRSSI  *int     `json:"channel_rssi,omitempty"`
			SNR          *float64 `json:"snr,omitempty"`
			UplinkToken  *string  `json:"uplink_token,omitempty"`
			ChannelIndex *int     `json:"channel_index,omitempty"`
			Location     struct {
				Latitude  *float64 `json:"latitude,omitempty"`
				Longitude *float64 `json:"longitude,omitempty"`
				Altitude  *int     `json:"altitude,omitempty"`
				Source    *string  `json:"source,omitempty"`
			} `json:"location"`
		} `json:"rx_metadata,omitempty"`
		Settings struct {
			DataRate struct {
				Lora struct {
					Bandwidth       *int `json:"bandwidth,omitempty"`
					SpreadingFactor *int `json:"spreading_factor,omitempty"`
				} `json:"lora"`
			} `json:"data_rate"`
			CodingRate *string `json:"coding_rate,omitempty"`
			Frequency  *string `json:"frequency,omitempty"`
			Timestamp  *int64  `json:"timestamp,omitempty"`
			Time       *string `json:"time,omitempty"`
		} `json:"settings"`
		ReceivedAt      *string `json:"received_at,omitempty"`
		ConsumedAirtime *string `json:"consumed_airtime,omitempty"`
		Locations       map[string]struct {
			Latitude  *float64 `json:"latitude,omitempty"`
			Longitude *float64 `json:"longitude,omitempty"`
			Altitude  *int     `json:"altitude,omitempty"`
			Source    *string  `json:"source,omitempty"`
		} `json:"locations,omitempty"`
		VersionIDs struct {
			BrandID         *string `json:"brand_id,omitempty"`
			ModelID         *string `json:"model_id,omitempty"`
			HardwareVersion *string `json:"hardware_version,omitempty"`
			FirmwareVersion *string `json:"firmware_version,omitempty"`
			BandID          *string `json:"band_id,omitempty"`
		} `json:"version_ids"`
		NetworkIDs struct {
			NetID     *string `json:"net_id,omitempty"`
			TenantID  *string `json:"tenant_id,omitempty"`
			ClusterID *string `json:"cluster_id,omitempty"`
		} `json:"network_ids"`
		Simulated bool `json:"simulated"` // Keep as is since bool can't be nil
	} `json:"uplink_message"`
}

func handleWebhook(c *gin.Context, envs *environmentVariables, mongoDb MongoDatabase, sensorCache map[string]Sensor) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	apiKey := c.GetHeader("X-Downlink-Apikey")
	if apiKey == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing X-Downlink-Apikey header"})
		return
	}

	// Verify API Sign
	if apiKey != envs.ttn_webhhook_api {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Webhook env is invalid"})
		return
	}

	var uplinkMessage UplinkMessage
	if err := c.ShouldBindJSON(&uplinkMessage); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sensor, exists := sensorCache[*uplinkMessage.EndDeviceIDs.DeviceID]
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
		data := &LDDS45RawData{}
		err = json.Unmarshal(jsonData, &data)
		if err != nil || data == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status": fmt.Sprintf("Error casting the decoded json: %v to expected data type for: %s", jsonData, LDDS45),
			})
			return
		}

		valid := data.DetermineValid()
		dataPayload := RawData[LDDS45RawData]{
			Timestamp: receivedAtTime,
			Sensor:    &sensor.Id,
			Data:      *data,
			Valid:     valid,
		}
		_, err := GetRawDataCollection[LDDS45RawData](mongoDb).InsertOne(ctx, dataPayload)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status": fmt.Sprintf("Error trying to insert raw data %s\n", err),
			})
			return
		}

		if valid {
			storeLDDS45CalibratedData(ctx, mongoDb, sensor.Id, data, receivedAtTime)
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
				Data:      litres,
				Units:     METRES_CUBE,
			}

			_, err = GetCalibratedDataCollection(mongoDb).InsertOne(ctx, calibrated)
			if err != nil {
				return fmt.Errorf("Error trying to insert calibrated data %w\n", err)
			}
		}
	}

	return nil
}
