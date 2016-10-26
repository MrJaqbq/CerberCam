package main

import (
	"fmt"
	"os"

	"github.com/streadway/amqp"
)

func closeQueue(queue amqp.Queue, channel *amqp.Channel, connection *amqp.Connection) {
	log.Debugf("Closing queue '%s' and its connections...", queue.Name)
	defer channel.Close()
	defer connection.Close()
}

func openQueue(queueName string) (amqp.Queue, *amqp.Channel, *amqp.Connection) {

	conn, err := amqp.Dial("amqp://guest:guest@cerbercam.cloudapp.net:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a send channel")

	q, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	return q, ch, conn
}

// Send data to queue
func Send(queueName string) {
	q, ch, conn := openQueue(queueName)

	log.Info("Sending message...")

	body := "ALERT"
	err := ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")

	log.Info("Message sent successfully!")

	defer closeQueue(q, ch, conn)
}

// Receive data from queue
func Receive(queueName string) {
	q, ch, conn := openQueue(queueName)

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	i := 0
	for d := range msgs {
		i++
		msg := Deserialize(d.Body)
		log.Infof("Received a message: %s", *msg.Email)

		// open output file
		fo, err := os.Create(fmt.Sprintf("photo_%d.jpg", i))
		failOnError(err, "Failed to create file")

		_, err = fo.Write(msg.Photo)
		failOnError(err, "Failed to write to file")

		// close fo on exit and check for its returned error
		defer func() {
			err := fo.Close()
			failOnError(err, "Failed to close file")
		}()
	}

	defer closeQueue(q, ch, conn)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Criticalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
