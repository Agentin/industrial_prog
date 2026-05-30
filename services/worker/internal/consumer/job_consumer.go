package consumer

import (
	"encoding/json"
	"log"
	"time"

	"github.com/streadway/amqp"
)

type JobMessage struct {
	Job       string `json:"job"`
	TaskID    string `json:"task_id"`
	Attempt   int    `json:"attempt"`
	MessageID string `json:"message_id"`
	CreatedAt string `json:"created_at"`
}

type JobConsumer struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	mainQueue    string
	dlqQueue     string
	maxAttempts  int
	processedIDs map[string]bool
}

func NewJobConsumer(amqpURL, mainQueue, dlqQueue string, maxAttempts int) (*JobConsumer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	_, err = ch.QueueDeclare(mainQueue, true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}
	_, err = ch.QueueDeclare(dlqQueue, true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}
	return &JobConsumer{
		conn:         conn,
		channel:      ch,
		mainQueue:    mainQueue,
		dlqQueue:     dlqQueue,
		maxAttempts:  maxAttempts,
		processedIDs: make(map[string]bool),
	}, nil
}

func (c *JobConsumer) Start() error {
	err := c.channel.Qos(1, 0, false)
	if err != nil {
		return err
	}
	msgs, err := c.channel.Consume(c.mainQueue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for d := range msgs {
			c.processMessage(d)
		}
	}()
	return nil
}

func (c *JobConsumer) processMessage(d amqp.Delivery) {
	var job JobMessage
	if err := json.Unmarshal(d.Body, &job); err != nil {
		log.Printf("Invalid job format: %v", err)
		d.Nack(false, false)
		return
	}
	if c.processedIDs[job.MessageID] {
		log.Printf("Duplicate message ignored: %s", job.MessageID)
		d.Ack(false)
		return
	}
	log.Printf("Processing job: task_id=%s attempt=%d msg_id=%s", job.TaskID, job.Attempt, job.MessageID)
	time.Sleep(2 * time.Second)
	if job.TaskID == "t_fail" || job.TaskID == "fail" {
		log.Printf("Simulated failure for task_id=%s", job.TaskID)
		if job.Attempt < c.maxAttempts {
			job.Attempt++
			c.retryJob(job)
		} else {
			c.sendToDLQ(job)
		}
		d.Ack(false)
		return
	}
	c.processedIDs[job.MessageID] = true
	log.Printf("Job processed successfully: %s", job.MessageID)
	d.Ack(false)
}

func (c *JobConsumer) retryJob(job JobMessage) {
	body, err := json.Marshal(job)
	if err != nil {
		log.Printf("Failed to marshal retry job: %v", err)
		return
	}
	err = c.channel.Publish("", c.mainQueue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
	})
	if err != nil {
		log.Printf("Failed to publish retry: %v", err)
	} else {
		log.Printf("Retry scheduled: attempt=%d msg_id=%s", job.Attempt, job.MessageID)
	}
}

func (c *JobConsumer) sendToDLQ(job JobMessage) {
	body, err := json.Marshal(job)
	if err != nil {
		log.Printf("Failed to marshal DLQ job: %v", err)
		return
	}
	err = c.channel.Publish("", c.dlqQueue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
	})
	if err != nil {
		log.Printf("Failed to publish to DLQ: %v", err)
	} else {
		log.Printf("Sent to DLQ: task_id=%s msg_id=%s", job.TaskID, job.MessageID)
	}
}

func (c *JobConsumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
