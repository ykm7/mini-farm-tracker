package core

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const DATABASE_NAME string = "db"

type SENSOR_MODELS string

const (
	LDDS45 SENSOR_MODELS = "LDDS45"
	S2120  SENSOR_MODELS = "S2120"
)

type DB_COLLECTIONS string

const (
	RAW_DATA_COLLECTION              DB_COLLECTIONS = "raw_data"
	CALIBRATED_DATA_COLLECTION       DB_COLLECTIONS = "calibrated_data"
	SENSOR_CONFIGURATIONS_COLLECTION DB_COLLECTIONS = "sensor_configurations"
	SENSORS_COLLECTION               DB_COLLECTIONS = "sensors"
	ASSETS_COLLECTION                DB_COLLECTIONS = "assets"
)

type UNITS string

const (
	UV_INDEX     UNITS = ""
	MM_METRE     UNITS = "mm"
	CM_METRE     UNITS = "cm"
	METRES       UNITS = "m"
	METRES_CUBE  UNITS = "m³"
	LITRES       UNITS = "L"
	MM_PER_HOUR  UNITS = "mm/hr"
	M_PER_SEC    UNITS = "m/s"
	DEGREE_C     UNITS = "℃"
	DEGREE       UNITS = "℃"
	PRESSURE     UNITS = "Pa"
	AIR_HUMIDITY UNITS = "%RH"
	LUX          UNITS = "Lux"
)

func StringToUnits(s string) (UNITS, error) {
	switch s {
	case string(MM_METRE):
		return MM_METRE, nil
	case string(CM_METRE):
		return CM_METRE, nil
	case string(METRES):
		return METRES, nil
	case string(METRES_CUBE):
		return METRES_CUBE, nil
	case string(LITRES):
		return LITRES, nil
	default:
		return "", fmt.Errorf("Cannot convert string [%s] to units", s)
	}
}

type AssetMetricsCylinderVolume struct {
	// Max static volume
	// This is not likely to be the manufactoring volume, but rather based on height of overflow outlet. Actually likely to be the same?
	Volume float64 `bson:"volume"`
	// Max static radius
	Radius float64 `bson:"radius"`
	// Max static volume
	Height float64 `bson:"height"`
}

func (cv *AssetMetricsCylinderVolume) CalcVolume(height float64) float64 {
	volume := math.Pi * math.Pow(cv.Radius, 2) * height
	return volume
}

/*
Represents static metrics tied to a asset/s
*/
type AssetMetrics struct {
	Volume *AssetMetricsCylinderVolume `bson:"volume,omitempty"`
}

type Asset struct {
	Id          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Description string             `bson:"description"`
	// To be used when authentication is added
	// User        primitive.ObjectID    `bson:"user"`
	Sensors *[]string     `bson:"sensors,omitempty"`
	Metrics *AssetMetrics `bson:"metrics,omitempty"`
}

type Sensor struct {
	Id          string        `bson:"_id"`
	Description string        `bson:"description"`
	Model       SENSOR_MODELS `bson:"model"`
}

type SensorData struct {
	LDDS45 *LDDS45RawData `bson:"LDDS45"`
	// further various sensor data types
	S2120 *S2120RawData `bson:"S2120"`
}

func (s *SensorData) DetermineValid() (bool, error) {
	if s.LDDS45 != nil {
		return s.LDDS45.determineValid(), nil
	}

	return false, fmt.Errorf("unknown sensor data to perform 'determineValid' on %v", s)
}

type RawData struct {
	Timestamp primitive.DateTime `bson:"timestamp"`
	Sensor    *string            `bson:"sensor,omitempty"`
	Valid     bool               `bson:"valid,omitempty"`
	Data      SensorData         `bson:"data"`
}

type RawDataFns interface {
	DetermineValid() bool
}

/*
Raw result from within the TTN payload. [mini-farm-tracker-server] [2025-01-21 07:14:44] 2025/01/21 07:14:44 'Decoded' payload: map[Bat:3.402 Distance:1752 mm Interrupt_flag:0 Sensor_flag:1 TempC_DS18B20:0.00]

Uplink formatter added to within the TTN when selecting the device from the repository.

TODO: Correct this in the future
JSON parsing for extracting from webhook
BSON for internal (and API)
Keeping the field names consistent for now. Need to migrate data.

Battery      float64 `json:"Bat" bson:"battery"`       // units are 'mv'
Distance     string  `json:"Distance" bson:"Distance"` // units are 'mm'
InterruptPin uint8   `json:"Interrupt_flag" bson:"interruptPin"`
Temperature  string  `json:"TempC_DS18B20" bson:"temperature"` // units are 'c'
SensorFlag   uint8   `json:"Sensor_flag" bson:"sensorFlag"`
*/
type LDDS45RawData struct {
	Battery      float64 `json:"Bat" bson:"battery"`       // units are 'mv'
	Distance     string  `json:"Distance" bson:"distance"` // units are 'mm'
	InterruptPin uint8   `json:"Interrupt_flag" bson:"interrupt_flag"`
	Temperature  string  `json:"TempC_DS18B20" bson:"tempC_DS18B20"` // units are 'c'
	SensorFlag   uint8   `json:"Sensor_flag" bson:"sensor_flag"`
}

