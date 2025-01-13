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
	//fmt.Println("val:", string(bytes))

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{ContentType: "application/json", Body: bytes})
	if err != nil {
		return err
	}

	return nil
}

type SimpleQueueType int

const (
	Durable SimpleQueueType = iota
	Transient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType, // an enum to represent "durable" or "transient"
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

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T),
) error {
	c, _, _ := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	cha, _ := c.Consume(queueName, "", false, false, false, false, nil)
	go func(ch <-chan amqp.Delivery) {
		var t T
		for message := range ch {
			json.Unmarshal(message.Body, &t)
			handler(t)
			amqp.Delivery.Ack(message, false)
		}
	}(cha)
	return nil
}
