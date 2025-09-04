package config

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQ represents a RabbitMQ connection
type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

// NewRabbitMQ creates a new RabbitMQ connection
func NewRabbitMQ() (*RabbitMQ, error) {
	// Get RabbitMQ connection parameters from environment variables
	host := getEnv("RABBITMQ_HOST", "localhost")
	port := getEnv("RABBITMQ_PORT", "5672")
	user := getEnv("RABBITMQ_USER", "guest")
	password := getEnv("RABBITMQ_PASSWORD", "guest")

	// Create connection URL
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, host, port)

	// Connect to RabbitMQ with retry mechanism
	var conn *amqp.Connection
	var err error

	// Retry connection up to 5 times with exponential backoff
	for i := 0; i < 5; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}

		retryDelay := time.Duration(1<<uint(i)) * time.Second
		logrus.Warnf("Failed to connect to RabbitMQ, retrying in %v: %v", retryDelay, err)
		time.Sleep(retryDelay)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ after multiple attempts: %w", err)
	}

	// Create a channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declare exchanges
	exchanges := []string{"user_events", "ticket_events", "payment_events", "notification_events"}
	for _, exchange := range exchanges {
		err = ch.ExchangeDeclare(
			exchange, // name
			"topic",  // type
			true,     // durable
			false,    // auto-deleted
			false,    // internal
			false,    // no-wait
			nil,      // arguments
		)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to declare exchange %s: %w", exchange, err)
		}
	}

	logrus.Info("Connected to RabbitMQ")
	return &RabbitMQ{Conn: conn, Channel: ch}, nil
}

// Close closes the RabbitMQ connection and channel
func (r *RabbitMQ) Close() {
	if r.Channel != nil {
		r.Channel.Close()
	}
	if r.Conn != nil {
		r.Conn.Close()
	}
}

// PublishMessage publishes a message to the specified exchange and routing key
func (r *RabbitMQ) PublishMessage(exchange, routingKey string, body []byte) error {
	return r.Channel.PublishWithContext(
		context.Background(),
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
}

// ConsumeMessages consumes messages from the specified queue
func (r *RabbitMQ) ConsumeMessages(queueName string, handler func([]byte) error) error {
	// Declare a queue
	q, err := r.Channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Start consuming
	msgs, err := r.Channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	// Process messages
	go func() {
		for msg := range msgs {
			err := handler(msg.Body)
			if err != nil {
				logrus.WithError(err).Error("Failed to process message")
				// Nack the message to requeue it
				msg.Nack(false, true)
			} else {
				// Ack the message
				msg.Ack(false)
			}
		}
	}()

	return nil
}