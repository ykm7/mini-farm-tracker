package core

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
)

const NUM_OF_WORKERS = 5

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

// Purpose of this is to take the conjob style aggregation requests which will be run periodically to "group"
// data pull; ie, sum daily rainfall.
func random() {

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
