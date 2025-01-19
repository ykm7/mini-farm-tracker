package core

import "go.mongodb.org/mongo-driver/bson/primitive"

const DATABASE_NAME string = "db"

type DB_COLLECTIONS string

const (
	RAW_DATA_COLLECTION              DB_COLLECTIONS = "raw_data"
	CALIBRATED_DATA_COLLECTION       DB_COLLECTIONS = "calibrated_data"
	SENSOR_CONFIGURATIONS_COLLECTION DB_COLLECTIONS = "sensor_configurations"
	SENSORS_COLLECTION               DB_COLLECTIONS = "sensors"
)

type Sensor struct {
	Id string `bson:"_id"`
}

type RawData struct {
	Id        primitive.ObjectID `bson:"_id,omitempty"`
	Timestamp primitive.DateTime `bson:"timestamp"`
	Sensor    string             `bson:"sensor"`
	Data      int64              `bson:"data"`
}

type CalibrateddData struct {
	Id        primitive.ObjectID `bson:"_id,omitempty"`
	Timestamp primitive.DateTime `bson:"timestamp"`
	Sensor    string             `bson:"sensor"`
	Data      int64              `bson:"data"`
}

type SensorConfiguration struct {
	Id        primitive.ObjectID `bson:"_id,omitempty"`
	Sensor    string             `bson:"sensor"`
	Timestamp primitive.DateTime `bson:"timestamp"`
}
