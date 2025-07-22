package core

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
)

const (
	HEALTH_ENDPOINT  = "/health"
	METRICS_ENDPOINT = "/metrics"
)

const (
	SENSOR_ID_PARAM = "sensor_id"
	START_DATE      = "start"
	END_DATE        = "end"
)

const (
	NO_ROUTE_LIMIT                               = rate.Limit(1)
	NO_ROUTE_BURST                               = 3
	NO_ROUTE_CONCURRENCY_LIMIT                   = 5
	NO_ROUTE_CONCURRENCY_SEMAPHORE               = "semaphore"
	NO_ROUTE_CONCURRENCY_SEMAPHORE_ACQUIRE_COUNT = 1
)

/*
*
Extracted to allow for easy extension on testing
*/
func noRoute() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		rateLimiter(NO_ROUTE_LIMIT, NO_ROUTE_BURST),
		ConcurrencyLimiter(NO_ROUTE_CONCURRENCY_LIMIT, true),
	}
}

func CustomLogger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{HEALTH_ENDPOINT, METRICS_ENDPOINT},
	})
}

func rateLimiter(r rate.Limit, burst int) gin.HandlerFunc {

	limiters := make(map[string]*rate.Limiter)
	var mu sync.Mutex

	return func(ctx *gin.Context) {

		ip := ctx.ClientIP()
		mu.Lock()

		limiter, exists := limiters[ip]
		if !exists {
			limiter = rate.NewLimiter(r, burst)
			limiters[ip] = limiter
		}

		mu.Unlock()

		if !limiter.Allow() {
			ctx.JSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

/*
Normally what would be done, however given the nature of the project would rather collect
this as a metric to make futher modifications.
*/
func ConcurrencyLimiter(maxConcurrent int64, causeError bool) gin.HandlerFunc {
	sem := semaphore.NewWeighted(maxConcurrent)
	return func(ctx *gin.Context) {
		if !sem.TryAcquire(NO_ROUTE_CONCURRENCY_SEMAPHORE_ACQUIRE_COUNT) {
			if causeError {
				log.Printf("Concurrency limit reached, returning 503")
				ctx.AbortWithStatus(http.StatusServiceUnavailable)
			} else {
				log.Printf("Concurrency limit reached, but allowing request")
				ctx.Next()
			}
			return
		}

		// log.Printf("Request acquired semaphore")
		ctx.Set(NO_ROUTE_CONCURRENCY_SEMAPHORE, sem)
		ctx.Next()
	}
}

func ReleaseSemaphore() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if sem, exists := ctx.Get(NO_ROUTE_CONCURRENCY_SEMAPHORE); exists {
				if s, ok := sem.(*semaphore.Weighted); ok {
					s.Release(NO_ROUTE_CONCURRENCY_SEMAPHORE_ACQUIRE_COUNT)
					// log.Printf("Semaphore released at the end of request")
				}
			}
		}()
		ctx.Next()
	}
}

