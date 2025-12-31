package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func consumer(id int) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{"localhost:9092"},
		Topic:          "jobs",
		GroupID:        "worker-group",
		GroupBalancers: []kafka.GroupBalancer{kafka.RangeGroupBalancer{}},
	})
	defer reader.Close()

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("consumer %d error: %v, retrying...", id, err)
			time.Sleep(time.Second)
			continue
		}

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
	go consumer(1)
	go consumer(2)

	select {}
}
