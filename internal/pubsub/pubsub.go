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

const (
	Durable = iota
	Transient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {

	ch, _ := conn.Channel()

	durableBool, autodeleteBool, exclusiveBool := false, false, false
	switch simpleQueueType {
	case Durable:
		{
			durableBool = true
			autodeleteBool = false
			exclusiveBool = false
		}
	case Transient:
		{
			durableBool = false
			autodeleteBool = true
			exclusiveBool = true
		}
	}

	q, _ := ch.QueueDeclare(queueName, durableBool, autodeleteBool, exclusiveBool, false, nil)

	ch.QueueBind(queueName, key, exchange, false, nil)

	return ch, q, nil

}
