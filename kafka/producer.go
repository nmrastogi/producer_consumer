package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func producer(id int){
	write := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "jobs",
	})
	defer write.Close();

	for i :=0; i<10; i++{
		job := fmt.Sprintf("job-%d", id*100+i)
		fmt.Printf("producer %d: produced %s\n", id, job)
		err := write.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(fmt.Sprintf("p%d", id)),
				Value: []byte(job),
			})
		if err != nil {
			log.Printf("producer %d error writing message: %v", id, err)
			return
		}
	time.Sleep(time.Millisecond * 100);
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
		log.Fatalf("Failed to ensure topic exists: %v", err)
	}

	// Give Kafka a moment to fully initialize the topic
	time.Sleep(1 * time.Second)

	log.Println("Starting producers...")
	go producer(1)
	go producer(2)

	// Wait for producers to finish (each produces 10 messages with 100ms delay = ~1 second)
	time.Sleep(time.Second * 2)
	fmt.Println("All producers completed")
}