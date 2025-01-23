package core

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const TEST_DATABASE_NAME string = "test_db"

/*
https://golang.testcontainers.org/modules/mongodb/#connectionstring
*/
func MockSetupMongo(ctx context.Context) (db *mongo.Database, deferFn func()) {
	mongoDBContainer, err := mongodb.Run(ctx, "mongo:8")

	deferFn = func() {
		if err := testcontainers.TerminateContainer(mongoDBContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}

	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	endpoint, err := mongoDBContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(endpoint))
	if err != nil {
		log.Printf("failed to connect to MongoDB: %s", err)
		return
	}

	db = mongoClient.Database(TEST_DATABASE_NAME)

	return
}

/*
Credit to: https://medium.com/canopas/golang-unit-tests-with-test-gin-context-80e1ac04adcd

Using this structure over initialising the "entire" gin server is definitely more performative however would bypass any side effects
which may exist in the default setup pathway.
*/
func MockGinContext(w *httptest.ResponseRecorder) *gin.Context {
	gin.SetMode(gin.TestMode)

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = &http.Request{
		Header: make(http.Header),
		URL:    &url.URL{},
	}

	return ctx
}

func MockContextAdd(c *gin.Context, headers http.Header) {
	if headers != nil {
		for key, values := range headers {
			for _, value := range values {
				c.Request.Header.Add(key, value)
			}
		}
	}
}

func MockJsonGet(c *gin.Context, params gin.Params, u url.Values) {
	c.Request.Method = "GET"
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	// set path params
	c.Params = params

	// set query params
	c.Request.URL.RawQuery = u.Encode()
}

func MockJsonPost(c *gin.Context, content interface{}) {
	c.Request.Method = "POST"
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	jsonbytes, err := json.Marshal(content)
	if err != nil {
		panic(err)
	}

	// the request body must be an io.ReadCloser
	// the bytes buffer though doesn't implement io.Closer,
	// so you wrap it in a no-op closer
	c.Request.Body = io.NopCloser(bytes.NewBuffer(jsonbytes))
}

// func MockJsonGet(c *gin.Context, method string, headers http.Header, params gin.Params, u url.Values) {
// 	c.Request.Method = method

// 	// Allow for addition headers - needed for dynamic api key header
// 	for key, values := range headers {
// 		for _, value := range values {
// 			c.Request.Header.Add(key, value)
// 		}
// 	}

// 	c.Request.Header.Set("Content-Type", "application/json")
// 	c.Set("user_id", 1)

// 	// set path params
// 	c.Params = params

// 	// set query params
// 	c.Request.URL.RawQuery = u.Encode()
// }
