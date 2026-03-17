package queue

import (
	"encoding/json"
	"fmt"
	"log"

	"viperai/internal/config"
	"viperai/internal/domain"

	"github.com/streadway/amqp"
)

var connection *amqp.Connection

type MessageQueue struct {
	channel  *amqp.Channel
	exchange string
	key      string
}

func Initialize() error {
	cfg := config.Get().Queue

	url := fmt.Sprintf(
		"amqp://%s:%s@%s:%d/%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.VHost,
	)

	var err error
	connection, err = amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return nil
}

func NewMessageQueue(queueName string) *MessageQueue {
	if connection == nil {
		if err := Initialize(); err != nil {
			log.Fatalf("RabbitMQ initialization failed: %v", err)
		}
	}

	ch, err := connection.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}

	return &MessageQueue{
		channel:  ch,
		exchange: "",
		key:      queueName,
	}
}

func (mq *MessageQueue) Publish(data []byte) error {
	_, err := mq.channel.QueueDeclare(mq.key, false, false, false, false, nil)
	if err != nil {
		return err
	}

	return mq.channel.Publish(mq.exchange, mq.key, false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
		},
	)
}

func (mq *MessageQueue) Consume(handler func(*amqp.Delivery) error) {
	q, err := mq.channel.QueueDeclare(mq.key, false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	msgs, err := mq.channel.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	for msg := range msgs {
		if err := handler(&msg); err != nil {
			log.Printf("Message handling error: %v", err)
		}
	}
}

func (mq *MessageQueue) Close() {
	mq.channel.Close()
}

type MessagePayload struct {
	ConversationID string `json:"conversation_id"`
	Content        string `json:"content"`
	UserID         int64  `json:"user_id"`
	IsFromUser     bool   `json:"is_from_user"`
}

func EncodeMessage(msg *domain.ChatMessage) []byte {
	payload := MessagePayload{
		ConversationID: msg.ConversationID,
		Content:        msg.Content,
		UserID:         msg.UserID,
		IsFromUser:     msg.IsFromUser,
	}
	data, _ := json.Marshal(payload)
	return data
}

func DecodeMessage(data []byte) (*MessagePayload, error) {
	var payload MessagePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

var MessageQueueInstance *MessageQueue

func StartConsumer(saveFunc func(*domain.ChatMessage) error) {
	MessageQueueInstance = NewMessageQueue("chat_messages")
	go MessageQueueInstance.Consume(func(delivery *amqp.Delivery) error {
		payload, err := DecodeMessage(delivery.Body)
		if err != nil {
			return err
		}

		msg := &domain.ChatMessage{
			ConversationID: payload.ConversationID,
			Content:        payload.Content,
			UserID:         payload.UserID,
			IsFromUser:     payload.IsFromUser,
		}

		return saveFunc(msg)
	})
}
