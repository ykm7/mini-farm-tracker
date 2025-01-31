package main

import (
	"mini-farm-tracker-server/core"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OldRawDataType interface {
	OldLDDS45RawData | OldRandomRawData
}

type OldRawData[T OldRawDataType] struct {
	Timestamp primitive.DateTime `bson:"timestamp"`
	Sensor    *string            `bson:"sensor,omitempty"`
	Valid     bool               `bson:"valid,omitempty"`
	Data      T                  `bson:"data"`
}

type OldRawDataFns interface {
	DetermineValid() bool
}

/*
Raw result from within the TTN payload. [mini-farm-tracker-server] [2025-01-21 07:14:44] 2025/01/21 07:14:44 'Decoded' payload: map[Bat:3.402 Distance:1752 mm Interrupt_flag:0 Sensor_flag:1 TempC_DS18B20:0.00]

Uplink formatter added to within the TTN when selecting the device from the repository.
*/
type OldLDDS45RawData struct {
	Battery      float64 `json:"Bat"`      // units are 'mv'
	Distance     string  `json:"Distance"` // units are 'mm'
	InterruptPin uint8   `json:"Interrupt_flag"`
	Temperature  string  `json:"TempC_DS18B20"` // units are 'c'
	SensorFlag   uint8   `json:"Sensor_flag"`
}

func (lDDS45RawData *OldLDDS45RawData) DetermineValid() bool {
	distanceSplit := strings.Split(lDDS45RawData.Distance, " ")
	_, units := distanceSplit[0], distanceSplit[1]

	_, ok := core.StringToUnits(units)
	return ok == nil
}

// Just to test the generics
type OldRandomRawData struct {
}

func main() {

}
