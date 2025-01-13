package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	// Connect to broker
	conString := "amqp://guest:guest@localhost:5672/"
	con, err := amqp.Dial(conString)
	if err != nil {
		fmt.Println("Connection Error:")
		fmt.Println(err)
		fmt.Println("Closing Program")
		return
	}
	fmt.Println("Connection Successful")
	//When returning main, close connection
	defer con.Close()

	ch, err := con.Channel()
	if err != nil {
		fmt.Println("Channel Error:")
		fmt.Println(err)
		fmt.Println("Closing Program")
		return
	}

	pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})

	//WAIT FOR INTERRUPT
	fmt.Println("Waiting for control C")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	fmt.Println("Closing Program")
}
