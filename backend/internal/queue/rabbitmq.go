// Package queue manages RabbitMQ publishing and consuming.
//
// Exchange topolojisi:
//   - transactions (fanout): POST /transactions gelince buraya yazılır.
//   - alerts       (fanout): fraud tespit edilince buraya yazılır (ileride).
//
// Mesaj akışı:
//   Publisher → transactions exchange → transactions.process queue → Consumer (fraud worker)
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"fraud-detection/internal/models"
)

const (
	exchangeTransactions = "transactions"
	exchangeAlerts       = "alerts"
	queueTransactions    = "transactions.process"
)

// TransactionMessage, queue üzerinde taşınan mesaj tipidir.
// models.Transaction bson tag'leri kullandığından JSON için ayrı bir tip tanımlanır.
type TransactionMessage struct {
	ID           string   `json:"id"`
	UserID       string   `json:"user_id"`
	Amount       float64  `json:"amount"`
	Lat          float64  `json:"lat"`
	Lon          float64  `json:"lon"`
	Status       string   `json:"status"`
	CreatedAt    string   `json:"created_at"`
	FraudReasons []string `json:"fraud_reasons,omitempty"`
}

func newMessage(tx models.Transaction) TransactionMessage {
	return TransactionMessage{
		ID:           tx.ID.Hex(),
		UserID:       tx.UserID,
		Amount:       tx.Amount,
		Lat:          tx.Lat,
		Lon:          tx.Lon,
		Status:       tx.Status.String(),
		CreatedAt:    tx.CreatedAt.Format(time.RFC3339),
		FraudReasons: tx.FraudReasons,
	}
}

// Client, tek bir RabbitMQ bağlantısını ve channel'ını yönetir.
// Bağlantı koparsa process yeniden başlatılır (orchestrator sorumluluğu).
type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// New bağlantıyı kurar ve exchange'leri declare eder.
func New(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}

	// Exchange'ler idempotent olarak declare edilir — varsa olduğu gibi kalır.
	// durable=true → RabbitMQ restart olsa bile exchange kaybolmaz.
	// fanout → routing key görmezden gelir, bağlı tüm queue'lara gönderir.
	for _, name := range []string{exchangeTransactions, exchangeAlerts} {
		if err := ch.ExchangeDeclare(name, "fanout", true, false, false, false, nil); err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("declare exchange %q: %w", name, err)
		}
	}

	return &Client{conn: conn, ch: ch}, nil
}

// PublishTransaction, bir transaction'ı JSON olarak transactions exchange'ine gönderir.
// Exchange fanout olduğu için routing key boş string geçilir (kullanılmaz).
// DeliveryMode=Persistent → mesaj RabbitMQ restart olsa bile kaybolmaz (disk'e yazılır).
func (c *Client) PublishTransaction(ctx context.Context, tx models.Transaction) error {
	body, err := json.Marshal(newMessage(tx))
	if err != nil {
		return fmt.Errorf("marshal transaction: %w", err)
	}

	return c.ch.PublishWithContext(ctx,
		exchangeTransactions,
		"",    // routing key (fanout'ta önemsiz)
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

// ConsumeTransactions, transactions.process queue'sunu dinler ve her mesaj için fn'i çağırır.
// Queue, transactions exchange'ine bind edilir — bu olmadan exchange mesajları nereye
// göndereceğini bilmez ve mesajlar kaybolur.
//
// fn hata dönerse mesaj Nack edilip yeniden kuyruğa alınır (requeue=true).
// fn başarılı olursa Ack edilir ve RabbitMQ mesajı siler.
func (c *Client) ConsumeTransactions(ctx context.Context, fn func(TransactionMessage) error) error {
	// Queue declare: durable=true → restart'ta kaybolmaz.
	// exclusive=false → birden fazla consumer bağlanabilir.
	q, err := c.ch.QueueDeclare(queueTransactions, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}

	// Bind: "transactions exchange'den gelen mesajları bu queue'ya ver"
	if err := c.ch.QueueBind(q.Name, "", exchangeTransactions, false, nil); err != nil {
		return fmt.Errorf("bind queue: %w", err)
	}

	// autoAck=false → mesajı işleyene kadar RabbitMQ "beklemede" sayar.
	// Uygulama çökerse mesaj başka consumer'a (veya yeniden başlayınca aynısına) iletilir.
	msgs, err := c.ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}

				var m TransactionMessage
				if err := json.Unmarshal(msg.Body, &m); err != nil {
					log.Printf("[queue] unmarshal hatası: %v — mesaj atılıyor", err)
					msg.Nack(false, false) // requeue=false: bozuk mesajı tekrar alma
					continue
				}

				if err := fn(m); err != nil {
					log.Printf("[queue] işlem hatası: %v — mesaj yeniden kuyruğa alınıyor", err)
					msg.Nack(false, true) // requeue=true
					continue
				}

				msg.Ack(false) // false = sadece bu mesajı onayla (toplu değil)
			}
		}
	}()

	return nil
}

func (c *Client) Close() error {
	c.ch.Close()
	return c.conn.Close()
}
