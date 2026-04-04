// Package queue manages RabbitMQ publishing and consuming.
//
// Exchange topolojisi:
//   - transactions (fanout): POST /transactions gelince buraya yazılır.
//   - events       (fanout): her işlem analiz edildikten sonra buraya yazılır.
//
// Mesaj akışı:
//
//	Publisher → transactions exchange → transactions.process queue → Consumer (fraud worker)
//	Fraud worker → events exchange → events.<uuid> queue → Consumer (API server → WebSocket)
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
	exchangeEvents       = "events"
	queueTransactions    = "transactions.process"
)

// TransactionMessage, queue üzerinde taşınan mesaj tipidir.
// models.Transaction bson tag'leri kullandığından JSON için ayrı bir tip tanımlanır.
type TransactionMessage struct {
	ID           string                   `json:"id"`
	UserID       string                   `json:"user_id"`
	Amount       float64                  `json:"amount"`
	Lat          float64                  `json:"lat"`
	Lon          float64                  `json:"lon"`
	Status       models.TransactionStatus `json:"status"`
	CreatedAt    string                   `json:"created_at"`
	FraudReasons []models.FraudReason     `json:"fraud_reasons,omitempty"`
}

func newMessage(tx models.Transaction) TransactionMessage {
	return TransactionMessage{
		ID:           tx.ID.Hex(),
		UserID:       tx.UserID,
		Amount:       tx.Amount,
		Lat:          tx.Lat,
		Lon:          tx.Lon,
		Status:       tx.Status,
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
	for _, name := range []string{exchangeTransactions, exchangeEvents} {
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

// TransactionEvent, her işlem analiz edildikten sonra events exchange'e gönderilen mesajdır.
type TransactionEvent struct {
	TransactionID string                   `json:"transaction_id"`
	UserID        string                   `json:"user_id"`
	Status        models.TransactionStatus `json:"status"`
	Amount        float64                  `json:"amount"`
	FraudReasons  []models.FraudReason     `json:"fraud_reasons"`
	CreatedAt     string                   `json:"created_at"`
}

// PublishEvent, bir TransactionEvent'i events exchange'ine gönderir.
func (c *Client) PublishEvent(ctx context.Context, event TransactionEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	return c.ch.PublishWithContext(ctx,
		exchangeEvents,
		"",    // fanout — routing key kullanılmaz
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

// ConsumeEvents, events fanout exchange'ini dinler.
// Her server instance'ı için exclusive + auto-delete queue açılır;
// böylece tüm instance'lar aynı event'leri alır (broadcast semantiği).
func (c *Client) ConsumeEvents(ctx context.Context, fn func(TransactionEvent) error) error {
	// İsim boş → RabbitMQ benzersiz isim üretir.
	// autoDelete=true, exclusive=true → bağlantı kopunca queue silinir.
	q, err := c.ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		return fmt.Errorf("declare events queue: %w", err)
	}

	if err := c.ch.QueueBind(q.Name, "", exchangeEvents, false, nil); err != nil {
		return fmt.Errorf("bind events queue: %w", err)
	}

	// exclusive queue → autoAck=true kullanmak güvenli (başka consumer yok)
	msgs, err := c.ch.Consume(q.Name, "", true, true, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume events: %w", err)
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
				var e TransactionEvent
				if err := json.Unmarshal(msg.Body, &e); err != nil {
					log.Printf("[queue] event unmarshal hatası: %v", err)
					continue
				}
				if err := fn(e); err != nil {
					log.Printf("[queue] event işlem hatası: %v", err)
				}
			}
		}
	}()

	return nil
}

func (c *Client) Close() error {
	c.ch.Close()
	return c.conn.Close()
}
