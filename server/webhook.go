package main

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleWebhook(c *gin.Context, envs *environmentVariables) {
	log.Println("Webhook recieved.")

	apiKey := c.GetHeader("X-Downlink-Apikey")
	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing X-Downlink-Apikey header"})
		return
	}

	if apiKey != envs.ttn_webhhook_api {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook env is invalid"})
		return
	}

	var buf bytes.Buffer
	_, err := io.Copy(&buf, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	asString := buf.String()
	log.Printf("As string %s\n", asString)

	// TODO: Verify API Sign

	// TODO: Check its actually a device I care about

	// TODO: Store data point within Mongo

	// body, err := io.Copy(c.Request.Body)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
	// 	return
	// }

	// TODO: Verify the webhook signature if The Things Stack provides one

	// Process the webhook payload
	// TODO: Implement your logic to handle the webhook data

	// Respond with a success status
	c.JSON(http.StatusOK, gin.H{"message": "Webhook received successfully"})
}
