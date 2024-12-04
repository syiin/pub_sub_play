package main

import "os"
import "os/signal"
import "fmt"
import "log"
import (
	amqp "github.com/rabbitmq/amqp091-go"
)
import "github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
import "github.com/bootdotdev/learn-pub-sub-starter/internal/routing"

func main() {
	connection_str := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(connection_str)
	defer connection.Close()
	if err != nil {
		log.Fatal("Could not connect to server")
	}
	fmt.Println("Connection successful!")

	channel, err := connection.Channel()
	if err != nil {
		log.Fatal("Could not open channel")
	}
	exchange := routing.ExchangePerilDirect
	pauseKey := routing.PauseKey
	playingState := routing.PlayingState{
		IsPaused: false,
	}
	pubsub.PublishJSON(channel, exchange, pauseKey, playingState)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
}
