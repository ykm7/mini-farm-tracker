package core

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

const NUM_OF_WORKERS = 5
const NUM_BATCH_COUNT = 3
const AGGREGATION_TIME_LIMIT = 30 * time.Second

/**
Considersations.
*/

// Overall this is overly developed as I don't have more than a single type.
// Partly its to pratice
type TaskJob interface {
	Job(ctx context.Context) error
}

type TaskMongoAggregation struct {
	mongoDb  MongoDatabase
	pipeline mongo.Pipeline
}

func (t *TaskMongoAggregation) Job(ctx context.Context) error {
	_, err := t.mongoDb.Collection("").Aggregate(context.TODO(), t.pipeline)

	return err
}

func worker(id int, tasks <-chan TaskJob, results chan<- error) {
	for task := range tasks {
		log.Printf("Worker %d processing task: %+v\n", id, task)

		// start actual task
		err := task.Job(context.TODO())

		// catch results
		results <- err
	}
}

/*
*
 */
func random() {

	getKey("as", "adfasdf", Volume)

	aggregationTasks := []TaskMongoAggregation{}

	t := TaskMongoAggregation{
		mongoDb:  nil,
		pipeline: mongo.Pipeline{},
	}

	aggregationTasks = append(aggregationTasks, t)

	taskNum := len(aggregationTasks)

	tasks := make(chan TaskJob, taskNum)
	taskErrors := make(chan error, taskNum)

	tasks <- &t

	for i := range taskNum {
		go worker(i, tasks, taskErrors)
	}

}

func taskHandler(items []TaskJob) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	taskNum := len(items)
	tasks := make(chan TaskJob, taskNum)
	taskErrors := make(chan error, taskNum)

	// Start worker goroutines
	for i := 0; i < 5; i++ { // Adjust number of workers as needed
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
		case err := <-taskErrors:
			if err != nil {
				errs = append(errs, err)
			}
		case <-ctx.Done():
			errs = append(errs, ctx.Err())
			return
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
