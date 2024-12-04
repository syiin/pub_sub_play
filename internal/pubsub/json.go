package pubsub

import "context"
import "encoding/json"
import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonBytes, _ := json.Marshal(val)
	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        jsonBytes,
	}
	return ch.PublishWithContext(context.Background(), exchange, key, false, false, msg)
}
