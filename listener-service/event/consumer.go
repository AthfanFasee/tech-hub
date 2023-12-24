package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	logs "github.com/AthfanFasee/listener/proto"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Consumer struct {
	conn *amqp.Connection
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}

	err := consumer.setup()
	if err != nil {
		return Consumer{}, err
	}

	return consumer, nil
}

func (consumer *Consumer) setup() error {
	// Channel acts as a communication pathway for the service to talk to RabbitMQ
	channel, err := consumer.conn.Channel()
	if err != nil {
		return err
	}

	return declareExchange(channel)
}

type PayLoad struct {
	Name string `json:"name"`
	Data any    `json:"data"`
}

func (consumer *Consumer) Listen(topics []string) error {
	ch, err := consumer.conn.Channel()
	if err != nil {
		return err
	}

	defer ch.Close()

	q, err := declareRandomQueue(ch)
	if err != nil {
		return err
	}
	// Whenever a message is sent to the "logs_topic" exchange with a topic that matches any in the topics slice,
	// it'll be routed to the queue represented by q.Name.
	for _, s := range topics {
		ch.QueueBind(
			q.Name,       // Queue(random) Name
			s,            // Topic name
			"logs_topic", // Exchange Name
			false,
			nil,
		)

		if err != nil {
			return err
		}
	}

	// Consume the message and handle it
	messages, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	forever := make(chan bool)

	go func() {
		for d := range messages {
			var payload PayLoad
			_ = json.Unmarshal(d.Body, &payload)

			go handlePayload(payload)
		}
	}()

	fmt.Printf("Waiting for message [Exchange, Queue] [logs_topic, %s]\n", q.Name)

	<-forever

	return nil
}

func handlePayload(payload PayLoad) {
	switch payload.Name {
	case "log", "error":
		err := logViaGRPC(payload)
		if err != nil {
			log.Println(err)
		}

	case "mail":
		// mail

	default:
		err := logViaGRPC(payload)
		if err != nil {
			log.Println(err)
		}
	}
}

// Logs to MongoDB via gRPC
func logViaGRPC(entry PayLoad) error {
	conn, err := grpc.Dial("logger-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return err
	}

	defer conn.Close()

	c := logs.NewLogServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: entry.Name,
			Data: entry.Data.(string),
		},
	})

	if err != nil {
		return err
	}

	return nil
}
