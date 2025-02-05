package core

import (
	"fmt"
	"math"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const DATABASE_NAME string = "db"

type SENSOR_MODELS string

const (
	LDDS45 SENSOR_MODELS = "LDDS45"
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
	MM_METRE    UNITS = "mm"
	CM_METRE    UNITS = "cm"
	METRES      UNITS = "m"
	METRES_CUBE UNITS = "m³"
	LITRES      UNITS = "L"
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
*/
type LDDS45RawData struct {
	Battery      float64 `json:"Bat"`      // units are 'mv'
	Distance     string  `json:"Distance"` // units are 'mm'
	InterruptPin uint8   `json:"Interrupt_flag"`
	Temperature  string  `json:"TempC_DS18B20"` // units are 'c'
	SensorFlag   uint8   `json:"Sensor_flag"`
}

func (lDDS45RawData *LDDS45RawData) determineValid() bool {
	distanceSplit := strings.Split(lDDS45RawData.Distance, " ")
	_, units := distanceSplit[0], distanceSplit[1]

	_, ok := StringToUnits(units)
	return ok == nil
}

// Just to test the generics
type RandomRawData struct {
}

type CalibratedData struct {
	Timestamp primitive.DateTime `bson:"timestamp"`
	Sensor    string             `bson:"sensor"`
	Data      float64            `bson:"data"`
	Units     UNITS              `bson:"units"`
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
