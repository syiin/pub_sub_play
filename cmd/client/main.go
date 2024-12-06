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
		handlerMove(gameState, channel),
	)

	// War
	pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		"war",
		"army_moves.*",
		pubsub.Durable,
		handlerWar(gameState),
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

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, channel *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(move)

		switch moveOutcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(
				channel,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			if err != nil {
				fmt.Println("Unknown move occurred")
			}
			return pubsub.NackRequeue
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		default:
			return pubsub.NackDiscard
		}
	}
}

func handlerWar(gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, _, _ := gs.HandleWar(rw)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			return pubsub.Ack
		default:
			log.Fatal("Received bad war outcome", outcome)
			return pubsub.NackDiscard
		}
	}
}
