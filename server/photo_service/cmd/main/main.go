package main

import (
	"os"
	"os/signal"
	amqpparams "photo_service/internal/amqp_params"
	deliveryhandler "photo_service/internal/delivery_handler"
	"photo_service/internal/repository"
	"syscall"

	"github.com/sirupsen/logrus"
)

func main() {
	amqpParams := amqpparams.New()
	msgs, err := amqpParams.Channel.Consume(
		amqpParams.Queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logrus.Fatalln(err)
	}
	repo := repository.New()

	// Starting listener goroutine
	logrus.Info("application is being launched!")
	go func() {
		for d := range msgs {
			deliveryhandler.Handle(d, repo)
		}
	}()

	// Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	stoppingSignal := <-stop
	logrus.WithField("signal", stoppingSignal).Info("application has been stopped")

	amqpParams.Channel.Close()
	amqpParams.Conn.Close()
}
