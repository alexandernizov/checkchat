package consumer

import (
	"fmt"
	"os"
	"time"

	"github.com/alexandernizov/checkChat/domain"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Consumer struct {
	topic       string
	consumer    *kafka.Consumer
	outMessages chan domain.KafkaMessage
}

func New(topic string, host string, groupId string) (*Consumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  host,
		"group.id":           groupId,
		"auto.offset.reset":  "latest",
		"enable.auto.commit": true,
	})
	if err != nil {
		fmt.Printf("Failed to create consumer: %s\n", err)
		os.Exit(1)
	}

	err = c.Subscribe(topic, nil)

	resultChan := make(chan domain.KafkaMessage)
	return &Consumer{topic: topic, consumer: c, outMessages: resultChan}, err
}

func (c *Consumer) Stop() {
	close(c.outMessages)
	c.consumer.Close()
}

func (c *Consumer) ResultChannel() <-chan domain.KafkaMessage {
	return c.outMessages
}

func (c *Consumer) Consume() {
	go c.ReadMessage()
}

func (c *Consumer) ReadMessage() {
	for {
		fmt.Println("searching message")
		msg, err := c.consumer.ReadMessage(1 * time.Second)
		if err != nil {
			if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.Code() == kafka.ErrTimedOut {
				continue
			}
			fmt.Printf("Consumer error: %v\n", err)
			continue
		}

		// Обработка сообщения
		newMessage := domain.KafkaMessage{Key: string(msg.Key), Value: string(msg.Value)}
		c.outMessages <- newMessage
	}
}
