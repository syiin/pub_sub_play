package pubsub

import "log"
import "bytes"
import "context"
import "encoding/gob"
import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var data bytes.Buffer
	enc := gob.NewEncoder(&data)
	err := enc.Encode(val)
	if err != nil {
		log.Fatalf("Could not encode: %v", err)
	}
	msg := amqp.Publishing{
		ContentType: "application/gob",
		Body:        data.Bytes(),
	}
	return ch.PublishWithContext(context.Background(), exchange, key, false, false, msg)
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {
	return subscribe(conn, exchange, queueName, key, simpleQueueType, handler, gobUnmarshaller)
}

func gobUnmarshaller[T any](data []byte) (T, error) {
	var result T
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	err := decoder.Decode(&result)
	return result, err

}
