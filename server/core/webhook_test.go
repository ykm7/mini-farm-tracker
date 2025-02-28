package core

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
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

		if diff := cmp.Diff(expected.sensors, results, cmpopts.SortSlices(func(a, b Sensor) bool {
			return a.Id < b.Id
		})); diff != "" {
			t.Errorf("Slices mismatch (-want +got):\n%s", diff)
		}
	}

	if expected.rawData != nil {
		results, err := cw.rawData.collection.Find(ctx, nil)
		if err != nil {
			panic(err)
		}

		// TODO: Ideally I'd like order of ALL the nested arrays to be ignored.
		rawDataCmp := cmpopts.SortSlices(func(a, b RawData) bool {
			return a.Timestamp < b.Timestamp
		})

		LDDS45Cmp := cmpopts.SortSlices(func(a, b LDDS45RawData) bool {
			return a.Distance < b.Distance
		})

		// S2120Cmp := cmpopts.SortSlices(func(a, b S2120RawData) bool {
		// 	return a.Messages. < b.Messages.
		// })

		S2120RawDataMsgCmp := cmpopts.SortSlices(func(a, b S2120RawDataMsg) bool {
			aType := reflect.TypeOf(a)
			bType := reflect.TypeOf(b)

			if aType != bType {
				return false
			}

			switch aCast := a.(type) {
			case *S2120RawDataMeasurement:
				aId := aCast.MeasurementId

				bCast := b.(*S2120RawDataMeasurement)
				bId := bCast.MeasurementId
				return aId < bId
			}

			return false
		})

		S2120RawDataMeasurementCmp := cmpopts.SortSlices(func(a, b *S2120RawDataMeasurement) bool {
			if a == nil || b == nil {
				return false
			}

			return a.MeasurementId < b.MeasurementId
		})

		if diff := cmp.Diff(expected.rawData, results, rawDataCmp, LDDS45Cmp, S2120RawDataMsgCmp, S2120RawDataMeasurementCmp); diff != "" {
			t.Errorf("Slices mismatch (-want +got):\n%s", diff)
		}
	}

	if expected.sensorConfigurations != nil {
		results, err := cw.sensorConfigurations.collection.Find(ctx, nil)
		if err != nil {
			panic(err)
		}

		if diff := cmp.Diff(expected.sensorConfigurations, results, cmpopts.SortSlices(func(a, b SensorConfiguration) bool {
			return a.Id.String() < b.Id.String()
		})); diff != "" {
			t.Errorf("Slices mismatch (-want +got):\n%s", diff)
		}
	}

	if expected.calibratedData != nil {
		results, err := cw.calibratedData.collection.Find(ctx, nil)
		if err != nil {
			panic(err)
		}

		if diff := cmp.Diff(expected.calibratedData, results, cmpopts.SortSlices(func(a, b CalibratedData) bool {
			return a.Timestamp < b.Timestamp // Assuming there's an ID field for sorting
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
	db, deferFn := MockSetupMongo(context.TODO())
	mongoDb := &MongoDatabaseImpl{Db: db}
	defer deferFn()

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
							MOCK_SENSOR_ID: {
								Id:    MOCK_SENSOR_ID,
								Model: S2120,
							},
						},
					},
				},
				uplinkMessage: createMockUplinkMessage(
					MOCK_SENSOR_ID,
					MOCK_RECEIVED_AT,
					map[string]interface{}{
						"err":     0,
						"payload": "",
						"valid":   true,
						// From: 4A00F6320000FFA33500084B01090000000027414C0019000009EC
						"messages": []map[string]interface{}{
							{
								"measurementId":    "4098",
								"measurementValue": 50,
								"type":             "Air Humidity",
							},
							{
								"measurementId":    "4097",
								"measurementValue": 24.6,
								"type":             "Air Temperature",
							},
							{
								"measurementId":    "4099",
								"measurementValue": 65443,
								"type":             "Light Intensity",
							},
							{
								"measurementId":    "4190",
								"measurementValue": 5.3,
								"type":             "UV Index",
							},
							{
								"measurementId":    "4105",
								"measurementValue": 0.8,
								"type":             "Wind Speed",
							},
							{
								"measurementId":    "4113",
								"measurementValue": 2.30898,
								"type":             "Rain Gauge",
							},
							{
								"measurementId":    "4104",
								"measurementValue": 265,
								"type":             "Wind Direction Sensor",
							},
							{
								"measurementId":    "4101",
								"measurementValue": 100490,
								"type":             "Barometric Pressure",
							},
							{
								"measurementId":    "4191",
								"measurementValue": 2.5,
								"type":             "Peak Wind Gust",
							},
							// We don't care about this
							{
								"measurementId":    "4213",
								"measurementValue": 999,
								"type":             "Rain Accumulation",
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
											MeasurementId:    "4097",
											MeasurementValue: float64(24.6),
											Type:             AirTemperature,
										},
										&S2120RawDataMeasurement{
											MeasurementId:    "4098",
											MeasurementValue: float64(50),
											Type:             AirHumidity,
										},
										&S2120RawDataMeasurement{
											MeasurementId:    "4099",
											MeasurementValue: float64(65443),
											Type:             LightIntensity,
										},
										&S2120RawDataMeasurement{
											MeasurementId:    "4190",
											MeasurementValue: float64(5.3),
											Type:             UVIndex,
										},
										&S2120RawDataMeasurement{
											MeasurementId:    "4105",
											MeasurementValue: float64(0.8),
											Type:             WindSpeed,
										},
										&S2120RawDataMeasurement{
											MeasurementId:    "4113",
											MeasurementValue: float64(2.30898),
											Type:             RainGauge,
										},
										&S2120RawDataMeasurement{
											MeasurementId:    "4104",
											MeasurementValue: float64(265),
											Type:             WindDirectionSensor,
										},
										&S2120RawDataMeasurement{
											MeasurementId:    "4101",
											MeasurementValue: float64(100490),
											Type:             BarometricPressure,
										},
										&S2120RawDataMeasurement{
											MeasurementId:    "4191",
											MeasurementValue: float64(2.5),
											Type:             PeakWindGust,
										},
										&S2120RawDataMeasurement{
											MeasurementId:    "4213",
											MeasurementValue: float64(999),
											Type:             "Rain Accumulation",
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
								AirTemperature: &CalibratedDataType{
									Data:  24.6,
									Units: DEGREE_C,
								},
								AirHumidity: &CalibratedDataType{
									Data:  50,
									Units: AIR_HUMIDITY,
								},
								LightIntensity: &CalibratedDataType{
									Data:  65443,
									Units: LUX,
								},
								UVIndex: &CalibratedDataType{
									Data:  5.3,
									Units: UV_INDEX,
								},
								WindSpeed: &CalibratedDataType{
									Data:  0.8,
									Units: M_PER_SEC,
								},
								RainfallHourly: &CalibratedDataType{
									Data:  2.30898,
									Units: MM_PER_HOUR,
								},
								WindDirection: &CalibratedDataType{
									Data:  265,
									Units: DEGREE,
								},
								BarometricPressure: &CalibratedDataType{
									Data:  100490,
									Units: PRESSURE,
								},
								PeakWindGust: &CalibratedDataType{
									Data:  2.5,
									Units: M_PER_SEC,
								},
								RainAccumulation: &CalibratedDataType{
									Data:  0.38483,
									Units: MM_METRE,
								},
							},
						},
					},
				},
			},
		},
		{
			name:    "Invalid 'S2120RawData' data - no messages",
			runTest: false,
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
							MOCK_SENSOR_ID: {
								Id:    MOCK_SENSOR_ID,
								Model: S2120,
							},
						},
					},
				},
				uplinkMessage: createMockUplinkMessage(
					MOCK_SENSOR_ID,
					MOCK_RECEIVED_AT,
					map[string]interface{}{
						"err":      0,
						"payload":  "",
						"valid":    true,
						"messages": []map[string]interface{}{},
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
									Messages: []S2120RawDataMsg{},
								},
							},
						},
					},
					calibratedData: []CalibratedData{
						{
							Timestamp:  mockConvertTimeStringToMongoTime(MOCK_RECEIVED_AT),
							Sensor:     MOCK_SENSOR_ID,
							DataPoints: CalibratedDataPoints{},
						},
					},
				},
			},
		},
		{
			name:    "Invalid 'S2120RawData' data - no data at all",
			runTest: false,
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
							MOCK_SENSOR_ID: {
								Id:    MOCK_SENSOR_ID,
								Model: S2120,
							},
						},
					},
				},
				uplinkMessage: createMockUplinkMessage(
					MOCK_SENSOR_ID,
					MOCK_RECEIVED_AT,
					map[string]interface{}{},
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
				code: http.StatusBadRequest,
				message: map[string]string{
					"status": "Error casting the decoded json: [110 117 108 108] (as string: null) to expected data type for: S2120",
				},
				postData: DateExpectedToFind{},
			},
		},
		{
			name:    "Valid 'LDDS45RawData' data",
			runTest: false,
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
							MOCK_SENSOR_ID: {
								Id:    MOCK_SENSOR_ID,
								Model: LDDS45,
							},
						},
					},
				},
				uplinkMessage: createMockUplinkMessage(
					MOCK_SENSOR_ID,
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
			runTest: false,
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
			runTest: false,
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

			// TODO: Raw data

			// calibrated data
			// w = httptest.NewRecorder()
			// mockCtx = MockGinContext(w)

			// // mockCtx.Request.URL.RawPath = "api/sensors/2cf7f1c0613006fe/data/raw_data"

			// MockJsonGet(mockCtx, gin.Params{{Key: SENSOR_ID_PARAM, Value: MOCK_SENSOR_ID}}, url.Values{})

			// getCalibratedDataWithSensorId(mockCtx, tt.args.server)

			// assert.Equal(t, tt.expected.code, w.Code)
			// log.Println(w.Body.String())
		})
	}
}