func SetupRouter(server *Server) *gin.Engine {
	r := gin.New()

	r.Use(CustomLogger())
	/**
	Leaving these currently for historic purpose and again investigate pros/cons of having headers within code vs hosted.
	*/
	// r.Use(HSTSHandler())
	// r.Use(CSPHandler())
	// r.Use(COOPHandler())
	r.Use(gin.Recovery())
	// The prometheus endpoint already applies compression and therefore needs to be excluded to prevent double-compression
	// Testing:
	// Requesting /metrics with default headers (Accept-Encoding: gzip, deflate, br)
	// ----------------------------------------------------------------------------------
	// Content-Encoding - gzip
	// Content-Type 	- text/plain; version=0.0.4; charset=utf-8; escaping=underscores
	// size				- 2.02kb
	// ----------------------------------------------------------------------------------
	// Requesting /metrics with headers to force uncompression (Accept-Encoding: identify)
	// ----------------------------------------------------------------------------------
	// Content-Type 	- text/plain; version=0.0.4; charset=utf-8; escaping=underscores
	// size				- 8.02kb
	r.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths([]string{METRICS_ENDPOINT})))
	r.Use(ReleaseSemaphore())

	config := cors.DefaultConfig()
	config.ExposeHeaders = []string{DATA_API_LIMIT_HEADER}

	if isProduction() {
		config.AllowOrigins = []string{"https://mini-farm-tracker.io", "https://www.mini-farm-tracker.io"}
	} else {
		// vue development
		config.AllowOrigins = []string{"http://localhost:5173"}
	}

	r.Use(cors.New(config))

	/**
	This section was deemed useful after noticing various "attacks" on the server.

	1. Various attempts to grab the project
		Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      51.638µs | 206.221.176.253 | GET      "/backup.rar"
		Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      47.694µs | 206.221.176.253 | GET      "/site.zip"
		Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      44.674µs | 206.221.176.253 | GET      "/backup.zip"
		Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      23.071µs | 206.221.176.253 | GET      "/api_mini-farm-tracker_io.rar"
		Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      28.328µs | 206.221.176.253 | GET      "/website.zip"
		Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      18.078µs | 206.221.176.253 | GET      "/api_mini-farm-tracker_io.zip"
		Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      31.558µs | 206.221.176.253 | GET      "/site.rar"
		Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      40.026µs | 206.221.176.253 | GET      "/api.mini-farm-tracker.io.zip"
		Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      16.631µs | 206.221.176.253 | GET      "/website.rar"
		Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      54.945µs | 206.221.176.253 | GET      "/apimini-farm-trackerio.rar"
		Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      51.673µs | 206.221.176.253 | GET      "/apimini-farm-trackerio.zip"
		Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      38.742µs | 206.221.176.253 | GET      "/api.mini-farm-tracker.io.rar"

	2. Various attempts to query for version control content
		Feb 20 05:46:17 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 13:46:15 | 404 |      61.282µs |    148.66.1.242 | GET      "/.git/HEAD"
		Feb 20 05:46:17 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 13:46:16 | 404 |      62.041µs |    148.66.1.242 | GET      "/.git/config"
		Feb 20 05:46:18 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 13:46:17 | 404 |      56.961µs |    148.66.1.242 | GET      "/.svn/entries"
		Feb 20 05:46:18 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 13:46:17 | 404 |      47.556µs |    148.66.1.242 | GET      "/.svn/wc.db"
		Feb 20 07:48:07 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 15:48:05 | 404 |      65.983µs |    148.66.1.242 | GET      "/.git/HEAD"
		Feb 20 07:48:08 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 15:48:06 | 404 |      55.219µs |    148.66.1.242 | GET      "/.git/config"
		Feb 20 07:48:08 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 15:48:06 | 404 |      74.434µs |    148.66.1.242 | GET      "/.svn/entries"
		Feb 20 07:48:08 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 15:48:07 | 404 |      65.825µs |    148.66.1.242 | GET      "/.svn/wc.db"
		Feb 20 09:00:48 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 17:00:46 | 404 |      63.688µs |    148.66.1.242 | GET      "/.git/HEAD"
		Feb 20 09:00:48 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 17:00:47 | 404 |       43.38µs |    148.66.1.242 | GET      "/.svn/entries"
		Feb 20 09:00:48 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 17:00:46 | 404 |      51.889µs |    148.66.1.242 | GET      "/.git/config"
		Feb 20 09:00:48 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 17:00:47 | 404 |      50.416µs |    148.66.1.242 | GET      "/.svn/wc.db"

	Solution:
	1. Rate limit per IP addresses
	2. Concurrency limitation to all routes which are not available.
	3. Both of these values are very aggressive to limit useful information which is able to be obtained

	*/
	r.NoRoute(noRoute()...)
	concurrencyForOtherRoutes := ConcurrencyLimiter(100, false)

	api := r.Group("/api")
	{
		sensorApi := api.Group("/sensors")
		{
			sensorApi.GET("", func(c *gin.Context) {
				handleWithoutSensorID(c, server)
			})
			sensorApi.GET(fmt.Sprintf(":%s", SENSOR_ID_PARAM), handleWithSensorID)

			sensorDataApi := sensorApi.Group(fmt.Sprintf(":%s/data", SENSOR_ID_PARAM))
			{
				sensorDataApi.GET("/raw_data", func(c *gin.Context) {
					getRawDataWithSensorId(c, server)
				})
				sensorDataApi.GET("/calibrated_data", func(ctx *gin.Context) {
					getCalibratedDataWithSensorId(ctx, server)
				})
				sensorDataApi.GET("/aggregated_data", func(ctx *gin.Context) {
					getAggregationData(ctx, server)
				})
			}
		}

		assetsApi := api.Group("/assets")
		{
			assetsApi.GET("", func(ctx *gin.Context) {
				handleAssetsWithoutId(ctx, server)
			})
		}
	}
	api.Use(concurrencyForOtherRoutes)

	r.GET("/ping", concurrencyForOtherRoutes, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/webhook", concurrencyForOtherRoutes, func(c *gin.Context) {
		handleWebhook(c, server)
	})

	log.Printf("Endpoint: %s not logged\n", HEALTH_ENDPOINT)
	r.GET(HEALTH_ENDPOINT, concurrencyForOtherRoutes, func(c *gin.Context) {

		var wg sync.WaitGroup
		results := make(chan error, 2)
		wg.Add(2)

		go func() {
			defer wg.Done()

			err := PingMongo(server.MongoDb)
			if err != nil {
				err = fmt.Errorf("error in mongo ping %w", err)
			}
			results <- err
		}()

		go func() {
			defer wg.Done()

			err := PingRedis(server.Redis)
			if err != nil {
				err = fmt.Errorf("error in redis ping %w", err)
			}

			results <- err
		}()

		wg.Wait()
		close(results)

		success := true
		for result := range results {
			if result != nil {
				log.Printf("Error/s found:\n%v\n", result)
				success = false
			}
		}

		if success {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "failed",
			})
		}
	})

	log.Printf("Endpoint: %s not logged\n", METRICS_ENDPOINT)
	promHandler := server.Metrics.HandlerWithRedisUpdate()
	r.GET(METRICS_ENDPOINT, gin.BasicAuth(gin.Accounts{
		server.Envs.Metrics_username: server.Envs.Metrics_password,
	}), gin.WrapH(promHandler))

	return r
}
