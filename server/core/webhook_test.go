package core

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

// Quite a few of these functions should be moved to a common location. No need to worry about it until we have more tests.
func createMockUplinkMessage(deviceId, receivedAt string, decodedPayload map[string]interface{}) UplinkMessage {
	return UplinkMessage{
		EndDeviceIDs: struct {
			DeviceID       *string "json:\"device_id,omitempty\""
			ApplicationIDs struct {
				ApplicationID *string "json:\"application_id,omitempty\""
			} "json:\"application_ids\""
			DevEUI  *string "json:\"dev_eui,omitempty\""
			JoinEUI *string "json:\"join_eui,omitempty\""
			DevAddr *string "json:\"dev_addr,omitempty\""
		}{
			DeviceID: Ptr(deviceId),
		},
		ReceivedAt: Ptr(receivedAt),
		UplinkMessage: struct {
			SessionKeyID   *string                "json:\"session_key_id,omitempty\""
			FCount         *int                   "json:\"f_cnt,omitempty\""
			FPort          *int                   "json:\"f_port,omitempty\""
			FrmPayload     *string                "json:\"frm_payload,omitempty\""
			DecodedPayload map[string]interface{} "json:\"decoded_payload,omitempty\""
			RxMetadata     []struct {
				GatewayIDs struct {
					GatewayID *string "json:\"gateway_id,omitempty\""
					EUI       *string "json:\"eui,omitempty\""
				} "json:\"gateway_ids\""
				Time         *string  "json:\"time,omitempty\""
				Timestamp    *int64   "json:\"timestamp,omitempty\""
				RSSI         *int     "json:\"rssi,omitempty\""
				ChannelRSSI  *int     "json:\"channel_rssi,omitempty\""
				SNR          *float64 "json:\"snr,omitempty\""
				UplinkToken  *string  "json:\"uplink_token,omitempty\""
				ChannelIndex *int     "json:\"channel_index,omitempty\""
				Location     struct {
					Latitude  *float64 "json:\"latitude,omitempty\""
					Longitude *float64 "json:\"longitude,omitempty\""
					Altitude  *int     "json:\"altitude,omitempty\""
					Source    *string  "json:\"source,omitempty\""
				} "json:\"location\""
			} "json:\"rx_metadata,omitempty\""
			Settings struct {
				DataRate struct {
					Lora struct {
						Bandwidth       *int "json:\"bandwidth,omitempty\""
						SpreadingFactor *int "json:\"spreading_factor,omitempty\""
					} "json:\"lora\""
				} "json:\"data_rate\""
				CodingRate *string "json:\"coding_rate,omitempty\""
				Frequency  *string "json:\"frequency,omitempty\""
				Timestamp  *int64  "json:\"timestamp,omitempty\""
				Time       *string "json:\"time,omitempty\""
			} "json:\"settings\""
			ReceivedAt      *string "json:\"received_at,omitempty\""
			ConsumedAirtime *string "json:\"consumed_airtime,omitempty\""
			Locations       map[string]struct {
				Latitude  *float64 "json:\"latitude,omitempty\""
				Longitude *float64 "json:\"longitude,omitempty\""
				Altitude  *int     "json:\"altitude,omitempty\""
				Source    *string  "json:\"source,omitempty\""
			} "json:\"locations,omitempty\""
			VersionIDs struct {
				BrandID         *string "json:\"brand_id,omitempty\""
				ModelID         *string "json:\"model_id,omitempty\""
				HardwareVersion *string "json:\"hardware_version,omitempty\""
				FirmwareVersion *string "json:\"firmware_version,omitempty\""
				BandID          *string "json:\"band_id,omitempty\""
			} "json:\"version_ids\""
			NetworkIDs struct {
				NetID     *string "json:\"net_id,omitempty\""
				TenantID  *string "json:\"tenant_id,omitempty\""
				ClusterID *string "json:\"cluster_id,omitempty\""
			} "json:\"network_ids\""
			Simulated bool "json:\"simulated\""
		}{
			DecodedPayload: decodedPayload,
		},
	}
}

func Ptr[T any](value T) *T {
	return &value
}

