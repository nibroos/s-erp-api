package config

import (
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

func GetDatabaseURL() string {
	env := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)

	return env
}

func GetTestDatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER_TEST"),
		os.Getenv("POSTGRES_PASSWORD_TEST"),
		os.Getenv("POSTGRES_HOST_TEST"),
		os.Getenv("POSTGRES_PORT_TEST"),
		os.Getenv("POSTGRES_DB_TEST"),
	)
}

func NewRabbitMQ() (*RabbitMQ, error) {
	user := os.Getenv("RABBITMQ_USER")
	password := os.Getenv("RABBITMQ_PASSWORD")
	host := os.Getenv("RABBITMQ_HOST")
	port := os.Getenv("RABBITMQ_PORT")
	vhost := os.Getenv("RABBITMQ_VHOST")

	url := fmt.Sprintf("amqp://%s:%s@%s:%s/%s", user, password, host, port, vhost)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %v", err)
	}

	return &RabbitMQ{
		Connection: conn,
		Channel:    ch,
	}, nil
}

func (r *RabbitMQ) Close() {
	if r.Channel != nil {
		if err := r.Channel.Close(); err != nil {
			log.Printf("Error closing RabbitMQ channel: %v", err)
		}
	}
	if r.Connection != nil {
		if err := r.Connection.Close(); err != nil {
			log.Printf("Error closing RabbitMQ connection: %v", err)
		}
	}
}
