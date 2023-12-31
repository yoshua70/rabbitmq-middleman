package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
)

var queueNameArg string
var rabbitMqUrlArg string
var listeningPortArg int

var messages []Message

// Upgrade a regular http connection to a websocket connection.
var httpToWebSocketUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Body string `json:"body"`
	Sent bool   `json:"sent"`
}

func parseCmdLineArgs() {
	args := os.Args[1:]

	if len(args) < 3 {
		panic(`Not enough arguments provided:
Usage: middleman [rabbitmq_url] [rabbitmq_queue_name] [listening_port]
`)
	}

	rabbitMqUrlArg = args[0]
	queueNameArg = args[1]
	port, err := strconv.Atoi(args[2])

	if err != nil {
		msg := fmt.Sprintf("Could not parse provided port number to integer: %s", err.Error())
		panic(msg)
	}

	listeningPortArg = port
}

// Connect to a RabbitMQ given its connection URL.
func connectToRabbitMQ(connectionUrl string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(connectionUrl)

	if err != nil {
		log.Printf("failed to connect to RabbitMQ: %s\n", err)
		return nil, err
	}

	return conn, nil
}

// Open a channel from a RabbitMQ connection.
func openRabbitMqChan(conn *amqp.Connection) (*amqp.Channel, error) {
	channel, err := conn.Channel()

	if err != nil {
		log.Printf("failed to open RabbitMQ channel: %s\n", err)
		return nil, err
	}

	return channel, nil
}

func declareQueue(channel *amqp.Channel, queueName string) (*amqp.Queue, error) {
	// QueueDeclare(name string, durable bool, autoDelete bool, exclusive bool, noWait bool, args amqp.Table)
	queue, err := channel.QueueDeclare(
		queueName,
		false, // durable
		false, // autoDelete
		false, // exclusive
		false, // no Wait
		nil,
	)

	if err != nil {
		log.Printf("failed to declare queue %s: %s\n", queueName, err)
		return nil, err
	}

	return &queue, nil
}

// Consume message from the given queue.
// Must be launched in a goroutine.
func consumeFromQueue(channel *amqp.Channel, queueName string) {

	messages, err := channel.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Printf("failed to register a consummer for queue %s: %s\n", queueName, err)
		return
	}

	for message := range messages {
		storeMessage(message)
	}

}

// Parse a message received from the queue.
// Messages must be in JSON format as the Message struct defined
// at the top of the file.
func messageParser(message string) (Message, error) {
	msg := Message{Sent: false}
	err := json.Unmarshal([]byte(message), &msg)

	if err != nil {
		log.Printf("failed to decode message: %s\n", err)
		return Message{}, err
	}

	return msg, nil
}

// Store a single message locally.
func storeMessage(message amqp.Delivery) {
	msg, err := messageParser(string(message.Body))
	if err == nil {
		log.Printf("stored message: %v\n", msg)
		messages = append(messages, msg)
	}

}

// Send message through a websocket connection.
func sendMessage(conn *websocket.Conn, message Message) {
	jsonMessage, err := json.Marshal(message)

	if err != nil {
		log.Printf("failed to encode message object into json: %s\n", err)
		return
	}

	err = conn.WriteJSON(jsonMessage)

	if err != nil {
		log.Printf("failed to write json in websocket pipe: %s\n", err)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := httpToWebSocketUpgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Printf("failed to upgrade http connection to websocket: %s\n", err)
		return
	}

	defer conn.Close()

	// Keep the connection alive and consume messages from RabbitMQ.
	for {
		for _, message := range messages {
			if !message.Sent {
				sendMessage(conn, message)
			}
		}
	}
}

func main() {

	parseCmdLineArgs()

	rmqConn, err := connectToRabbitMQ(rabbitMqUrlArg)

	if err != nil {
		panic("failed to connect to RabbitMQ: " + err.Error())
	}

	rmqChan, _ := openRabbitMqChan(rmqConn)
	defer rmqConn.Close()
	defer rmqChan.Close()

	declareQueue(rmqChan, queueNameArg)

	go consumeFromQueue(rmqChan, queueNameArg)

	http.HandleFunc("/ws", handleWebSocket)
	log.Printf("websocket server is running on :%d\n", listeningPortArg)

	http.ListenAndServe(fmt.Sprintf(":%d", listeningPortArg), nil)
}
