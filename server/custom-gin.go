package main

import "github.com/gin-gonic/gin"

const HEALTH_ENDPOINT = "/health"

func CustomLogger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{HEALTH_ENDPOINT},
	})
}
