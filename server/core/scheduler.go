package core

import (
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

// TODO: Would be an environment variable
const LOCATION = "Australia/Perth"

func SetupPeriodicTasks(server *Server) {
	loc, err := time.LoadLocation("Australia/Perth")
	if err != nil {
		log.Fatalf("Could not load timezone: %v", err)
	}

	c := cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger)), cron.WithLocation(loc))

	// These similar need to add to the current pool - push tasks to channel, group via the debouncer.
	c.AddFunc("@hourly", func() {
		fmt.Println("Every hour on the half hour")

		// tasks := []TaskMongoAggregation{}

		// for _, t := range tasks {
		// 	server.Tasks <- &t
		// }
	})
	c.AddFunc("@weekly ", func() {
		fmt.Println("Every hour on the half hour")

		hourlyPipeline := createAggregationPipeline("rainfallHourly", "hourly", "%Y-%m-%d-%H")

		tasks := []TaskMongoAggregation[CalibratedData]{
			TaskMongoAggregation[CalibratedData]{
				mongoCollection: GetCalibratedDataCollection(server.MongoDb),
				pipeline:        hourlyPipeline,
			},
		}

		for _, t := range tasks {
			server.Tasks <- &t
		}
	})
	c.AddFunc("@monthly", func() {
		fmt.Println("Every hour on the half hour")

		// tasks := []TaskMongoAggregation{}

		// for _, t := range tasks {
		// 	server.Tasks <- &t
		// }
	})
	c.AddFunc("@yearly", func() {
		fmt.Println("Every hour on the half hour")

		// tasks := []TaskMongoAggregation{}

		// for _, t := range tasks {
		// 	server.Tasks <- &t
		// }
	})

	c.Start()

	// TODO: Need to handle exits - bit of a mess currently

	go func() {
		<-server.ExitContext.Done()
		log.Println("Error identified in core server, existing scheduler...")
		c.Stop()
	}()

}

func SetupTaskHandler(server *Server) {
	debounce(time.Second*1, server.Tasks, taskHandler)
}
