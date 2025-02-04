package core

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mapToJSONString(m any) (string, error) {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func TestSetupRouter(t *testing.T) {
	db, deferFn := MockSetupMongo(context.TODO())
	defer deferFn()

	mongoDb := &MongoDatabaseImpl{Db: db}

	server := &Server{
		MongoDb: mongoDb,
		Sensors: NewSyncCache[string, Sensor](),
	}

	router := SetupRouter(server)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	expectedJson, err := mapToJSONString(map[string]string{"message": "pong"})
	assert.ErrorIs(t, err, nil)
	assert.Equal(t, expectedJson, w.Body.String())
}
