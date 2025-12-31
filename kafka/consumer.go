package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func consumer(id int){
	read := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "jobs",
		GroupID:  "worker-group",
	})
	defer read.Close()
	for {
		msg, err := read.ReadMessage(context.Background());
		if err != nil{
			log.Fatal(err)
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

	select {} // block forever
}


