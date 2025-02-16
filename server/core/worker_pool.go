package core

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

const AGGREGATION_TIME_LIMIT = 30 * time.Second

/**
Considersations.
*/

type TaskJobResult struct {
	result any
	err    error
}

// Overall this is overly developed as I don't have more than a single type.
// Partly its to pratice
type TaskJob interface {
	Job(ctx context.Context) TaskJobResult
}

type TaskRedisCheck struct {
	key    string
	client *redis.Client
	// we would want this value to reflect the interval at which the job occurs
	ttl time.Duration
}

func NewTaskMongoAggregation[T any](mongoCollection MongoCollection[T], pipeline mongo.Pipeline, redisCheck *TaskRedisCheck) TaskMongoAggregation[T] {
	return TaskMongoAggregation[T]{
		mongoCollection: mongoCollection,
		pipeline:        pipeline,
		redisCheck:      redisCheck,
	}
}

type TaskMongoAggregation[T any] struct {
	redisCheck      *TaskRedisCheck
	mongoCollection MongoCollection[T]
	pipeline        mongo.Pipeline
}

func (t *TaskMongoAggregation[T]) Job(ctx context.Context) TaskJobResult {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// This allows for an options caching check
	if t.redisCheck != nil {
		_, alreadyHeld, err := GetLock(t.redisCheck.client, t.redisCheck.key, t.redisCheck.ttl)
		if alreadyHeld {
			log.Printf("Unable to acquire lock for key %s, already claimed (this is expected for multiple applications)\n", t.redisCheck.key)
			// Any process is already performing this job
			return TaskJobResult{}
		}

		if err != nil {
			return TaskJobResult{err: err}
		}

		// we deliberately are NOT releasing the lock but instead setting the TTL to be released in the future.
	}

	_, err := t.mongoCollection.Aggregate(ctx, t.pipeline)

	result := TaskJobResult{
		err: err,
	}

	return result
}

func worker(id int, tasks <-chan TaskJob, results chan<- TaskJobResult) {
	for task := range tasks {
		log.Printf("Worker %d processing task: %+v\n", id, task)

		err := task.Job(context.TODO())

		results <- err
	}
}

func taskHandler(items []TaskJob, goroutineCount int) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	taskNum := len(items)
	tasks := make(chan TaskJob, taskNum)
	taskErrors := make(chan TaskJobResult, taskNum)

	// Start worker goroutines
	for i := 0; i < goroutineCount; i++ { // Adjust number of workers as needed
		go worker(i, tasks, taskErrors)
	}

	// Send tasks to workers
	for _, item := range items {
		tasks <- item
	}
	close(tasks)

	// Collect errors
	var errs []error
	for i := 0; i < taskNum; i++ {
		select {
		case result := <-taskErrors:
			if result.err != nil {
				errs = append(errs, result.err)
			}
		case <-ctx.Done():
			errs = append(errs, ctx.Err())
			break
		}
	}

	// Handle errors (log them, send to error channel, etc.)
	if len(errs) > 0 {
		log.Printf("Encountered %d errors during task processing", len(errs))
		for _, err := range errs {
			log.Printf("Error: %v", err)
		}
	}
}
