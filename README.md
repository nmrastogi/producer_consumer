# Producer-Consumer Patterns in Go

This repository demonstrates producer-consumer patterns in Go using two different approaches:
1. **In-memory channels** - A simple producer-consumer implementation using Go channels
2. **Apache Kafka** - A distributed messaging system implementation

## Features

- ✅ In-memory producer-consumer using Go channels and goroutines
- ✅ Kafka-based producer-consumer with retry logic and error handling
- ✅ Automatic topic creation and Kafka readiness checks
- ✅ Consumer group support with load balancing
- ✅ Exponential backoff retry mechanism

## Prerequisites

- **Go 1.25+** - [Install Go](https://go.dev/doc/install)
- **Docker** - For running Kafka (optional, only needed for Kafka examples)

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd producer_consumer
```

2. Install dependencies:
```bash
go mod download
```

## Running the Programs

### 1. In-Memory Producer-Consumer

This example uses Go channels for communication between producers and consumers:

```bash
go run producer_consumer.go
```

**What it does:**
- Creates 2 producers that each produce 10 jobs
- Creates 2 consumers that consume jobs from a shared channel
- Uses a buffered channel (capacity 100) to hold jobs
- Demonstrates concurrent processing with goroutines

### 2. Kafka Producer-Consumer

#### Step 1: Start Kafka

Start Kafka using Docker:

```bash
docker run -d --name kafka -p 9092:9092 \
  -e KAFKA_NODE_ID=1 \
  -e KAFKA_PROCESS_ROLES=broker,controller \
  -e KAFKA_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093 \
  -e KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=PLAINTEXT:PLAINTEXT,CONTROLLER:PLAINTEXT \
  -e KAFKA_CONTROLLER_QUORUM_VOTERS=1@localhost:9093 \
  -e KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER \
  -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
  apache/kafka:latest
```

Verify Kafka is running:
```bash
docker ps --filter "name=kafka"
```

#### Step 2: Run the Producer

In one terminal, start the producer:

```bash
cd kafka
go run producer.go topic_init.go
```

**Important:** You must include both `producer.go` and `topic_init.go` because the producer uses helper functions from `topic_init.go`.

The producer will:
- Create 2 producer goroutines
- Each producer sends 10 messages to the "jobs" topic
- Messages are formatted as "job-{id}"

#### Step 3: Run the Consumer

In another terminal, start the consumer:

```bash
cd kafka
go run consumer.go topic_init.go
```

**Important:** You must include both `consumer.go` and `topic_init.go` because the consumer uses helper functions from `topic_init.go`.

The consumer will:
- Wait for Kafka to be ready
- Ensure the "jobs" topic exists
- Start 2 consumer goroutines in a consumer group
- Process messages with retry logic and exponential backoff
- Display consumed messages with partition and offset information

#### Stopping Kafka

When done, stop and remove the Kafka container:

```bash
docker stop kafka
docker rm kafka
```

## Project Structure

```
producer_consumer/
├── producer_consumer.go    # In-memory producer-consumer using channels
├── kafka/
│   ├── producer.go         # Kafka producer implementation
│   ├── consumer.go         # Kafka consumer with retry logic
│   └── topic_init.go       # Helper functions for topic creation and Kafka readiness
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
└── README.md              # This file
```

## Dependencies

- [github.com/segmentio/kafka-go](https://github.com/segmentio/kafka-go) - Kafka client library for Go

## Key Concepts Demonstrated

### In-Memory Pattern
- **Channels**: Buffered channels for job queue
- **Goroutines**: Concurrent producers and consumers
- **WaitGroup**: Synchronization for consumer completion
- **Channel closing**: Graceful shutdown pattern

### Kafka Pattern
- **Producer**: Writing messages to Kafka topics
- **Consumer Groups**: Multiple consumers sharing work
- **Retry Logic**: Exponential backoff for error handling
- **Topic Management**: Automatic topic creation
- **Connection Handling**: Timeout-based message reading

## Error Handling

The Kafka consumer includes robust error handling:
- **Retry mechanism**: Automatically retries on errors
- **Exponential backoff**: Increases wait time between retries (up to 10 seconds)
- **Timeout handling**: Uses context timeouts for message reading
- **Connection recovery**: Handles Kafka coordinator unavailability

## Notes

- The Kafka consumer uses `StartOffset: kafka.LastOffset` to start reading from the end if no offset exists
- The consumer group "worker-group" allows multiple consumers to share the workload
- The producer auto-creates the "jobs" topic if it doesn't exist
- Both programs demonstrate concurrent processing patterns in Go

## License

This project is for educational purposes.

