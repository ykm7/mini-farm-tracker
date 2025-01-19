package core

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

func ReadEnvs() *environmentVariables {
	if !isProduction() {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	return &environmentVariables{
		ttn_webhhook_api: os.Getenv("TTN_WEBHOOK_API"),
		mongo_conn:       os.Getenv("MONGO_CONN"),
	}
}

/*
Gins mode is set to "release" if the
environment variable GIN_MODE == "release"
*/
func isProduction() bool {
	return gin.Mode() == "release"
}
