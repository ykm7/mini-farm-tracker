package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("hello")
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// if err := r.Run(":3000"); err != nil {
	// 	log.Fatalf("Failed to run server: %v", err)
	// }

	log.Fatal(autotls.Run(r, "mini-farm-tracker.io"))
}
