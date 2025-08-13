package contracts

import "context"

// Message represents a message received from a Pub/Sub channel.
type Message struct {
	Channel string
	Payload string
}

// This is the contract that our application services will depend on.
type PubSub interface {
	Publish(ctx context.Context, channel string, message string) error
	Subscribe(ctx context.Context, channels ...string) (<-chan *Message, error)
}