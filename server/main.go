package main

import (
	"log"
	"mini-farm-tracker-server/core"
)

func main() {
	log.Println("Starting up...")

	// values for Mongo and TTN
	envs := core.ReadEnvs()

	mongoDb, mongoDeferFn := core.SetupMongo(envs)
	defer mongoDeferFn()

	r := core.SetupRouter(envs, mongoDb)

	log.Println("Server listening...")
	// port defaults 8080 but for clarify, declaring
	log.Fatal(r.Run(":8080"))
}
