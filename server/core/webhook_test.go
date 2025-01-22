package core

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_handleWebhook(t *testing.T) {

	defaultHeader := http.Header{
		"Content-Type": {"application/json"},
	}

	type args struct {
		uplinkMessage UplinkMessage
		testHeaders   http.Header
		envs          *environmentVariables
		sensorCache   map[string]Sensor
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
			args: args{
				testHeaders:   http.Header{},
				uplinkMessage: UplinkMessage{},
				sensorCache:   map[string]Sensor{},
				envs:          &environmentVariables{},
			},
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
				testHeaders: http.Header{
					"X-Downlink-Apikey": []string{"RANDOM_TEST_KEY"},
				},
				uplinkMessage: UplinkMessage{},
				sensorCache:   map[string]Sensor{},
				envs:          &environmentVariables{},
			},
			expected: expected{
				code: http.StatusBadRequest,
				message: map[string]string{
					"error": "Webhook env is invalid",
				},
			},
		},
		// {
		// 	plenty more
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO - Move to a more test suite setup. Shouldn't have to spin up mongo and the gin server each time.
			db, deferFn := MockSetupMongo(context.TODO())
			defer deferFn()

			router := SetupRouter(tt.args.envs, db, tt.args.sensorCache)
			w := httptest.NewRecorder()
			// END TODO

			// Allow for addition headers - needed for dynamic api key header
			for key, values := range defaultHeader {
				for _, value := range values {
					tt.args.testHeaders.Add(key, value)
				}
			}

			jsonData, err := json.Marshal(tt.args.uplinkMessage)
			if err != nil {
				t.Fatalf("Failed to marshal UplinkMessage: %v", err)
			}

			req, _ := http.NewRequest("POST", "/webhook", bytes.NewBuffer(jsonData))

			req.Header = tt.args.testHeaders.Clone()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expected.code, w.Code)

			expectedJson, err := mapToJSONString(tt.expected.message)
			assert.ErrorIs(t, err, nil)
			assert.Equal(t, expectedJson, w.Body.String())
		})
	}
}
