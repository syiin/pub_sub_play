package main

import "fmt"
import "log"
import (
	amqp "github.com/rabbitmq/amqp091-go"
)
import "github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
import "github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
import "github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"

func main() {
	gamelogic.PrintServerHelp()
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

	// Log
	logKey := fmt.Sprintf("%s.%s", routing.GameLogSlug, "*")
	_, _, err = pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		logKey,
		pubsub.SimpleQueueDurable)

	if err != nil {
		log.Fatal("Could not declare and bind log queue")
	}

	pubsub.SubscribeGob(
		connection,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		logKey,
		pubsub.SimpleQueueDurable
	)
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "pause":
			fmt.Println("Pausing")
			err = pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				})
			if err != nil {
				log.Printf("Could not publish: %v", err)
			}
		case "resume":
			fmt.Println("Resuming")
			err = pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				})
			if err != nil {
				log.Printf("Could not publish: %v", err)
			}
		case "quit":
			log.Printf("Goodbye")
			break
		default:
			fmt.Println("Unknown command")
		}

	}
}
