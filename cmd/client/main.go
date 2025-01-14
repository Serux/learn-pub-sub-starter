package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

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

	publishCh, err := con.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println("username Error:")
		fmt.Println(err)
		fmt.Println("Closing Program")
		return
	}

	gamestate := &gamelogic.GameState{}
	gamestate = gamelogic.NewGameState(username)

	//handle war
	//warch, _, _ := pubsub.DeclareAndBind(con, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, routing.WarRecognitionsPrefix+"."+username, pubsub.Transient)
	pubsub.SubscribeJSON(con, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, routing.WarRecognitionsPrefix+".*", pubsub.Durable, handlerWar(gamestate, publishCh), true)

	//handle pause
	//pubsub.DeclareAndBind(con, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, pubsub.Transient)
	pubsub.SubscribeJSON(con, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, pubsub.Transient, handlerPause(gamestate), true)
	//handle move
	//pubsub.DeclareAndBind(con, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+username, routing.ArmyMovesPrefix+".*", pubsub.Transient)
	pubsub.SubscribeJSON(con, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+username, routing.ArmyMovesPrefix+".*", pubsub.Transient, handlerMove(gamestate, publishCh), true)

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "spawn":
			{
				err := gamestate.CommandSpawn(input)
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
				am, err := gamestate.CommandMove(input)
				if err != nil {
					fmt.Println(err)
					continue
				}
				pubsub.PublishJSON(publishCh, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+username, am)

			}
		case "status":
			{
				gamestate.CommandStatus()
			}
		case "help":
			{
				gamelogic.PrintClientHelp()
			}
		case "spam":
			{
				if len(input) != 2 {
					fmt.Println("2 params only pls")
					continue
				}
				i, err := strconv.ParseInt(input[1], 0, 0)
				if err != nil {
					fmt.Println("2 nan")
					continue
				}
				for j := 0; j < int(i); j++ {
					l := gamelogic.GetMaliciousLog()
					pubsub.PublishJSON(publishCh, routing.ExchangePerilTopic, routing.GameLogSlug+"."+username, l)
				}

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
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {

	return func(ps routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")
		//fmt.Println("PAUSE: ", ps.IsPaused)
		gs.HandlePause(ps)
		return pubsub.Ack
	}

}
func handlerMove(gl *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.Acktype {

	return func(ps gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		out := gl.HandleMove(ps)
		switch out {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			//ch, _, _ := pubsub.DeclareAndBind(con, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, routing.WarRecognitionsPrefix+"."+gl.GetUsername(), pubsub.Transient)

			err := pubsub.PublishJSON(ch,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gl.GetUsername(),
				gamelogic.RecognitionOfWar{
					Attacker: ps.Player,
					Defender: gl.GetPlayerSnap(),
				})
			if err != nil {
				return pubsub.NackRequeue
			}

			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		default:
			return pubsub.NackDiscard
		}
	}

}

func handlerWar(gl *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.Acktype {

	return func(ps gamelogic.RecognitionOfWar) pubsub.Acktype {
		defer fmt.Print("> ")
		out, w, l := gl.HandleWar(ps)
		switch out {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:

			m := fmt.Sprintf("%v won a war against %v", w, l)
			err := pubsub.PublishGOB(ch,
				routing.ExchangePerilTopic,
				routing.GameLogSlug+"."+gl.GetUsername(),
				routing.GameLog{
					CurrentTime: time.Now(),
					Username:    gl.GetUsername(),
					Message:     m,
				})
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:

			m := fmt.Sprintf("%v won a war against %v", w, l)
			err := pubsub.PublishGOB(ch,
				routing.ExchangePerilTopic,
				routing.GameLogSlug+"."+gl.GetUsername(),
				routing.GameLog{
					CurrentTime: time.Now(),
					Username:    gl.GetUsername(),
					Message:     m,
				})
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			m := fmt.Sprintf("A war between %v and %v resulted in a dwaw", w, l)
			err := pubsub.PublishGOB(ch,
				routing.ExchangePerilTopic,
				routing.GameLogSlug+"."+gl.GetUsername(),
				routing.GameLog{
					CurrentTime: time.Now(),
					Username:    gl.GetUsername(),
					Message:     m,
				})
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			return pubsub.NackDiscard
		}
	}

}
