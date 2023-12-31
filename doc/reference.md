## Struct

The `Message` struct is used to represent a message consumed from the RabbitMQ queue:

```go
type Message struct {
    Body string `json:"body"`
    Sent bool `json:"sent"`
}
```

The `Body` field holds the value from the `Delivery.Body` field where `Delivery` is a struct from the Go RabbitMQ library that holds messages coming from a queue.

The `Sent` filed is used to keep track of wherever or not a message was previously sent to the client, thus avoiding sending already sent message over and over again. This is needed since we store the consumed message from the queue when no client is connected and then send the stored messages to a client when it connects. By using this field, only the messages which have not been sent before will be send to the client.

## Functions

### RabbitMQ-related functions

#### `connectToRabbitMQ`

```go
func connectToRabbitMQ(connectionUrl string) (*amqp.Connection, error)
```

The function connects to a RabbitMQ instance using the provided connection URL and the `Dial` function from the RabbitMQ library. It then returns a pointer to a RabbitMQ connection object along with an error if there was one.

In case of an error while trying to connect to RabbitMQ, the code does not panic, you have to check for `err != nil`.

#### `openRabbitMqChan`

```go
func openRabbitMqChan(conn *amqp.Connection) (*amqp.Channel, error)
```

The function establishes a channel for the provided RabbitMQ connection.

In case of an error while trying to establish a channel, the code does not panic, you have to check for `err != nil`.

#### `declareQueue`

```go
func declareQueue(channel *amqp.Channel, queueName string) (*amqp.Queue, error)
```

The function declares a queue with name `queueName` for the given channel.

In case of an error while trying to declare a queue, the code does not panic, you have to check for `err != nil`.

#### `consumeFromQueue`

```go
func consumeFromQueue(channel *amqp.Channel, queueName string)
```

The function consumes messages from the queue `queueName` and store them using the `storeMessage` function.

### Message-related functions
