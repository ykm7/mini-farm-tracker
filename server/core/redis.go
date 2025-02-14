package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
)

func GetRedisClient(envs *environmentVariables) (client *redis.Client, deferFn func()) {
	client = redis.NewClient(&redis.Options{
		Addr: os.Getenv(envs.Redis_conn),
	})

	deferFn = func() {
		err := client.Close()
		if err != nil {
			// Handle error
			panic(err)
		}
	}

	return client, deferFn
}

func getKey(sensorId, period string, metric MetricTypes) string {
	return fmt.Sprintf("%s-%s-%s", sensorId, period, metric)
}

func getLock(key string, duration time.Duration) (*redislock.Lock, error) {
	client := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	})
	locker := redislock.New(client)

	lock, err := locker.Obtain(context.Background(), key, duration, nil)
	if err != nil {
		if errors.Is(err, redislock.ErrNotObtained) {
			// Handle the case where the lock is already held
			// I wouldn't want to consider the lock already been held as an error.
			// This is to be used to sync between 2 or more running "pods".
			return nil, fmt.Errorf("lock already held for key %s", key)
		}

		return nil, err
	}
	return lock, nil
}

func howToUse() {
	lock, err := getLock("task-key", AGGREGATION_TIME_LIMIT)
	if err != nil {
		// Handle error
		return
	}
	defer lock.Release(context.Background())

	// Execute your task here
}
