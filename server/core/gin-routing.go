package core

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	compress "github.com/lf4096/gin-compress"
	"golang.org/x/sync/semaphore"
)

const HEALTH_ENDPOINT = "/health"
const SENSOR_ID_PARAM = "sensor_id"
const START_DATE = "start"
const END_DATE = "end"

func CustomLogger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{HEALTH_ENDPOINT},
	})
}

/*
*
Applying Chrome's Lighthouse best practices
https://developer.chrome.com/docs/lighthouse/best-practices/has-hsts?utm_source=lighthouse&utm_medium=devtools

* The max-age directive specifies the amount of time the user's browser is forced to visit a website only using TLS (in seconds). After that time, it will be possible to reach the site using plain HTTP again if there is no HSTS header provided by the website (or temporary redirects from HTTP to HTTPS are in place).

* Setting the includeSubDomains directive will enforce the header on any subdomains of the page URL sending the header initially. For example, having an HSTS header sent by google.com which includes the includeSubDomains directive would also enforce the HSTS header on mail.google.com.

* Setting the preload directive and submitting the domain to the HSTS preload service will compile the domain into browser binaries that use the preloaded HSTS list (not just Google Chrome).
*/
func HSTSHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		c.Next()
	}
}

func ConcurrencyLimiter(maxConcurrent int64) gin.HandlerFunc {
	sem := semaphore.NewWeighted(maxConcurrent)
	return func(c *gin.Context) {
		if err := sem.Acquire(context.Background(), 1); err != nil {

			/**
			Normally what would be done, however given the nature of the project would rather collect
			this as a metric to make futher modifications.
			*/
			// c.AbortWithStatus(503) // Service Unavailable
			// return

			log.Printf("Gin middleware concurrency limiter limit hit with a value of %d\n", maxConcurrent)
			c.Next()
		}
		defer sem.Release(1)
		c.Next()
	}
}

func SetupRouter(server *Server) *gin.Engine {
	r := gin.New()

	r.Use(CustomLogger())
	r.Use(HSTSHandler())
	r.Use(ConcurrencyLimiter(100))
	r.Use(compress.Compress())
	r.Use(gin.Recovery())

	config := cors.DefaultConfig()
	config.ExposeHeaders = []string{DATA_API_LIMIT_HEADER}

	if isProduction() {
		config.AllowOrigins = []string{"https://mini-farm-tracker.io", "https://www.mini-farm-tracker.io"}
	} else {
		// vue development
		config.AllowOrigins = []string{"http://localhost:5173"}
	}

	r.Use(cors.New(config))
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
			}
		}

		assetsApi := api.Group("/assets")
		{
			assetsApi.GET("", func(ctx *gin.Context) {
				handleAssetsWithoutId(ctx, server)
			})
		}
	}

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/webhook", func(c *gin.Context) {
		handleWebhook(c, server)
	})

	log.Printf("Endpoint: %s not logged\n", HEALTH_ENDPOINT)
	r.GET(HEALTH_ENDPOINT, func(c *gin.Context) {

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

	return r
}
