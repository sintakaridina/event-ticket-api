package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// RabbitMQConnection represents a connection to RabbitMQ
type RabbitMQConnection struct {
	Connection *amqp.Connection
}

// NewRabbitMQConnection creates a new RabbitMQ connection
func NewRabbitMQConnection() (*RabbitMQConnection, error) {
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

	// Connect to RabbitMQ with retry logic
	var conn *amqp.Connection
	var err error

	maxRetries := 10
	retryDelay := 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}

		logrus.Warnf("Failed to connect to RabbitMQ (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(retryDelay)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w", maxRetries, err)
	}

	logrus.Info("Connected to RabbitMQ")
	return &RabbitMQConnection{Connection: conn}, nil
}

// Close closes the RabbitMQ connection
func (r *RabbitMQConnection) Close() error {
	if r.Connection != nil {
		return r.Connection.Close()
	}
	return nil
}

// DeclareExchange declares a RabbitMQ exchange
func DeclareExchange(conn *RabbitMQConnection, exchangeName string) error {
	ch, err := conn.Connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchangeName, // name
		"topic",     // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare an exchange: %w", err)
	}

	logrus.Infof("Declared exchange: %s", exchangeName)
	return nil
}

// PublishMessage publishes a message to a RabbitMQ exchange
func PublishMessage(conn *RabbitMQConnection, exchangeName, routingKey string, message interface{}) error {
	ch, err := conn.Connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	// Convert message to JSON
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Publish message
	err = ch.Publish(
		exchangeName, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	logrus.Infof("Published message to exchange %s with routing key %s", exchangeName, routingKey)
	return nil
}

// ConsumeMessages consumes messages from a RabbitMQ queue
func ConsumeMessages(conn *RabbitMQConnection, queueName, exchangeName, routingKey string, handler func([]byte) error) error {
	ch, err := conn.Connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	// Declare a queue
	q, err := ch.QueueDeclare(
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

	// Bind the queue to the exchange
	err = ch.QueueBind(
		q.Name,       // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind a queue: %w", err)
	}

	// Set QoS
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Consume messages
	msgs, err := ch.Consume(
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

	logrus.Infof("Started consuming messages from queue: %s", queueName)

	// Process messages
	for msg := range msgs {
		logrus.Infof("Received a message: %s", msg.RoutingKey)

		// Process the message
		err := handler(msg.Body)
		if err != nil {
			logrus.Errorf("Error processing message: %v", err)
			// Nack the message and requeue it
			msg.Nack(false, true)
			continue
		}

		// Acknowledge the message
		msg.Ack(false)
	}

	return nil
}

// ReconnectIfNeeded checks if the RabbitMQ connection is closed and reconnects if needed
func (r *RabbitMQConnection) ReconnectIfNeeded() error {
	if r.Connection == nil || r.Connection.IsClosed() {
		logrus.Warn("RabbitMQ connection is closed, reconnecting...")
		newConn, err := NewRabbitMQConnection()
		if err != nil {
			return fmt.Errorf("failed to reconnect to RabbitMQ: %w", err)
		}
		r.Connection = newConn.Connection
		logrus.Info("Reconnected to RabbitMQ")
	}
	return nil
}