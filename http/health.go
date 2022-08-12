package http

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/segmentio/kafka-go"
)

var (
	kafkaURLs   = []string{"localhost:9092"}
	healthTopic = "health"
	once        sync.Once
)

func (s *Server) handleCheckHealth(c *fiber.Ctx) error {
	// Setup a simple kafka worker
	once.Do(func() {
		s.ReadMessages(healthTopic)
	})

	s.WriteMessage(healthTopic, []byte(fmt.Sprintf("Health checked at %s", time.Now())))

	return c.SendString("OK")
}

func (s *Server) ReadMessages(topic string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  kafkaURLs,
		GroupID:  "example",
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	go func() {
		for {
			msg, err := r.ReadMessage(context.TODO())
			if err != nil {
				s.Core.Logger.Error(err)
				break
			}
			s.Core.Logger.Info(string(msg.Value))
		}

		err := r.Close()
		if err != nil {
			s.Core.Logger.Error(err)
		}
	}()
}

func (s *Server) WriteMessage(topic string, message []byte) error {
	w := &kafka.Writer{
		Addr:     kafka.TCP(kafkaURLs[0]),
		Topic:    healthTopic,
		Balancer: &kafka.LeastBytes{},
	}

	err := w.WriteMessages(context.Background(),
		kafka.Message{
			Value: message,
		},
	)
	defer w.Close()

	if err != nil {
		s.Core.Logger.Error(err)
	}

	s.Core.Logger.Infof("Message is stored in topic(%s)", topic)

	return nil
}
