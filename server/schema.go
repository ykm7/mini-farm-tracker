package main

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
	Id primitive.ObjectID `bson:"_id"`
}

type RawData struct {
	Id        primitive.ObjectID `bson:"_id"`
	Timestamp primitive.DateTime `bson:"timestamp"`
	Sensor    primitive.ObjectID `bson:"_id"`
}

type CalibrateddData struct {
	Id        primitive.ObjectID `bson:"_id"`
	Timestamp primitive.DateTime `bson:"timestamp"`
	Sensor    primitive.ObjectID `bson:"_id"`
}

type SensorConfiguration struct {
	Id     primitive.ObjectID `bson:"_id"`
	Sensor primitive.ObjectID `bson:"_id"`
}
