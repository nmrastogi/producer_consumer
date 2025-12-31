package main

import (
	"context"
	"fmt"
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
			panic(err)
		}
	time.Sleep(time.Millisecond * 100);
	}
}

func main() {
	go producer(1);
	go producer(2);
	time.Sleep(time.Second * 2);
	fmt.Println("all producers completed");
}