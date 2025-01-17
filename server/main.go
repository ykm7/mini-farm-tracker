package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("hello")

	r := gin.Default()
	// r.Use(cors.Default())

	config := cors.DefaultConfig()

	// config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	// config.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	// config.AllowCredentials = true

	if gin.Mode() == "release" {
		config.AllowOrigins = []string{"https://mini-farm-tracker.io", "https://www.mini-farm-tracker.io"}
	} else {
		// vue development
		config.AllowOrigins = []string{"http://localhost:5173"}
	}

	r.Use(cors.New(config))
	// r.SetTrustedProxies([]string{"mini-farm-tracker.io"})
	// r.ForwardedByClientIP = true

	api := r.Group("/api")
	{
		api.GET("/data", dataFn)
	}

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// if gin.Mode() == "release" {
	// 	log.Println("Starting in release mode")
	// 	// log.Fatal(autotls.Run(r, "api.mini-farm-tracker.io"))
	// 	log.Fatal(autotls.Run(r, "mini-farm-tracker.io"))
	// } else {
	// 	log.Println("Starting in local dev mode")
	// 	log.Fatal(r.Run())
	// }

	log.Fatal(r.Run())
}
