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

func PingRedis(client *redis.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return client.Ping(ctx).Err()
}

func getKey(dataType CalibratedDataNames, aggregationType AGGREGATION_TYPE, groupByFormat AGGREGATION_PERIOD) string {
	return fmt.Sprintf("%s-%s-%s", dataType, aggregationType, groupByFormat)
}

func GetLock(client *redis.Client, key string, duration time.Duration) (lock *redislock.Lock, alreadyHeld bool, err error) {

	locker := redislock.New(client)
	lock, err = locker.Obtain(context.Background(), key, duration, nil)
	if err != nil {
		if errors.Is(err, redislock.ErrNotObtained) {
			// Handle the case where the lock is already held
			// I wouldn't want to consider the lock already been held as an error.
			// This is to be used to sync between 2 or more running "pods".
			return nil, true, nil
		} else {
			return nil, false, err
		}
	}
	return lock, false, nil
}