func (lDDS45RawData *LDDS45RawData) determineValid() bool {
	distanceSplit := strings.Split(lDDS45RawData.Distance, " ")
	_, units := distanceSplit[0], distanceSplit[1]

	_, ok := StringToUnits(units)
	return ok == nil
}

/*
*
Based on decoder from:
https://github.com/Seeed-Solution/TTN-Payload-Decoder/blob/master/SenseCAP_S2120_Weather_Station_Decoder.js#L110
*/
type S2120RawData struct {
	Err      int               `json:"err" bson:"err"`
	Payload  string            `json:"payload" bson:"payload"`
	Valid    bool              `json:"valid" bson:"valid"`
	Messages []S2120RawDataMsg `json:"messages" bson:"messages"`
}

func (s *S2120RawData) UnmarshalJSON(data []byte) error {
	type Alias S2120RawData
	aux := &struct {
		Messages []json.RawMessage `json:"messages"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	s.Messages = make([]S2120RawDataMsg, len(aux.Messages))
	for idx, raw := range aux.Messages {
		var rawMeasurement S2120RawDataMeasurement
		if err := json.Unmarshal(raw, &rawMeasurement); err == nil {
			s.Messages[idx] = &rawMeasurement
			continue
		}

		var rawStatus S2120RawDataStatus
		if err := json.Unmarshal(raw, &rawStatus); err == nil {
			s.Messages[idx] = &rawStatus
			continue
		}

		return fmt.Errorf("unable to determine type for S2120RawData message %s", string(raw))
	}

	return nil
}

func (s *S2120RawData) UnmarshalBSON(data []byte) error {
	type Alias S2120RawData
	aux := &struct {
		Messages []bson.Raw `bson:"messages"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := bson.Unmarshal(data, &aux); err != nil {
		return err
	}

	s.Messages = make([]S2120RawDataMsg, len(aux.Messages))
	for idx, raw := range aux.Messages {
		var rawMeasurement S2120RawDataMeasurement
		if err := bson.Unmarshal(raw, &rawMeasurement); err == nil {
			s.Messages[idx] = &rawMeasurement
			continue
		}

		var rawStatus S2120RawDataStatus
		if err := bson.Unmarshal(raw, &rawStatus); err == nil {
			s.Messages[idx] = &rawStatus
			continue
		}

		return fmt.Errorf("unable to determine type for S2120RawData message %s", string(raw))
	}

	return nil
}

// func (s S2120RawData) MarshalBSON() ([]byte, error) {
// 	type Alias S2120RawData
// 	return bson.Marshal(&struct {
// 		Messages []bson.Raw `bson:"messages"`
// 		*Alias
// 	}{
// 		Alias: (*Alias)(&s),
// 	})
// }

// func (s *S2120RawData) Unmarshal(data []byte) error {
// 	type Alias S2120RawData
// 	aux := &struct {
// 		Messages []json.RawMessage `bson:"messages"`
// 		*Alias
// 	}{
// 		Alias: (*Alias)(s),
// 	}

// 	if err := json.Unmarshal(data, &aux); err != nil {
// 		return err
// 	}

// 	s.Messages = make([]S2120RawDataMsg, len(aux.Messages))
// 	for idx, raw := range aux.Messages {
// 		var rawMeasurement *S2120RawDataMeasurement
// 		if err := json.Unmarshal(raw, &rawMeasurement); err == nil {
// 			s.Messages[idx] = rawMeasurement
// 			continue
// 		}

// 		var rawStatus *S2120RawDataStatus
// 		if err := json.Unmarshal(raw, &rawStatus); err == nil {
// 			s.Messages[idx] = rawStatus
// 			continue
// 		}

// 		return fmt.Errorf("unable to determine type for S2120RawData message %s", string(raw))
// 	}

// 	return nil
// }

type S2120RawDataMsg interface {
	is()
}

type S2120RawDataMeasurementType string

const (
	AirTemperature      S2120RawDataMeasurementType = "Air Temperature"
	AirHumidity         S2120RawDataMeasurementType = "Air Humidity"
	LightIntensity      S2120RawDataMeasurementType = "Light Intensity"
	UVIndex             S2120RawDataMeasurementType = "UV Index"
	WindSpeed           S2120RawDataMeasurementType = "Wind Speed"
	WindDirectionSensor S2120RawDataMeasurementType = "Wind Direction Sensor"
	RainGauge           S2120RawDataMeasurementType = "Rain Gauge"
	BarometricPressure  S2120RawDataMeasurementType = "Barometric Pressure"
)

type S2120RawDataMeasurementError string

const (
	SensorErrorEvent S2120RawDataMeasurementError = "sensor_error_event"
)

type S2120RawDataMeasurement struct {
	MeasurementValue any                         `json:"measurementValue" bson:"measurementValue"`
	MeasurementId    string                      `json:"measurementId" bson:"measurementId"`
	Type             S2120RawDataMeasurementType `json:"type" bson:"type"`
}

func (s *S2120RawDataMeasurement) is() {}

type S2120RawDataStatus struct {
	/**
	string|number
	*/
	BatteryPercent  *any    `json:"Battery(%),omitempty" bson:"Battery(%),omitempty"`
	HardwareVersion *string `json:"Hardware Version,omitempty" bson:"Hardware Version,omitempty"`
	FirmwareVersion *string `json:"Firmware Version,omitempty" bson:"Firmware Version,omitempty"`
	MeasureInterval *int    `json:"MeasureInterval,omitempty" bson:"measureInterval,omitempty"`
	GpsInterval     *int    `json:"GpsInterval,omitempty" bson:"gpsInterval,omitempty"`
}

func (s *S2120RawDataStatus) is() {}

/*
Based on the error message:

	messages = [{
		measurementId: '4101', type: 'sensor_error_event', errCode: errorCode, descZh
	}]
*/
type S2120RawDataError struct {
	MeasurementValue any                          `bson:"measurementValue"`
	MeasurementId    string                       `bson:"measurementId"`
	Type             S2120RawDataMeasurementError `bson:"type"`
	ErrCode          [2]int                       `bson:"errCode"`
}

type S2120RawDataStatusMsg struct {
}

type CalibratedData struct {
	Timestamp  primitive.DateTime   `bson:"timestamp"`
	Sensor     string               `bson:"sensor"`
	DataPoints CalibratedDataPoints `bson:"dataPoints"`
}

/*
*
Available from the different:
https://cdn.shopify.com/s/files/1/1386/3791/files/SenseCAP_S2120_LoRaWAN_8-in-1_Weather_Station_User_Guide.pdf?v=1662178525
*/
type CalibratedDataPoints struct {
	Volume             *CalibratedDataType `json:"Volume,omitempty" bson:"volume,omitempty"`
	AirTemperature     *CalibratedDataType `json:"AirTemperature,omitempty" bson:"airTemperature,omitempty"`
	AirHumidity        *CalibratedDataType `json:"AirHumidity,omitempty" bson:"airHumidity,omitempty"`
	LightIntensity     *CalibratedDataType `json:"LightIntensity,omitempty" bson:"lightIntensity,omitempty"`
	UVIndex            *CalibratedDataType `json:"UvIndex,omitempty" bson:"uvIndex,omitempty"`
	WindSpeed          *CalibratedDataType `json:"WindSpeed,omitempty" bson:"windSpeed,omitempty"`
	WindDirection      *CalibratedDataType `json:"WindDirection,omitempty" bson:"windDirection,omitempty"`
	RainfallHourly     *CalibratedDataType `json:"RainfallHourly,omitempty" bson:"rainfallHourly,omitempty"`
	BarometricPressure *CalibratedDataType `json:"bBarometricPressure,omitempty" bson:"barometricPressure,omitempty"`
}

type CalibratedDataType struct {
	Data  float64 `bson:"data"`
	Units UNITS   `bson:"units"`
}

type SensorConfiguration struct {
	Id      primitive.ObjectID `bson:"_id,omitempty"`
	Sensor  string             `bson:"sensor"`
	Asset   primitive.ObjectID `bson:"asset"`
	Applied primitive.DateTime `bson:"applied"`
	// to indicate that this sensor is no longer applied.
	// Initially thought there would initially be another config to "take over" but this is cleaner
	Unapplied *primitive.DateTime `bson:"unapplied"`
	// Based on the installation, an offset based on the sensor type may need to be applied.
	Offset *struct {
		Distance *struct {
			Distance float64 `bson:"distance"`
			// Not required for logic, for for UI
			Units UNITS `bson:"units"`
		} `bson:"distance"`
	} `bson:"Offset,omitempty"`
}

/*
https://www.thethingsindustries.com/docs/the-things-stack/concepts/data-formats/#uplink-messages
*/
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
