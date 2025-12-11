package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/streadway/amqp"
)

type Order struct {
	ID        uint    `json:"id"`
	UserID    uint    `json:"user_id"`
	ProductID uint    `json:"product_id"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
}

func main() {
	// Get RabbitMQ config from environment or use defaults
	rabbitHost := getEnv("RABBITMQ_HOST", "localhost")
	rabbitPort := getEnv("RABBITMQ_PORT", "5672")
	rabbitURL := fmt.Sprintf("amqp://guest:guest@%s:%s/", rabbitHost, rabbitPort)

	// Connect to RabbitMQ
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel:", err)
	}
	defer ch.Close()

	// Declare exchange
	err = ch.ExchangeDeclare(
		"orders",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Failed to declare exchange:", err)
	}

	// Declare queue
	q, err := ch.QueueDeclare(
		"order_notifications",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Failed to declare queue:", err)
	}

	// Bind queue to exchange
	err = ch.QueueBind(
		q.Name,
		"order.*",
		"orders",
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Failed to bind queue:", err)
	}

	// Consume messages
	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Failed to register consumer:", err)
	}

	log.Println("Waiting for messages...")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var order Order
			if err := json.Unmarshal(d.Body, &order); err != nil {
				log.Printf("Error parsing message: %v", err)
				continue
			}

			log.Printf("Received order: ID=%d, UserID=%d, Amount=%.2f, Status=%s",
				order.ID, order.UserID, order.Amount, order.Status)

			// Process the order (send email, update inventory, etc.)
		}
	}()

	<-forever
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
