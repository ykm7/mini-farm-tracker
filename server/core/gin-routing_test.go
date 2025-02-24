package core

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
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
	server := &Server{}

	router := SetupRouter(server)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	expectedJson, err := mapToJSONString(map[string]string{"message": "pong"})
	assert.ErrorIs(t, err, nil)
	assert.Equal(t, expectedJson, w.Body.String())
}

func TestNoRouteSingleRequest(t *testing.T) {
	server := &Server{}

	router := SetupRouter(server)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping_random", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestNoRouteTriggerRateLimiter(t *testing.T) {
	server := &Server{}

	router := SetupRouter(server)

	req, _ := http.NewRequest("GET", "/ping_random", nil)
	for range NO_ROUTE_BURST {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

/*
*
Overall this was more to test functionality of the concurrency rather than be all that useful.
By default the NoRoute finishes so quickly that it cannot be used to without arbitrarily adding sleep
*/
func TestNoRouteTriggerConcurrency(t *testing.T) {
	server := &Server{}

	router := SetupRouter(server)

	// Idea where was that the semaphore within the `ConcurrencyLimiter` might not have time to setup the sem.
	tempSlowNoRouteFn := func(c *gin.Context) {
		time.Sleep(1 * time.Second)
		c.Status(http.StatusNotFound)
	}

	noRoutes := []gin.HandlerFunc{}
	noRoutes = append(noRoutes, noRoute()...)
	noRoutes = append(noRoutes, tempSlowNoRouteFn)

	router.NoRoute(noRoutes...)

	additionalRequests := 7

	returnStatus := make(chan int, NO_ROUTE_CONCURRENCY_LIMIT+additionalRequests)

	var wg sync.WaitGroup
	for i := range NO_ROUTE_CONCURRENCY_LIMIT + additionalRequests {
		wg.Add(1)
		go func(v int) {
			defer wg.Done()

			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/ping_random", nil)
			req.RemoteAddr = fmt.Sprintf("192.168.1.%d:1234", v+1)

			router.ServeHTTP(w, req)

			returnStatus <- w.Code
		}(i)
	}

	wg.Wait()
	close(returnStatus)

	statuses := make(map[int]int)

	for status := range returnStatus {
		statuses[status] += 1
	}

	/**
	At this point should have:
	* (NO_ROUTE_CONCURRENCY_LIMIT) number of 404 return codes.
	* (additionalRequests) number of 503
	*/
	assert.Equal(t, statuses[http.StatusNotFound], NO_ROUTE_CONCURRENCY_LIMIT)
	assert.Equal(t, statuses[http.StatusServiceUnavailable], additionalRequests)
}
