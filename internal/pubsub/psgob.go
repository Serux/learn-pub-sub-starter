package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishGOB[T any](
	ch *amqp.Channel,
	exchange,
	key string,
	val T,
) error {
	var network bytes.Buffer
	enc := gob.NewEncoder(&network)
	enc.Encode(val)

	bytes := network.Bytes()

	err := ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{ContentType: "application/gob", Body: bytes})
	if err != nil {
		return err
	}

	return nil
}

func SubscribeGOB[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) Acktype,
	dlx bool,
) error {
	c, _, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType, dlx)
	if err != nil {
		fmt.Println("Declaring:", err)
	}
	err = c.Qos(10, 0, false)
	if err != nil {
		fmt.Println("Qos:", err)
	}
	cha, err := c.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		fmt.Println("consume:", err)
	}
	go func(ch <-chan amqp.Delivery) {
		var t T
		for message := range ch {
			network := bytes.NewBuffer(message.Body)
			dec := gob.NewDecoder(network)
			dec.Decode(&t)
			ackt := handler(t)
			switch ackt {
			case Ack:
				amqp.Delivery.Ack(message, false)
			case NackRequeue:
				amqp.Delivery.Nack(message, false, true)
			case NackDiscard:
				amqp.Delivery.Nack(message, false, false)

			}

		}
	}(cha)
	return nil
}
