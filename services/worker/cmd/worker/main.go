package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/streadway/amqp"
)

func main() {
	amqpURL := os.Getenv("RABBIT_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}
	queueName := os.Getenv("QUEUE_NAME")
	if queueName == "" {
		queueName = "task_events"
	}

	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open channel:", err)
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Fatal("Failed to declare queue:", err)
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		log.Fatal("Failed to set QoS:", err)
	}

	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal("Failed to register consumer:", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for d := range msgs {
			var event map[string]interface{}
			if err := json.Unmarshal(d.Body, &event); err != nil {
				log.Printf("Invalid message: %v", err)
				d.Nack(false, false)
				continue
			}
			log.Printf("Received event: %v", event)
			d.Ack(false)
		}
	}()

	log.Println("Worker started, waiting for messages...")
	<-stop
	log.Println("Shutting down worker")
}
