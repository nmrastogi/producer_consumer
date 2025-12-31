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

**If Kafka container doesn't exist, create it:**

```bash
docker run -d --name kafka -p 9092:9092 \
  -e KAFKA_NODE_ID=1 \
  -e KAFKA_PROCESS_ROLES=broker,controller \
  -e KAFKA_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093 \
  -e KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=PLAINTEXT:PLAINTEXT,CONTROLLER:PLAINTEXT \
  -e KAFKA_CONTROLLER_QUORUM_VOTERS=1@localhost:9093 \
  -e KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER \
  -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
  -e KAFKA_OFFSETS_TOPIC_NUM_PARTITIONS=1 \
  -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
  apache/kafka:latest
```

**If container already exists, remove it first:**

```bash
docker rm -f kafka
```

Then run the `docker run` command above.

**Verify Kafka is running:**

```bash
docker ps --filter "name=kafka"
```

**Wait for Kafka to initialize (30 seconds recommended):**

```bash
echo "Waiting for Kafka to initialize..." && sleep 30
```

**Check Kafka logs (optional):**

```bash
docker logs kafka --tail 20
```

#### Step 2: Run the Consumer

**In Terminal 1, start the consumer:**

```bash
cd kafka
go run consumer.go topic_init.go
```

**Important:** You must include both `consumer.go` and `topic_init.go` because the consumer uses helper functions from `topic_init.go`.

The consumer will:
- Wait for Kafka to be ready
- Delete and recreate the "jobs" topic (to start fresh and reset offsets)
- Wait 5 seconds for Kafka to fully initialize
- Start 2 consumer goroutines in the "worker-group" consumer group
- Kafka assigns partitions to consumers (with 3 partitions and 2 consumers, load is shared)
- Process messages with retry logic and exponential backoff
- Display consumed messages with partition and offset information
- Both consumers will process messages from different partitions

**Leave this terminal running** - the consumer will wait for messages.

#### Step 3: Run the Producer

**In Terminal 2, start the producer:**

```bash
cd kafka
go run producer.go topic_init.go
```

**Important:** You must include both `producer.go` and `topic_init.go` because the producer uses helper functions from `topic_init.go`.

The producer will:
- Wait for Kafka to be ready
- Delete and recreate the "jobs" topic (to start fresh)
- Create 2 producer goroutines
- Each producer sends 10 messages (20 total) to the "jobs" topic
- Uses Hash balancer with unique keys to distribute messages across all 3 partitions
- Messages are formatted as "job-{id}" with keys like "p1-m0", "p1-m1", etc.
- Exit after all messages are sent (~3 seconds)

#### Step 4: Stop and Clean Up Kafka

**After you're done testing, stop and remove the Kafka container:**

**Option 1: Stop first, then remove (recommended):**
```bash
# Stop the container
docker stop kafka

# Remove the container
docker rm kafka
```

**Option 2: Stop and remove in one command:**
```bash
docker stop kafka && docker rm kafka
```

**Option 3: Force remove (if container is stuck or won't stop):**
```bash
docker rm -f kafka
```

**Verify the container is removed:**
```bash
docker ps -a --filter "name=kafka"
```

**Note:** If you see "No such container" or no output, the container has been successfully removed.

#### Troubleshooting Commands

**Check if Kafka container exists:**

```bash
docker ps -a --filter "name=kafka"
```

**View Kafka logs:**

```bash
docker logs kafka
```

**View last 50 lines of Kafka logs:**

```bash
docker logs kafka --tail 50
```

**Restart Kafka container:**

```bash
docker restart kafka
```

**Check Kafka container status:**

```bash
docker inspect kafka --format='{{.State.Status}}'
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
- **Producer**: Writing messages to Kafka topics with Hash balancer for partition distribution
- **Consumer Groups**: Multiple consumers sharing work via partition assignment
- **Partition Distribution**: Hash balancer with unique keys distributes messages across 3 partitions
- **Load Balancing**: With 2 consumers and 3 partitions, Kafka assigns partitions evenly
- **Retry Logic**: Exponential backoff for error handling
- **Topic Management**: Automatic topic deletion and recreation for fresh starts
- **Connection Handling**: Timeout-based message reading with coordinator error handling

## Error Handling

The Kafka consumer includes robust error handling:
- **Retry mechanism**: Automatically retries on errors
- **Exponential backoff**: Increases wait time between retries (up to 10 seconds)
- **Timeout handling**: Uses context timeouts for message reading
- **Connection recovery**: Handles Kafka coordinator unavailability
- **Coordinator errors**: Gracefully handles "Group Coordinator Not Available" errors during initialization

## How Partition Distribution Works

The system is configured to distribute messages across multiple partitions so both consumers can process messages concurrently:

1. **Topic Configuration**: The "jobs" topic has 3 partitions (0, 1, 2)

2. **Producer Distribution**: 
   - Uses Hash balancer with unique keys per message (e.g., "p1-m0", "p1-m1", "p2-m0")
   - Keys hash to different partitions, distributing messages evenly

3. **Consumer Group Assignment**:
   - Both consumers join the "worker-group" consumer group
   - Kafka assigns partitions: Consumer 1 gets partitions 0 and 1, Consumer 2 gets partition 2
   - This ensures both consumers process messages concurrently

4. **Result**: Messages are distributed across all partitions, and both consumers actively process messages from their assigned partitions

## Complete Command Reference

### Quick Start (All Commands)

```bash
# 1. Start Kafka (if not already running)
docker run -d --name kafka -p 9092:9092 \
  -e KAFKA_NODE_ID=1 \
  -e KAFKA_PROCESS_ROLES=broker,controller \
  -e KAFKA_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093 \
  -e KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=PLAINTEXT:PLAINTEXT,CONTROLLER:PLAINTEXT \
  -e KAFKA_CONTROLLER_QUORUM_VOTERS=1@localhost:9093 \
  -e KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER \
  -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
  -e KAFKA_OFFSETS_TOPIC_NUM_PARTITIONS=1 \
  -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
  apache/kafka:latest

# 2. Wait for Kafka to initialize
sleep 30

# 3. Terminal 1 - Start Consumer
cd kafka
go run consumer.go topic_init.go

# 4. Terminal 2 - Start Producer
cd kafka
go run producer.go topic_init.go

# 5. Cleanup (when done)
docker stop kafka && docker rm kafka

# Verify cleanup
docker ps -a --filter "name=kafka"
```

## Notes

- **Topic Management**: Both producer and consumer delete and recreate the "jobs" topic to ensure a fresh start
- **Partition Distribution**: The producer uses a Hash balancer with unique keys (p1-m0, p1-m1, etc.) to distribute messages across all 3 partitions
- **Consumer Group**: The "worker-group" consumer group allows multiple consumers to share the workload via partition assignment
- **Load Balancing**: With 2 consumers and 3 partitions, Kafka assigns:
  - Consumer 1: Partitions 0 and 1
  - Consumer 2: Partition 2
- **Initialization**: Kafka needs 30-60 seconds to fully initialize the consumer offsets topic
- **Error Handling**: The consumer automatically handles coordinator initialization errors and timeouts
- **KRaft Mode**: The `KAFKA_OFFSETS_TOPIC_*` environment variables are important for KRaft mode to work properly
- Both programs demonstrate concurrent processing patterns in Go

## License

This project is for educational purposes.

