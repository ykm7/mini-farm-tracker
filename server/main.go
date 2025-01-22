package main

import (
	"context"
	"log"
	"mini-farm-tracker-server/core"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Println("Starting up...")

	// values for Mongo and TTN
	envs := core.ReadEnvs()

	mongoDb, mongoDeferFn := core.SetupMongo(envs)
	defer mongoDeferFn()

	// The idea is to keep a cache of the sensor information to prevent constant polling.
	// Also as I haven't used it directly myself.
	sensorCache := map[string]core.Sensor{}
	core.ListenToSensors(context.Background(), mongoDb, sensorCache)

	r := core.SetupRouter(envs, mongoDb, sensorCache)

	srv := &http.Server{
		// port defaults 8080 but for clarify, declaring
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		// cleanup shutdown - https://gin-gonic.com/docs/examples/graceful-restart-or-stop/
		log.Println("Server listening...")
		if err := r.Run(srv.Addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	select {
	case <-ctx.Done():
		log.Println("timeout of 5 seconds.")
	}
	log.Println("Server exiting")
}
