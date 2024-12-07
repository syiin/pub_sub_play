package main

import "fmt"
import "github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
import "github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
import "github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"

func handlerGameLogs() func(routing.GameLog) pubsub.AckType {
	return func(gameLog routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")
		gamelogic.WriteLog(gameLog)
		return pubsub.Ack
	}
}
