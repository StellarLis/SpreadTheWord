package amqp

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type Amqp struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	Queue   *amqp.Queue
}

func New() *Amqp {
	conn, err := amqp.Dial("amqp://guest:guest@host.docker.internal:5672")
	if err != nil {
		logrus.Fatalln(err)
	}
	ch, err := conn.Channel()
	if err != nil {
		logrus.Fatalln(err)
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
		logrus.Fatalln(err)
	}
	return &Amqp{Conn: conn, Channel: ch, Queue: &q}
}

func (a *Amqp) Close() {
	a.Channel.Close()
	a.Conn.Close()
}
