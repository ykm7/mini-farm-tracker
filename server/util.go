package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type environmentVariables struct {
	ttn_webhhook_api string
	mongo_conn       string
}

func readEnvs() (envs environmentVariables) {
	// envs := environmentVariables{}

	if !isProduction() {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	envs.mongo_conn = os.Getenv("MONGO_CONN")
	envs.ttn_webhhook_api = os.Getenv("TTN_WEBHOOK_API")

	return
}

func isProduction() bool {
	return gin.Mode() == "release"
}
