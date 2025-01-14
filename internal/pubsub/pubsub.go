package pubsub

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type Acktype int

const (
	Ack Acktype = iota
	NackRequeue
	NackDiscard
)

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
	dlx bool,
) (*amqp.Channel, amqp.Queue, error) {

	ch, err := conn.Channel()
	if err != nil {
		return ch, amqp.Queue{}, err
	}

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
	t := amqp.Table{}
	if dlx {
		t["x-dead-letter-exchange"] = "peril_dlx"
	}
	q, err := ch.QueueDeclare(queueName, durableBool, autodeleteBool, exclusiveBool, false, t)
	if err != nil {
		return ch, amqp.Queue{}, err
	}

	err = ch.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return ch, amqp.Queue{}, err
	}
	return ch, q, nil

}
