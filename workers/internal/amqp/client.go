package amqp

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeMessages = "messages"
)

func Dial(url string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("amqp dial: %w", err)
	}
	return conn, nil
}

func OpenChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("amqp open channel: %w", err)
	}
	return ch, nil
}

// DeclareExchange declares a durable topic exchange.
func DeclareExchange(ch *amqp.Channel, name, kind string) error {
	if err := ch.ExchangeDeclare(name, kind, true, false, false, false, nil); err != nil {
		return fmt.Errorf("amqp declare exchange %q: %w", name, err)
	}
	return nil
}

func DeclareQueue(ch *amqp.Channel, queue string) error {
	_, err := ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("amqp declare queue %q: %w", queue, err)
	}
	return nil
}

func DeclareQueueWithArgs(ch *amqp.Channel, queue string, args amqp.Table) error {
	_, err := ch.QueueDeclare(queue, true, false, false, false, args)
	if err != nil {
		return fmt.Errorf("amqp declare queue %q: %w", queue, err)
	}
	return nil
}

func BindQueue(ch *amqp.Channel, queue, exchange, routingKey string) error {
	if err := ch.QueueBind(queue, routingKey, exchange, false, nil); err != nil {
		return fmt.Errorf("amqp bind queue %q to %q (%s): %w", queue, exchange, routingKey, err)
	}
	return nil
}

// SetQoS limits unacked deliveries pushed to this consumer (prefetchCount=1 == one-at-a-time)
func SetQoS(ch *amqp.Channel, prefetchCount int) error {
	if err := ch.Qos(prefetchCount, 0, false); err != nil {
		return fmt.Errorf("amqp qos: %w", err)
	}
	return nil
}

func ConsumeQueue(ch *amqp.Channel, queue, consumer string, autoAck bool) (<-chan amqp.Delivery, error) {
	msgs, err := ch.Consume(queue, consumer, autoAck, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("amqp consume queue %q: %w", queue, err)
	}
	return msgs, nil
}

func PublishMessage(ctx context.Context, ch *amqp.Channel, exchange, routingKey, contentType string, body []byte) error {
	err := ch.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
		ContentType:  contentType,
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
	if err != nil {
		return fmt.Errorf("amqp publish to %q (%s): %w", exchange, routingKey, err)
	}
	return nil
}

func PublishMessageWithHeaders(ctx context.Context, ch *amqp.Channel, exchange, routingKey, contentType string, body []byte, headers amqp.Table) error {
	err := ch.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
		ContentType:  contentType,
		DeliveryMode: amqp.Persistent,
		Body:         body,
		Headers:      headers,
	})
	if err != nil {
		return fmt.Errorf("amqp publish to %q (%s): %w", exchange, routingKey, err)
	}
	return nil
}
