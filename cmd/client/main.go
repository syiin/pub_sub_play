package main

// import "os"
// import "os/signal"
import "fmt"
import "log"
import "github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
import "github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
import "github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	connection_str := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(connection_str)
	defer connection.Close()
	if err != nil {
		log.Fatal("Could not connect to server")
	}
	fmt.Println("Connection successful!")
	username, err := gamelogic.ClientWelcome()

	channel, err := connection.Channel()
	if err != nil {
		log.Fatal("Could not open channel")
	}

	// Create and bind queue
	pauseQueueName := fmt.Sprintf("%s.%s", routing.PauseKey, username)
	pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilDirect,
		pauseQueueName,
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
	)

	// Pause
	gameState := gamelogic.NewGameState(username)
	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilDirect,
		pauseQueueName,
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gameState),
	)
	if err != nil {
		log.Fatalf("Could not subscribe to army moves: %v", err)
	}

	// Army Moves
	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		"army_moves."+username,
		"army_moves.*",
		pubsub.SimpleQueueTransient,
		handlerMove(gameState, channel),
	)

	if err != nil {
		log.Fatalf("Could not subscribe to army moves: %v", err)
	}

	// War
	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.SimpleQueueDurable,
		handlerWar(gameState, channel),
	)
	if err != nil {
		log.Fatalf("Could not subscribe to war declarations: %v", err)
	}

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "spawn":
			err := gameState.CommandSpawn(words)
			if err != nil {
				fmt.Println("Spawn was unsuccessful")
			}
		case "move":
			armyMove, err := gameState.CommandMove(words)
			if err != nil {
				fmt.Println("Move was unsuccessful")
			}
			err = pubsub.PublishJSON(
				channel,
				routing.ExchangePerilTopic,
				"army_moves."+username,
				armyMove,
			)
			if err != nil {
				log.Printf("Could not publish: %v", err)
			}
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spam is not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			break
		}
	}
}
