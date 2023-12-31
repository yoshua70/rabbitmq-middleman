# RabbitMQ Middleman

A middleman to consume messages from a RabbitMQ queue and sent them through a websocket connection to a client.

## Requirements

This project was developed using Go v1.21.4.

You will also need a RabbitMQ instance. It is recommended to use a docker image.

## Usage

Clone the current repository and build an executable of the project:

```sh
go build -o middleman main.go
```

The basic usage is the following:

```sh
middleman [rabbitmq_url] [rabbitmq_queue_name] [listening_port]
```
