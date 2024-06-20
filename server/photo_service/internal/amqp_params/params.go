package amqpparams

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type AmqpParams struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	Queue   amqp.Queue
}

func New() *AmqpParams {
	conn, err := amqp.Dial("amqp://guest:guest@host.docker.internal:5672")
	if err != nil {
		logrus.Fatalln("Error opening connection: ", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		logrus.Fatalln("Error opening channel: ", err)
	}
	q, err := ch.QueueDeclare(
		"photo_queue",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logrus.Fatalln("Error creating queue: ", err)
	}
	return &AmqpParams{Conn: conn, Channel: ch, Queue: q}
}
