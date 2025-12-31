package main

import (
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// ensureTopicExists creates the topic if it doesn't exist
func ensureTopicExists(broker string, topic string) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return fmt.Errorf("failed to connect to kafka: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}
	controllerConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return fmt.Errorf("failed to connect to controller: %w", err)
	}
	defer controllerConn.Close()

	// Delete topic if it exists (to start fresh)
	deleteTopics := []string{topic}
	err = controllerConn.DeleteTopics(deleteTopics...)
	if err != nil {
		log.Printf("Note: Could not delete topic (may not exist): %v", err)
	} else {
		log.Printf("Deleted existing topic '%s' to start fresh", topic)
		// Wait a moment for deletion to complete
		time.Sleep(2 * time.Second)
	}

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     3,
			ReplicationFactor: 1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	log.Printf("Topic '%s' created successfully with 3 partitions", topic)
	// Give Kafka a moment to fully create the topic
	time.Sleep(1 * time.Second)
	return nil
}

// waitForKafkaReady waits for Kafka to be ready
func waitForKafkaReady(broker string, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		conn, err := kafka.Dial("tcp", broker)
		if err == nil {
			conn.Close()
			return nil
		}
		log.Printf("Waiting for Kafka to be ready... (attempt %d/%d)", i+1, maxRetries)
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("Kafka not ready after %d attempts", maxRetries)
}

