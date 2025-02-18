package core

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/robfig/cron/v3"
)

/*
interestingly, need the static "1" values to prevent issues with the date/time being considered incomplete
*/
type AGGREGATION_PERIOD string

const (
	HOURLY_PERIOD AGGREGATION_PERIOD = "%Y-%m-%d-%H"
	DAILY_PERIOD  AGGREGATION_PERIOD = "%Y-%m-%d"
	// I should be able to use "%Y-%W-1" but "%U" is more widely supported
	WEEKLY_PERIOD  AGGREGATION_PERIOD = "%Y-%U-1"
	MONTHLY_PERIOD AGGREGATION_PERIOD = "%Y-%m-01"
	YEARLY_PERIOD  AGGREGATION_PERIOD = "%Y-01-01"
)

type AGGREGATION_TYPE string

const (
	HOURLY_TYPE  AGGREGATION_TYPE = "HOURLY"
	DAILY_TYPE   AGGREGATION_TYPE = "DAILY"
	WEEKLY_TYPE  AGGREGATION_TYPE = "WEEKLY"
	MONTHLY_TYPE AGGREGATION_TYPE = "MONTHLY"
	YEARLY_TYPE  AGGREGATION_TYPE = "YEARLY"
)

// TODO: This needs further considersation.
// The idea is that this value is used with the TTL values to claim ownership of the aggregation action
// (via the unique action key within redis) but is required to be "freed" prior to the subsequent calls.
// 1 minute is somewhat arbitarily selected but should be viable.
const TTL_SUBSTRACTION = 1 * time.Minute

const (
	HOURLY_TTL  = time.Hour - TTL_SUBSTRACTION
	DAILY_TTL   = time.Hour*24 - TTL_SUBSTRACTION
	WEEKLY_TTL  = time.Hour*24*7 - TTL_SUBSTRACTION
	MONTHLY_TTL = time.Hour*24*7*4 - TTL_SUBSTRACTION
	YEARLY_TTL  = time.Hour*24*7*52 - TTL_SUBSTRACTION
)

// TODO: Would be an environment variable
const LOCATION = "Australia/Perth"

func SetupPeriodicTasks(server *Server) {
	loc, err := time.LoadLocation("Australia/Perth")
	if err != nil {
		log.Fatalf("Could not load timezone: %v", err)
	}

	c := cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger)), cron.WithLocation(loc))

	source := GetCalibratedDataCollection(server.MongoDb)
	target := GetAggregatedDataCollection(server.MongoDb)

	// c.AddFunc("* * * * *", func() {
	// 	fmt.Println("Every minute")
	// })
	// c.AddFunc("@hourly", func() {
	// 	fmt.Println("Every hour")
	// })
	c.AddFunc("@daily", func() {
		fmt.Println("Every day")

		aggregation := DAILY_TYPE
		period := DAILY_PERIOD
		ttl := DAILY_TTL
		timeRange := time.Now().In(loc).AddDate(0, 0, -1)

		metricType := RAIN_FALL_HOURLY_DATA_NAMES
		rainfallTask := NewTaskMongoAggregation(
			source,
			target,
			CreateAggregationPipeline(metricType, aggregation, period, timeRange),
			&TaskRedisCheck{
				key:    getKey(metricType, aggregation, period),
				client: server.Redis,
				ttl:    ttl,
			},
		)

		server.Tasks <- &rainfallTask
	})
	c.AddFunc("@weekly ", func() {
		fmt.Println("Every week")

		aggregation := WEEKLY_TYPE
		period := WEEKLY_PERIOD
		ttl := WEEKLY_TTL
		timeRange := time.Now().In(loc).AddDate(0, 0, -7)

		metricType := RAIN_FALL_HOURLY_DATA_NAMES
		rainfallTask := NewTaskMongoAggregation(
			source,
			target,
			CreateAggregationPipeline(metricType, aggregation, period, timeRange),
			&TaskRedisCheck{
				key:    getKey(metricType, aggregation, period),
				client: server.Redis,
				ttl:    ttl,
			},
		)

		server.Tasks <- &rainfallTask
	})
	c.AddFunc("@monthly", func() {
		fmt.Println("Every month")

		aggregation := MONTHLY_TYPE
		period := MONTHLY_PERIOD
		ttl := MONTHLY_TTL
		timeRange := time.Now().In(loc).AddDate(0, -1, 0)

		metricType := RAIN_FALL_HOURLY_DATA_NAMES
		rainfallTask := NewTaskMongoAggregation(
			source,
			target,
			CreateAggregationPipeline(metricType, aggregation, period, timeRange),
			&TaskRedisCheck{
				key:    getKey(metricType, aggregation, period),
				client: server.Redis,
				ttl:    ttl,
			},
		)

		server.Tasks <- &rainfallTask
	})
	c.AddFunc("@yearly", func() {
		fmt.Println("Every year")

		aggregation := YEARLY_TYPE
		period := YEARLY_PERIOD
		ttl := YEARLY_TTL
		timeRange := time.Now().In(loc).AddDate(-1, 0, 0)

		metricType := RAIN_FALL_HOURLY_DATA_NAMES
		rainfallTask := NewTaskMongoAggregation(
			source,
			target,
			CreateAggregationPipeline(metricType, aggregation, period, timeRange),
			&TaskRedisCheck{
				key:    getKey(metricType, aggregation, period),
				client: server.Redis,
				ttl:    ttl,
			},
		)

		server.Tasks <- &rainfallTask
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

	Debounce(server.ExitContext, time.Second*1, 100, server.Tasks, TaskHandler, goroutineCount)
}
