package pubsub

import "fmt"
import "log"

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int
type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)
const (
	Ack AckType = iota
	NackDiscard
	NackRequeue
)

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
	unmarshaller func([]byte) (T, error),
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
			unmarshalled, _ = unmarshaller(msg.Body)
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
