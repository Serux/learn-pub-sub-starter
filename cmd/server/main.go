package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
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

	//pubsub.DeclareAndBind(con, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+".*", pubsub.Durable)
	err = pubsub.SubscribeGOB(con, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+".*", pubsub.Durable, handlerLog(), false)
	if err != nil {
		fmt.Println(err)
		return
	}

	gamelogic.PrintServerHelp()
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "pause":
			{
				fmt.Println("Pause Message Sent")
				pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
			}
		case "resume":
			{
				fmt.Println("Resume Message Sent")
				pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
			}
		case "quit":
			{
				fmt.Println("Closing Program")
				return
			}
		default:
			fmt.Println("Unknown command")
		}

	}

	/*
		//WAIT FOR INTERRUPT
		fmt.Println("Waiting for control C")
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt)
		<-signalChan
	*/

}

func handlerLog() func(routing.GameLog) pubsub.Acktype {

	return func(gl routing.GameLog) pubsub.Acktype {
		defer fmt.Print("> ")
		err := gamelogic.WriteLog(gl)
		if err != nil {
			fmt.Println(err)
		}
		return pubsub.Ack
	}

}
