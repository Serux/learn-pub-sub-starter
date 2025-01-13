package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")

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

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println("username Error:")
		fmt.Println(err)
		fmt.Println("Closing Program")
		return
	}

	pubsub.DeclareAndBind(con, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, pubsub.Transient)

	gamestatus := &gamelogic.GameState{}
	gamestatus = gamelogic.NewGameState(username)

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "spawn":
			{
				err := gamestatus.CommandSpawn(input)
				//infantry
				//cavalry
				//artillery

				//americas
				//europe
				//africa
				//asia
				//antarctica
				//australia
				if err != nil {
					fmt.Println(err)
				}
			}
		case "move":
			{
				_, err := gamestatus.CommandMove(input)
				if err != nil {
					fmt.Println(err)
				}
			}
		case "status":
			{
				gamestatus.CommandStatus()
			}
		case "help":
			{
				gamelogic.PrintClientHelp()
			}
		case "spam":
			{
				fmt.Println("Spamming not allowed yet!")
			}
		case "quit":
			{
				gamelogic.PrintQuit()
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

		fmt.Println("Closing Program")
	*/

}
