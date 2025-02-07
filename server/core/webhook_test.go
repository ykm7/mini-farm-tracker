package core

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func mockConvertTimeStringToMongoTime(s string) primitive.DateTime {
	receivedAt, err := convertTimeStringToMongoTime(s)
	if err != nil {
		panic(err)
	}

	return receivedAt
}

func setupMockCollections(mongoDb MongoDatabase, cw *CollectionToData) {
	cw.sensors.collection = GetSensorCollection(mongoDb)
	cw.rawData.collection = GetRawDataCollection(mongoDb)
	cw.sensorConfigurations.collection = GetSensorConfigurationCollection(mongoDb)
	cw.calibratedData.collection = GetCalibratedDataCollection(mongoDb)
	cw.assets.collection = GetAssetsCollection(mongoDb)
}

func insertMockData(cw *CollectionToData) {
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

func validateDataExistingsWithinMockDb(t *testing.T, expected *DateExpectedToFind, cw *CollectionToData) {
	ctx := context.Background()

	if expected.sensors != nil {
		results, err := cw.sensors.collection.Find(ctx, nil)
		if err != nil {
			panic(err)
		}

		if diff := cmp.Diff(expected.sensors, results, cmpopts.SortSlices(func(a, b Asset) bool {
			return a.Name < b.Name
		})); diff != "" {
			t.Errorf("Slices mismatch (-want +got):\n%s", diff)
		}
	}

	if expected.rawData != nil {
		results, err := cw.rawData.collection.Find(ctx, nil)
		if err != nil {
			panic(err)
		}

		if diff := cmp.Diff(expected.rawData, results, cmpopts.SortSlices(func(a, b RawData) bool {
			return a.Timestamp < b.Timestamp
		})); diff != "" {
			t.Errorf("Slices mismatch (-want +got):\n%s", diff)
		}
	}

	if expected.sensorConfigurations != nil {
		results, err := cw.sensorConfigurations.collection.Find(ctx, nil)
		if err != nil {
			panic(err)
		}

		if diff := cmp.Diff(expected.sensorConfigurations, results, cmpopts.SortSlices(func(a, b Asset) bool {
			return a.Name < b.Name
		})); diff != "" {
			t.Errorf("Slices mismatch (-want +got):\n%s", diff)
		}
	}

	if expected.calibratedData != nil {
		results, err := cw.calibratedData.collection.Find(ctx, nil)
		if err != nil {
			panic(err)
		}

		if diff := cmp.Diff(expected.calibratedData, results, cmpopts.SortSlices(func(a, b Asset) bool {
			return a.Name < b.Name // Assuming there's an ID field for sorting
		})); diff != "" {
			t.Errorf("Slices mismatch (-want +got):\n%s", diff)
		}
	}

	if expected.assets != nil {
		results, err := cw.assets.collection.Find(ctx, nil)
		if err != nil {
			panic(err)
		}

		if diff := cmp.Diff(expected.assets, results, cmpopts.SortSlices(func(a, b Asset) bool {
			return a.Name < b.Name // Assuming there's an ID field for sorting
		})); diff != "" {
			t.Errorf("Slices mismatch (-want +got):\n%s", diff)
		}
	}
}

func clearMockCollections(dbWrap *MongoDatabaseImpl, cw *CollectionToData) {
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

// type DateExpectedToFind[T any] struct {
// 	collectionName DB_COLLECTIONS
// 	data           []T
// }

type CollectionToData struct {
	sensors              CollectionWrapper[Sensor]
	rawData              CollectionWrapper[RawData]
	sensorConfigurations CollectionWrapper[SensorConfiguration]
	calibratedData       CollectionWrapper[CalibratedData]
	assets               CollectionWrapper[Asset]
}

type DateExpectedToFind struct {
	sensors              []Sensor
	rawData              []RawData
	sensorConfigurations []SensorConfiguration
	calibratedData       []CalibratedData
	assets               []Asset
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
	MOCK_ASSET_ID := primitive.NewObjectID()

	defaultHeader := http.Header{
		"Content-Type": {"application/json"},
	}

	type args struct {
		uplinkMessage     UplinkMessage
		additionalHeaders http.Header
		server            *Server
		preData           CollectionToData
	}
	type expected struct {
		code     int
		message  map[string]string
		postData DateExpectedToFind
	}

	tests := []struct {
		name     string
		runTest  bool
		args     args
		expected expected
	}{
		{
			name:    "Valid 'S2120RawData' data",
			runTest: true,
			args: args{
				additionalHeaders: http.Header{
					"X-Downlink-Apikey": []string{"RANDOM_TEST_KEY"},
				},
				server: &Server{
					Envs: &environmentVariables{
						Ttn_webhhook_api: "RANDOM_TEST_KEY",
					},
					Sensors: &syncCacheImpl[string, Sensor]{
						cache: map[string]Sensor{
							MOCK_DEVICE_ID: {
								Id:    MOCK_SENSOR_ID,
								Model: S2120,
							},
						},
					},
				},
				uplinkMessage: createMockUplinkMessage(
					MOCK_DEVICE_ID,
					MOCK_RECEIVED_AT,
					map[string]interface{}{
						"err":     0,
						"payload": "",
						"valid":   true,
						"messages": []map[string]interface{}{
							{
								"measurementValue": 1.44,
								"measurementId":    "555",
								"type":             string(RainGauge),
							},
							{
								"measurementValue": 0.5,
								"measurementId":    "4104",
								"type":             string(WindDirectionSensor),
							},
						},
					},
				),
				preData: CollectionToData{
					sensorConfigurations: CollectionWrapper[SensorConfiguration]{
						data: []SensorConfiguration{
							{
								Sensor:  MOCK_SENSOR_ID,
								Asset:   MOCK_ASSET_ID,
								Applied: mockConvertTimeStringToMongoTime("2025-01-26T13:35:18.467+00:00"),
							},
						},
					},
					assets: CollectionWrapper[Asset]{
						data: []Asset{
							{
								Id: MOCK_ASSET_ID,
							},
						},
					},
				},
			},
			expected: expected{
				code: http.StatusOK,
				message: map[string]string{
					"message": "Webhook received successfully",
				},
				postData: DateExpectedToFind{
					rawData: []RawData{
						{
							Timestamp: mockConvertTimeStringToMongoTime(MOCK_RECEIVED_AT),
							Sensor:    &MOCK_SENSOR_ID,
							Valid:     true,
							Data: SensorData{
								S2120: &S2120RawData{
									Messages: []S2120RawDataMsg{
										&S2120RawDataMeasurement{
											MeasurementId:    "555",
											MeasurementValue: 1.44,
											Type:             RainGauge,
										},
										&S2120RawDataMeasurement{
											MeasurementId:    "4104",
											MeasurementValue: 0.5,
											Type:             WindDirectionSensor,
										},
									},
								},
							},
						},
					},
					calibratedData: []CalibratedData{
						{
							Timestamp: mockConvertTimeStringToMongoTime(MOCK_RECEIVED_AT),
							Sensor:    MOCK_SENSOR_ID,
							DataPoints: CalibratedDataPoints{
								RainfallHourly: &CalibratedDataType{
									Data:  1.44,
									Units: MM_PER_HOUR,
								},
								WindDirection: &CalibratedDataType{
									Data:  0.5,
									Units: DEGREE,
								},
							},
						},
					},
				},
			},
		},
		{
			name:    "Valid 'LDDS45RawData' data",
			runTest: true,
			args: args{
				additionalHeaders: http.Header{
					"X-Downlink-Apikey": []string{"RANDOM_TEST_KEY"},
				},
				server: &Server{
					Envs: &environmentVariables{
						Ttn_webhhook_api: "RANDOM_TEST_KEY",
					},
					Sensors: &syncCacheImpl[string, Sensor]{
						cache: map[string]Sensor{
							MOCK_DEVICE_ID: {
								Id:    MOCK_SENSOR_ID,
								Model: LDDS45,
							},
						},
					},
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
				preData: CollectionToData{
					sensorConfigurations: CollectionWrapper[SensorConfiguration]{
						data: []SensorConfiguration{
							{
								Sensor:  MOCK_SENSOR_ID,
								Asset:   MOCK_ASSET_ID,
								Applied: mockConvertTimeStringToMongoTime("2025-01-26T13:35:18.467+00:00"),
								Offset: &struct {
									Distance *struct {
										Distance float64 "bson:\"distance\""
										Units    UNITS   "bson:\"units\""
									} "bson:\"distance\""
								}{
									Distance: &struct {
										Distance float64 "bson:\"distance\""
										Units    UNITS   "bson:\"units\""
									}{
										Distance: 0,
										Units:    METRES,
									},
								},
							},
						},
					},
					assets: CollectionWrapper[Asset]{
						data: []Asset{
							{
								Id: MOCK_ASSET_ID,
								Metrics: &AssetMetrics{
									Volume: &AssetMetricsCylinderVolume{
										Radius: float64(5),
										Height: float64(2.2),
									},
								},
							},
						},
					},
				},
			},
			expected: expected{
				code: http.StatusOK,
				message: map[string]string{
					"message": "Webhook received successfully",
				},
				postData: DateExpectedToFind{
					rawData: []RawData{
						{
							Timestamp: mockConvertTimeStringToMongoTime(MOCK_RECEIVED_AT),
							Sensor:    &MOCK_SENSOR_ID,
							Valid:     true,
							Data: SensorData{
								LDDS45: &LDDS45RawData{
									Battery:      3.413,
									Distance:     "1404 mm",
									InterruptPin: 0,
									Temperature:  "0.00",
									SensorFlag:   1,
								},
							},
						},
					},
					calibratedData: []CalibratedData{
						{
							Timestamp: mockConvertTimeStringToMongoTime(MOCK_RECEIVED_AT),
							Sensor:    MOCK_SENSOR_ID,
							DataPoints: CalibratedDataPoints{
								Volume: &CalibratedDataType{
									Data:  float64(62517.69),
									Units: METRES_CUBE,
								},
							},
						},
					},
				},
			},
		},
		{
			name:    "No 'X-Downlink-Apikey' header provided",
			runTest: true,
			args: args{
				server: &Server{},
			},
			expected: expected{
				code: http.StatusBadRequest,
				message: map[string]string{
					"error": "Missing X-Downlink-Apikey header",
				},
			},
		},
		{
			name:    "Mismatch 'X-Downlink-Apikey' header provided",
			runTest: true,
			args: args{
				additionalHeaders: http.Header{
					"X-Downlink-Apikey": []string{"RANDOM_TEST_KEY"},
				},
				server: &Server{
					Envs: &environmentVariables{},
				},
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
			if !tt.runTest {
				t.Skipf("Skipping: %s\n", tt.name)
			}

			// TODO: mongo setup ? - no enforced schema or anything so limited scope here.
			// TODO: mongo collection/data tear down after the test is completed.

			tt.args.server.MongoDb = mongoDb

			defer clearMockCollections(mongoDb, &tt.args.preData)

			setupMockCollections(mongoDb, &tt.args.preData)
			insertMockData(&tt.args.preData)

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

			handleWebhook(mockCtx, tt.args.server)

			assert.Equal(t, tt.expected.code, w.Code)

			expectedJson, err := mapToJSONString(tt.expected.message)
			assert.ErrorIs(t, err, nil)
			assert.Equal(t, expectedJson, w.Body.String())

			// Check that the data is valid
			validateDataExistingsWithinMockDb(t, &tt.expected.postData, &tt.args.preData)
		})
	}
}
