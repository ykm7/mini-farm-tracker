package core

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/robfig/cron/v3"
)

type AGGREGATION_PERIOD string

const (
	HOURLY  AGGREGATION_PERIOD = "%Y-%m-%d-%H"
	DAILY   AGGREGATION_PERIOD = "%Y-%m-%d"
	WEEKLY  AGGREGATION_PERIOD = "%Y-%W"
	MONTHLY AGGREGATION_PERIOD = "%Y-%m"
	YEARLY  AGGREGATION_PERIOD = "%Y"
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
		fmt.Println("Every hour")

		// tasks := []TaskMongoAggregation{}

		// for _, t := range tasks {
		// 	server.Tasks <- &t
		// }
	})
	c.AddFunc("@weekly ", func() {
		fmt.Println("Every week")

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
		fmt.Println("Every month")
	})
	c.AddFunc("@yearly", func() {
		fmt.Println("Every year")
	})

	c.Start()

	// TODO: Need to handle exits - bit of a mess currently

	go func() {
		<-server.ExitContext.Done()
		log.Println("Error identified in core server, existing scheduler...")
		c.Stop()
	}()

}

/*
* Given the jobs are to be IO bound with the expected waiting and context switching,
additional goroutines over the CPU count can be benefical. If the tasks were CPU bound, than
max parallelise achieves benefits with max CPU count.

* TODO: Further considersation; only tasks to process are IO bound, if we were to mix CPU bound tasks,
maybe a separate handler would be better.
*/
func SetupTaskHandler(server *Server) {
	goroutineCount := runtime.NumCPU() * 4

	debounce(time.Second*1, 100, server.Tasks, taskHandler, goroutineCount)
}
