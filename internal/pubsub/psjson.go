package pubsub

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	bytes, err := json.Marshal(val)
	if err != nil {
		return err
	}
	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{ContentType: "application/json", Body: bytes})
	if err != nil {
		return err
	}

	return nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) Acktype,
	dlx bool,
) error {
	c, _, _ := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType, dlx)
	cha, _ := c.Consume(queueName, "", false, false, false, false, nil)
	go func(ch <-chan amqp.Delivery) {
		var t T
		for message := range ch {
			json.Unmarshal(message.Body, &t)
			ackt := handler(t)
			switch ackt {
			case Ack:
				//message.Ack(false)
				//fmt.Println("Ack")
				amqp.Delivery.Ack(message, false)
			case NackRequeue:
				//message.Nack(false, true)
				//fmt.Println("NackRequeue")
				amqp.Delivery.Nack(message, false, true)

			case NackDiscard:
				//message.Nack(false, false)
				//fmt.Println("NackDiscard")
				amqp.Delivery.Nack(message, false, false)

			}

		}
	}(cha)
	return nil
}
