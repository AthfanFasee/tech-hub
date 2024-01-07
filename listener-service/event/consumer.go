package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	logs "github.com/AthfanFasee/listener/proto/logs"
	mail "github.com/AthfanFasee/listener/proto/mail"
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

// Sets up a channel
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

// Listens to events and reatcs accordingly
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
	// Whenever a message is sent to the "tech_hub" exchange with a topic that matches any in the topics slice,
	// it'll be routed to the queue represented by q.Name
	for _, s := range topics {
		ch.QueueBind(
			q.Name,     // Queue(random) Name
			s,          // Topic name
			"tech_hub", // Exchange Name
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

	fmt.Printf("Waiting for message [Exchange, Queue] [tech_hub, %s]\n", q.Name)

	<-forever

	return nil
}

// Handls payload and reacts according to payload.Name
func handlePayload(payload PayLoad) {
	switch payload.Name {
	case "log", "error":
		err := logViaGRPC(payload)
		if err != nil {
			log.Println(err)
		}

	case "mail":
		err := SendMailViaGRPC(payload)
		if err != nil {
			log.Println(err)
		}

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

// Sends welcome mail via gRPC
func SendMailViaGRPC(entry PayLoad) error {
	conn, err := grpc.Dial("mail-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return err
	}

	defer conn.Close()

	c := mail.NewMailServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.SendWelcomeMail(ctx, &mail.SendWelcomeMailRequest{
		MailEntry: entry.Data.(*mail.Mail),
	})

	if err != nil {
		return err
	}

	return nil
}
