package pubsub

import "context"
import "encoding/json"
import "log"
import "fmt"
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

const (
	Durable int = iota
	Transient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		log.Fatal("Could not open channel")
	}

	var durable, autoDelete, exclusive bool

	switch simpleQueueType {
	case Durable:
		fmt.Println("Durable...")
		durable = true
		autoDelete = false
		exclusive = false
	case Transient:
		fmt.Println("Transient...")
		durable = false
		autoDelete = true
		exclusive = true
	default:
		fmt.Println("Unknown simpleQueueType value:", simpleQueueType)
	}
	queue, err := channel.QueueDeclare(queueName, durable, autoDelete, exclusive, false, amqp.Table{})
	if err != nil {
		log.Fatal("Could not declare queue")
	}

	err = channel.QueueBind(queueName, key, exchange, false, amqp.Table{})
	if err != nil {
		log.Fatal("Could not bind queue: %v", err)
	}
	fmt.Println("Queue Type:", simpleQueueType, ", AutoDelete:", autoDelete, ", Exclusive:", exclusive)
	return channel, queue, err
}
