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

	exitChan := make(chan struct{})
	ctx := context.Background()
	innerCtx, innerCtxCancel := core.ContextWithQuitChannel(ctx, exitChan)
	defer innerCtxCancel()

	// values for Mongo and TTN
	envs := core.ReadEnvs()

	database, mongoDeferFn := core.SetupMongo(envs)
	defer mongoDeferFn()

	redis, redisDeferFn := core.GetRedisClient(envs)
	defer redisDeferFn()

	mongoDb := &core.MongoDatabaseImpl{Db: database}

	server := &core.Server{
		Envs:        envs,
		MongoDb:     mongoDb,
		Redis:       redis,
		Sensors:     core.NewSyncCache[string, core.Sensor](),
		Tasks:       make(chan core.TaskJob),
		ExitContext: innerCtx,
		ExitChan:    exitChan,
	}

	core.ListenToSensors(server)

	core.SetupPeriodicTasks(server)
	go core.SetupTaskHandler(server)

	r := core.SetupRouter(server)

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

	select {
	case <-quit:
		log.Println("Received OS signal, shutting down...")
	case <-server.ExitChan:
		log.Println("Change stream on the 'sensors' collection exited, shutting down...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	select {
	case <-ctx.Done():
		log.Println("timeout of 2 seconds.")
	}
	log.Println("Server exiting")
}
