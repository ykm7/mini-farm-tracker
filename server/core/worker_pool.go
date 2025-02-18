package core

import (
	"context"
	"fmt"
	"log"
	"reflect"
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

func NewTaskMongoAggregation[T any, S interface{}](source MongoCollection[T], target MongoCollection[S], pipeline mongo.Pipeline, redisCheck *TaskRedisCheck) TaskMongoAggregation[T, S] {
	return TaskMongoAggregation[T, S]{
		source:     source,
		target:     target,
		pipeline:   pipeline,
		redisCheck: redisCheck,
	}
}

type TaskMongoAggregation[T any, S interface{}] struct {
	redisCheck *TaskRedisCheck
	source     MongoCollection[T]
	target     MongoCollection[S]
	pipeline   mongo.Pipeline
}

func (t *TaskMongoAggregation[T, S]) Job(ctx context.Context) TaskJobResult {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jobResult := TaskJobResult{}

	// This allows for an options caching check
	if t.redisCheck != nil {
		_, alreadyHeld, err := GetLock(t.redisCheck.client, t.redisCheck.key, t.redisCheck.ttl)
		if alreadyHeld {
			log.Printf("Unable to acquire lock for key %s, already claimed (this is expected for multiple applications)\n", t.redisCheck.key)
			// Any process is already performing this job
			return jobResult
		}

		if err != nil {
			jobResult.err = err
			return jobResult
		}

		// we deliberately are NOT releasing the lock but instead setting the TTL to be released in the future.
	}

	results, err := t.source.Aggregate(ctx, t.pipeline)
	if err != nil {
		jobResult.err = err
		return jobResult
	}

	if len(results) == 0 {
		jobResult.err = fmt.Errorf("aggregation results is zero length... this is likely wrong")
		return jobResult
	}

	docs := make([]S, len(results))
	for i, doc := range results {
		if convertedDoc, ok := doc.(S); ok {
			docs[i] = convertedDoc
		} else {
			jobResult.err = fmt.Errorf("unable to convert %+v to type of %s\n", doc, reflect.TypeFor[S]())
			return jobResult
		}
	}

	_, err = t.target.InsertMany(ctx, docs)
	if err != nil {
		jobResult.err = err
		return jobResult
	}

	return jobResult
}

func worker(id int, tasks <-chan TaskJob, results chan<- TaskJobResult) {
	for task := range tasks {
		log.Printf("Worker %d processing task: %+v\n", id, task)

		err := task.Job(context.TODO())

		results <- err
	}
}

func TaskHandler(items []TaskJob, goroutineCount int) {
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
