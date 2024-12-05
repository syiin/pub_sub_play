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
		pubsub.Transient,
	)

	// Pause
	gameState := gamelogic.NewGameState(username)
	pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilDirect,
		pauseQueueName,
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gameState),
	)

	// Army Moves
	pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		"army_moves."+username,
		"army_moves.*",
		pubsub.Transient,
		handlerMove(gameState),
	)

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
	// signalChan := make(chan os.Signal, 1)
	// signal.Notify(signalChan, os.Interrupt)
	// <-signalChan
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) {
	return func(move gamelogic.ArmyMove) {
		defer fmt.Print("> ")
		gs.HandleMove(move)
	}
}