func setupMockCollections[T RawDataType](mongoDb MongoDatabase, cw *CollectionToData[T]) {
	cw.sensors.collection = GetSensorCollection(mongoDb)
	cw.rawData.collection = GetRawDataCollection[T](mongoDb)
	cw.sensorConfigurations.collection = GetSensorConfigurationCollection(mongoDb)
	cw.calibratedData.collection = GetCalibratedDataCollection(mongoDb)
	cw.assets.collection = GetAssetsCollection(mongoDb)
}

func InsertData[T RawDataType](cw *CollectionToData[T]) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, d := range cw.sensors.data {
		cw.sensors.collection.InsertOne(ctx, d)
	}

	for _, d := range cw.rawData.data {
		cw.rawData.collection.InsertOne(ctx, d)
	}

	for _, d := range cw.sensorConfigurations.data {
		cw.sensorConfigurations.collection.InsertOne(ctx, d)
	}

	for _, d := range cw.calibratedData.data {
		cw.calibratedData.collection.InsertOne(ctx, d)
	}

	for _, d := range cw.assets.data {
		cw.assets.collection.InsertOne(ctx, d)
	}
}

func ClearCollections[T RawDataType](dbWrap *MongoDatabaseImpl, cw *CollectionToData[T]) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Additional Sanity check

	testMongoName := dbWrap.Db.Name()
	if testMongoName != "test_db" {
		log.Panicln("While clearing up test mongo collections, database name is NOT 'test_db'")
	}

	cw.sensors.collection.DeleteMany(ctx, bson.D{})
	cw.rawData.collection.DeleteMany(ctx, bson.D{})
	cw.sensorConfigurations.collection.DeleteMany(ctx, bson.D{})
	cw.calibratedData.collection.DeleteMany(ctx, bson.D{})
	cw.assets.collection.DeleteMany(ctx, bson.D{})
}

type CollectionWrapper[T any] struct {
	collection MongoCollection[T]
	data       []T
}

type CollectionToData[T RawDataType] struct {
	sensors              CollectionWrapper[Sensor]
	rawData              CollectionWrapper[RawData[T]]
	sensorConfigurations CollectionWrapper[SensorConfiguration]
	calibratedData       CollectionWrapper[CalibratedData]
	assets               CollectionWrapper[Asset]
}

