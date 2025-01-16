package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("hello")

	r := gin.Default()

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

	log.Println("Starting server on port 3000...")

	// log.Fatal(autotls.Run(r, "mini-farm-tracker.io"))
	log.Fatal(r.Run(":3000"))

	log.Println("Server has started successfully.")
}
