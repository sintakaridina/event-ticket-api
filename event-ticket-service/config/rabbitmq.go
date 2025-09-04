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
	// Get RabbitMQ connection parameters from environment variables
	host := getEnv("RABBITMQ_HOST", "localhost")
	port := getEnv("RABBITMQ_PORT", "5672")
	user := getEnv("RABBITMQ_USER", "guest")
	password := getEnv("RABBITMQ_PASSWORD", "guest")
	vhost := getEnv("RABBITMQ_VHOST", "/")

	// Create connection URL
	url := fmt.Sprintf("amqp://%s:%s@%s:%s%s", user, password, host, port, vhost)

	// Connect to RabbitMQ with retry
	var conn *amqp.Connection
	var err error
	for i := 0; i < 5; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		logrus.WithError(err).Warnf("Failed to connect to RabbitMQ, retrying in %d seconds...", i+1)
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create channel
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Create RabbitMQ instance
	rmq := &RabbitMQ{
		conn:    conn,
		channel: channel,
	}

	// Declare exchanges
	exchanges := []string{"user_events", "ticket_events", "payment_events", "notification_events"}
	for _, exchange := range exchanges {
		err = channel.ExchangeDeclare(
			exchange, // name
			"topic",  // type
			true,     // durable
			false,    // auto-deleted
			false,    // internal
			false,    // no-wait
			nil,      // arguments
		)
		if err != nil {
			rmq.Close()
			return nil, fmt.Errorf("failed to declare exchange %s: %w", exchange, err)
		}
	}

	logrus.Info("Connected to RabbitMQ")
	return rmq, nil
}

// Close closes the RabbitMQ connection and channel
func (r *RabbitMQ) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}

// PublishMessage publishes a message to an exchange with a routing key
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
func (r *RabbitMQ) ConsumeMessages(ctx context.Context, queueName, exchange, routingKey string, handler func([]byte) error) error {
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
		return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
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
		return fmt.Errorf("failed to bind queue %s to exchange %s: %w", queueName, exchange, err)
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
		return fmt.Errorf("failed to consume from queue %s: %w", queueName, err)
	}

	// Process messages
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					logrus.Warnf("Channel closed for queue %s", queueName)
					return
				}

				// Process message
				err := handler(msg.Body)
				if err != nil {
					logrus.WithError(err).Errorf("Failed to process message from queue %s", queueName)
					// Nack message and requeue
					msg.Nack(false, true)
				} else {
					// Ack message
					msg.Ack(false)
				}
			}
		}
	}()

	return nil
}