func Test_handleWebhook(t *testing.T) {

	// initEnvironmentVariables := &environmentVariables{}
	// TODO: if the cache was queried multiple times within the function possible to open
	// it up to race conditions. Worth being aware of when adding the configuration stuff in the future
	// initsensorCache := map[string]Sensor{}

	db, deferFn := MockSetupMongo(context.TODO())
	mongoDb := &MongoDatabaseImpl{Db: db}
	defer deferFn()

	MOCK_DEVICE_ID := "MOCK_DEVICE_ID"
	MOCK_SENSOR_ID := "112233445566778899"
	MOCK_RECEIVED_AT := "2025-01-28T03:14:25.480959673Z"

	defaultHeader := http.Header{
		"Content-Type": {"application/json"},
	}

	type args struct {
		uplinkMessage     UplinkMessage
		additionalHeaders http.Header
		envs              *environmentVariables
		sensorCache       map[string]Sensor
		preData           CollectionToData[LDDS45RawData]
	}
	type expected struct {
		code     int
		message  map[string]string
		postData CollectionToData[LDDS45RawData]
	}

	tests := []struct {
		name     string
		args     args
		expected expected
	}{
		{
			name: "In-valid 'LDDS45RawData' data",
			args: args{
				additionalHeaders: http.Header{
					"X-Downlink-Apikey": []string{"RANDOM_TEST_KEY"},
				},
				envs: &environmentVariables{
					ttn_webhhook_api: "RANDOM_TEST_KEY",
				},
				uplinkMessage: createMockUplinkMessage(
					MOCK_DEVICE_ID,
					MOCK_RECEIVED_AT,
					map[string]interface{}{
						"Bat":            3.413,
						"Distance":       "1404 mm",
						"Interrupt_flag": 0,
						"Sensor_flag":    1,
						"TempC_DS18B20":  "0.00",
					},
				),
				sensorCache: map[string]Sensor{
					MOCK_DEVICE_ID: {
						Id:    MOCK_SENSOR_ID,
						Model: LDDS45,
					},
				},
			},
			expected: expected{
				code: http.StatusOK,
				message: map[string]string{
					"message": "Webhook received successfully",
				},
			},
		},
		{
			name: "No 'X-Downlink-Apikey' header provided",
			args: args{},
			expected: expected{
				code: http.StatusBadRequest,
				message: map[string]string{
					"error": "Missing X-Downlink-Apikey header",
				},
			},
		},
		{
			name: "Mismatch 'X-Downlink-Apikey' header provided",
			args: args{
				additionalHeaders: http.Header{
					"X-Downlink-Apikey": []string{"RANDOM_TEST_KEY"},
				},
				envs: &environmentVariables{},
			},
			expected: expected{
				code: http.StatusBadRequest,
				message: map[string]string{
					"error": "Webhook env is invalid",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: mongo setup ? - no enforced schema or anything so limited scope here.
			// TODO: mongo collection/data tear down after the test is completed.

			defer ClearCollections(mongoDb, &tt.args.preData)

			setupMockCollections(mongoDb, &tt.args.preData)

			w := httptest.NewRecorder()
			mockCtx := MockGinContext(w)

			// Allow for addition headers - needed for dynamic api key header
			if tt.args.additionalHeaders != nil {
				for key, values := range defaultHeader {
					for _, value := range values {
						tt.args.additionalHeaders.Add(key, value)
					}
				}
			}

			MockJsonPost(mockCtx, tt.args.uplinkMessage)
			MockContextAdd(mockCtx, tt.args.additionalHeaders.Clone())

			handleWebhook(mockCtx, tt.args.envs, mongoDb, tt.args.sensorCache)

			assert.Equal(t, tt.expected.code, w.Code)

			expectedJson, err := mapToJSONString(tt.expected.message)
			assert.ErrorIs(t, err, nil)
			assert.Equal(t, expectedJson, w.Body.String())

			// Check that the data is valid
		})
	}
}

// func Test_storeLDDS45CalibratedData(t *testing.T) {
// 	/*
// 		All in all, I don't like this solution.
// 		TODO: Will be implementing previously implemented Mongodb solution.
// 		Within testContainer
// 		Have an "init" collections documents which are written to mongo
// 		Have an "post" collection documents which are the expected documents to be found.
// 		Compare.
// 	*/
// 	EXPECTED_SENSOR_ID := "EXPECTED_SENSOR_ID"

// 	// mockSensorCollection := &MockMongoCollection[any]{
// 	// 	FindOneFn: func(ctx context.Context, filter interface{}, result *any) error {
// 	// 		*result = SensorConfiguration{
// 	// 			Sensor: "sensor id",
// 	// 		}
// 	// 		return nil
// 	// 	},
// 	// }

// 	// mockAssetCollection := &MockMongoCollection[any]{
// 	// 	FindOneFn: func(ctx context.Context, filter interface{}, result *any) error {
// 	// 		*result = Asset{}
// 	// 		return nil
// 	// 	},
// 	// }

// 	// mockCalibratedDataCollection := &mockSensorCollection[any]{
// 	// 	InsertOneFn: func(ctx context.Context, document T) (*mongo.InsertOneResult, error) {

// 	// 		return nil, nil
// 	// 	},
// 	// }

// 	mockDb := NewMockMongoDatabase()
// 	// mockDb.SetCollection(string(SENSOR_CONFIGURATIONS_COLLECTION), mockSensorCollection)
// 	// mockDb.SetCollection(string(ASSETS_COLLECTION), mockAssetCollection)
// 	// mockDb.SetCollection(string(CALIBRATED_DATA_COLLECTION), mockCalibratedDataCollection)

// 	type args struct {
// 		ctx            context.Context
// 		mongoDb        MongoDatabase
// 		sensorId       string
// 		data           *LDDS45RawData
// 		receivedAtTime primitive.DateTime
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name: "Successful insertion of data.",
// 			args: args{
// 				mongoDb:        mockDb,
// 				sensorId:       EXPECTED_SENSOR_ID,
// 				data:           &LDDS45RawData{},
// 				ctx:            context.TODO(),
// 				receivedAtTime: primitive.NewDateTimeFromTime(time.Now()),
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := storeLDDS45CalibratedData(tt.args.ctx, tt.args.mongoDb, tt.args.sensorId, tt.args.data, tt.args.receivedAtTime); (err != nil) != tt.wantErr {
// 				t.Errorf("storeLDDS45CalibratedData() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }
