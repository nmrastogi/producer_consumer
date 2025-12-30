package main

import (
	"fmt"
	"sync"
	"time"
)

func producer(id int, jobs chan<- int) {
	for i := 0; i<10; i++{
		job :=id*100+i;
		fmt.Printf("producer %d: produced %d\n", id, job);
		jobs <- job;
		time.Sleep(time.Millisecond * 100);
	}
}

func consumer(id int, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.done();
	for job := range jobs{
		fmt.Printf("consumer %d: consumed %d\n", id, job);
		time.Sleep(time.Millisecond * 200);
	}
}

func main() {
	jobs := make(chan int, 100);
	var wg sync.WaitGroup;

	for i :=1; i<=2; i++{
		wg.Add(1);
		go consumer(i, jobs, &wg);
	}

	for i :=1; i<=2; i++{
		go producer(i,jobs);
	}

	time.Sleep(time.Second * 2);
	close(jobs);
	wg.Wait();

}