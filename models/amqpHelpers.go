package models

import (
	"os"

	"github.com/streadway/amqp"
)

//GetAMQPClient returns the AMQP clients
func GetAMQPClient() (*amqp.Connection, error) {
	conn, err := amqp.Dial(os.Getenv("AMQP_CONNECTION_URL"))
	if err != nil {
		return nil, err
	}
	return conn, nil
}
