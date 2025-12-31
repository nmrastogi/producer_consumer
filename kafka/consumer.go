package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

func consumer(id int) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{"localhost:9092"},
		Topic:          "jobs",
		GroupID:        "worker-group",
		GroupBalancers: []kafka.GroupBalancer{kafka.RangeGroupBalancer{}},
		// StartOffset defaults to FirstOffset for consumer groups, which allows reading existing messages
	})
	defer reader.Close()

	retryDelay := time.Second
	maxRetryDelay := 10 * time.Second
	timeoutCount := 0

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		msg, err := reader.ReadMessage(ctx)
		cancel()

		if err != nil {
			// Check if it's a timeout error (context deadline exceeded)
			// The kafka-go library wraps the error, so we check the error message
			errMsg := err.Error()
			isTimeout := errors.Is(err, context.DeadlineExceeded) || 
				strings.Contains(errMsg, "context deadline exceeded") ||
				strings.Contains(errMsg, "deadline exceeded")
			
			// Check if it's a group coordinator error (transient initialization issue)
			isCoordinatorError := strings.Contains(errMsg, "Group Coordinator Not Available") ||
				strings.Contains(errMsg, "group coordinator") ||
				strings.Contains(errMsg, "[15]")
			
			if isTimeout || isCoordinatorError {
				timeoutCount++
				// Only log every 5th time to reduce noise
				if timeoutCount%5 == 0 {
					if isCoordinatorError {
						log.Printf("consumer %d: waiting for Kafka coordinator to initialize... (attempt %d)", id, timeoutCount)
					} else {
						log.Printf("consumer %d: waiting for messages... (checked %d times)", id, timeoutCount)
					}
				}
				// Use shorter delay for transient errors (coordinator initialization or no messages)
				time.Sleep(1 * time.Second)
				continue
			} else {
				log.Printf("consumer %d error: %v, retrying in %v...", id, err, retryDelay)
				time.Sleep(retryDelay)
				// Exponential backoff with max limit
				retryDelay = time.Duration(float64(retryDelay) * 1.5)
				if retryDelay > maxRetryDelay {
					retryDelay = maxRetryDelay
				}
				continue
			}
		}

		// Reset retry delay and timeout count on success
		retryDelay = time.Second
		timeoutCount = 0

		fmt.Printf(
			"consumer %d consumed %s (partition=%d offset=%d)\n",
			id,
			string(msg.Value),
			msg.Partition,
			msg.Offset,
		)

		time.Sleep(200 * time.Millisecond)
	}
}

func main() {
	// Wait for Kafka to be ready
	log.Println("Waiting for Kafka to be ready...")
	if err := waitForKafkaReady("localhost:9092", 10); err != nil {
		log.Fatalf("Kafka not ready: %v", err)
	}

	// Ensure topic exists
	log.Println("Ensuring topic 'jobs' exists...")
	if err := ensureTopicExists("localhost:9092", "jobs"); err != nil {
		log.Printf("Warning: Could not ensure topic exists: %v", err)
	}

	// Give Kafka time to fully initialize, especially the consumer offsets topic
	// This is important for KRaft mode where the __consumer_offsets topic needs to be created
	log.Println("Waiting for Kafka to fully initialize (this may take 10-30 seconds)...")
	time.Sleep(5 * time.Second)

	log.Println("Starting consumers...")
	go consumer(1)
	go consumer(2)

	select {}
}
