package contracts

import "context"

// Queue defines the interface for a job queue system.
type Queue interface {
	Enqueue(ctx context.Context, queueName string, job interface{}) error
	Dequeue(ctx context.Context, queueName string, result interface{}) error
}