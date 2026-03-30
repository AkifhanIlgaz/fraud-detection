package worker

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Connect(url string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}

	return conn, nil
}
