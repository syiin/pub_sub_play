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

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
	handler func(T),
) error {
	channel, _, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
	)
	deliveryChan, err := channel.Consume(queueName, "", false, false, false, false, nil)
	go func() {
		for msg := range deliveryChan {
			var unmarshalled T
			json.Unmarshal(msg.Body, unmarshalled)
			handler(unmarshalled)
			msg.Ack(false)
		}
	}()
	return err

}
