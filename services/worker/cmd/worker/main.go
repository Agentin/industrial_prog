package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/student/tech-ip-sem2/services/worker/internal/consumer"
)

func main() {
	amqpURL := os.Getenv("RABBIT_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}
	mainQueue := os.Getenv("JOB_QUEUE_NAME")
	if mainQueue == "" {
		mainQueue = "task_jobs"
	}
	dlqQueue := os.Getenv("DLQ_QUEUE_NAME")
	if dlqQueue == "" {
		dlqQueue = "task_jobs_dlq"
	}
	maxAttempts := 3

	consumer, err := consumer.NewJobConsumer(amqpURL, mainQueue, dlqQueue, maxAttempts)
	if err != nil {
		log.Fatal("Failed to create consumer:", err)
	}
	defer consumer.Close()

	err = consumer.Start()
	if err != nil {
		log.Fatal("Failed to start consumer:", err)
	}
	log.Println("Job worker started, waiting for messages...")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down job worker")
}
