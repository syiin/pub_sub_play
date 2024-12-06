package pubsub

import "context"
import "fmt"
import "log"
import "encoding/json"
import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType string

const (
	Ack         AckType = "Ack"
	NackRequeue AckType = "NackRequeue"
	NackDiscard AckType = "NackDiscard"
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
	handler func(T) AckType,
) error {
	channel, _, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
	)
	if err != nil {
		fmt.Println("Channel not opened")
	}
	deliveryChan, err := channel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		fmt.Println("Delivery channel cannot be opened")
	}
	go func() {
		for msg := range deliveryChan {
			var unmarshalled T
			json.Unmarshal(msg.Body, &unmarshalled)
			returnedAck := handler(unmarshalled)
			switch returnedAck {
			case Ack:
				msg.Ack(false)
				log.Println("Ack called with", unmarshalled)
			case NackRequeue:
				msg.Nack(false, true)
				log.Println("NackRequeue called with", unmarshalled)
			case NackDiscard:
				msg.Nack(false, false)
				log.Println("NackDiscard called with", unmarshalled)
			}
		}
	}()
	return err

}
