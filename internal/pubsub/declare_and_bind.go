package pubsub

import "log"
import "fmt"
import (
	amqp "github.com/rabbitmq/amqp091-go"
)

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
		log.Println("Durable...")
		durable = true
		autoDelete = false
		exclusive = false
	case Transient:
		log.Println("Transient...")
		durable = false
		autoDelete = true
		exclusive = true
	default:
		fmt.Println("Unknown simpleQueueType value:", simpleQueueType)
	}
	table := amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",
	}
	queue, err := channel.QueueDeclare(
		queueName,
		durable,
		autoDelete,
		exclusive,
		false,
		table)
	if err != nil {
		log.Fatal("Could not declare queue", err)
	}

	err = channel.QueueBind(queueName, key, exchange, false, amqp.Table{})
	if err != nil {
		log.Fatal("Could not bind queue: %v", err)
	}
	fmt.Println("Queue Type:", simpleQueueType, ", AutoDelete:", autoDelete, ", Exclusive:", exclusive)
	return channel, queue, err
}
