package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// RabbitMQ represents a RabbitMQ connection
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewRabbitMQ creates a new RabbitMQ connection
func NewRabbitMQ() (*RabbitMQ, error) {
	// Get RabbitMQ configuration from environment variables
	host := os.Getenv("RABBITMQ_HOST")
	port := os.Getenv("RABBITMQ_PORT")
	user := os.Getenv("RABBITMQ_USER")
	password := os.Getenv("RABBITMQ_PASSWORD")
	vhost := os.Getenv("RABBITMQ_VHOST")

	// Set default values if not provided
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5672"
	}
	if user == "" {
		user = "guest"
	}
	if password == "" {
		password = "guest"
	}
	if vhost == "" {
		vhost = "/"
	}

	// Create connection URL
	url := fmt.Sprintf("amqp://%s:%s@%s:%s%s", user, password, host, port, vhost)

	// Connect to RabbitMQ with retry
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
		return nil, fmt.Errorf("failed to connect to RabbitMQ after 5 attempts: %w", err)
	}

	// Create channel
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &RabbitMQ{
		conn:    conn,
		channel: channel,
	}, nil
}

// DeclareExchange declares an exchange
func (r *RabbitMQ) DeclareExchange(name string) error {
	return r.channel.ExchangeDeclare(
		name,     // name
		"topic", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
}

// PublishMessage publishes a message to an exchange
func (r *RabbitMQ) PublishMessage(exchange, routingKey string, body []byte) error {
	return r.channel.Publish(
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

// ConsumeMessages consumes messages from a queue
func (r *RabbitMQ) ConsumeMessages(exchange, queueName, routingKey string, handler func([]byte) error) error {
	// Declare queue
	q, err := r.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	// Bind queue to exchange
	err = r.channel.QueueBind(
		q.Name,     // queue name
		routingKey, // routing key
		exchange,   // exchange
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind a queue: %w", err)
	}

	// Set QoS
	err = r.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Consume messages
	msgs, err := r.channel.Consume(
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
				// Nack message and requeue
				msg.Nack(false, true)
			} else {
				// Ack message
				msg.Ack(false)
			}
		}
	}()

	return nil
}

// Close closes the RabbitMQ connection
func (r *RabbitMQ) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}

// ReconnectIfNeeded checks if the connection is closed and reconnects if needed
func (r *RabbitMQ) ReconnectIfNeeded(ctx context.Context) error {
	if r.conn.IsClosed() {
		logrus.Info("RabbitMQ connection is closed, reconnecting...")
		newRmq, err := NewRabbitMQ()
		if err != nil {
			return fmt.Errorf("failed to reconnect to RabbitMQ: %w", err)
		}

		// Close old connection and channel
		r.Close()

		// Update connection and channel
		r.conn = newRmq.conn
		r.channel = newRmq.channel

		logrus.Info("Reconnected to RabbitMQ")
	}

	return nil
}