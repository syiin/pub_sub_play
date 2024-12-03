package main

import "os"
import "os/signal"
import "fmt"
import "log"
import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	connection_str := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(connection_str)
	if err != nil {
		log.Fatal("Could not connect to server")
	}
	fmt.Println("Connection successful!")
	defer connection.Close()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("Starting Peril server...")
}
