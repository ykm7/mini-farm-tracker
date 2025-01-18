package main

import (
	"log"
)

func main() {
	log.Println("Starting up...")

	// values for Mongo and TTN
	envs := readEnvs()

	_ = setupMongo(envs)
	r := setupRouter(envs)

	log.Println("Server listening...")
	// port defaults 8080 but for clarify, declaring
	log.Fatal(r.Run(":8080"))
}
