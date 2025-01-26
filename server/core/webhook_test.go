package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Test_handleWebhook(t *testing.T) {

	// initEnvironmentVariables := &environmentVariables{}
	// TODO: if the cache was queried multiple times within the function possible to open
	// it up to race conditions. Worth being aware of when adding the configuration stuff in the future
	// initsensorCache := map[string]Sensor{}

	db, deferFn := MockSetupMongo(context.TODO())
	mongoDb := &MongoDatabaseImpl{Db: db}

	defer deferFn()

	defaultHeader := http.Header{
		"Content-Type": {"application/json"},
	}

	type args struct {
		uplinkMessage     UplinkMessage
		additionalHeaders http.Header
		envs              *environmentVariables
		sensorCache       map[string]Sensor
	}
	type expected struct {
		code    int
		message map[string]string
	}

	tests := []struct {
		name     string
		args     args
		expected expected
	}{
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
		})
	}
}

func Test_storeLDDS45CalibratedData(t *testing.T) {
	mockCollection := &MockMongoCollection[any]{
		FindOneFn: func(ctx context.Context, filter interface{}, result *any) error {
			*result = SensorConfiguration{
				Sensor: "sensor id",
			}
			return nil
		},
	}

	mockDb := NewMockMongoDatabase()
	mockDb.SetCollection(string(SENSOR_CONFIGURATIONS_COLLECTION), mockCollection)

	// In your test function
	// typedCollection := getTypedCollection[YourType](mockDb, "your_collection_name")

	type args struct {
		ctx            context.Context
		mongoDb        MongoDatabase
		sensorId       string
		data           *LDDS45RawData
		receivedAtTime primitive.DateTime
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{
				mongoDb: mockDb,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := storeLDDS45CalibratedData(tt.args.ctx, tt.args.mongoDb, tt.args.sensorId, tt.args.data, tt.args.receivedAtTime); (err != nil) != tt.wantErr {
				t.Errorf("storeLDDS45CalibratedData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
