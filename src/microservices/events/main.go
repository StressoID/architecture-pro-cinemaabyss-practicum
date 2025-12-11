package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
)

type EventsService struct {
	producer sarama.SyncProducer
	consumer sarama.ConsumerGroup
	brokers  []string
}

type MovieEvent struct {
	MovieID    int       `json:"movie_id"`
	Title      string    `json:"title"`
	Action     string    `json:"action"`
	UserID     *int      `json:"user_id,omitempty"`
	Rating     *float64  `json:"rating,omitempty"`
	Genres     []string  `json:"genres,omitempty"`
	Description *string  `json:"description,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

type UserEvent struct {
	UserID    int       `json:"user_id"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
	Username  *string   `json:"username,omitempty"`
	Email     *string   `json:"email,omitempty"`
}

type PaymentEvent struct {
	PaymentID int       `json:"payment_id"`
	UserID    int       `json:"user_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	MethodType *string  `json:"method_type,omitempty"`
}

type EventResponse struct {
	Status   string      `json:"status"`
	Partition int32      `json:"partition"`
	Offset   int64       `json:"offset"`
	Event    interface{} `json:"event"`
}

type ConsumerGroupHandler struct{}

func (ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (h ConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}
			log.Printf("Received event from topic %s, partition %d, offset %d: %s",
				message.Topic, message.Partition, message.Offset, string(message.Value))
			sess.MarkMessage(message, "")
		case <-sess.Context().Done():
			return nil
		}
	}
}

func NewEventsService() (*EventsService, error) {
	brokersStr := os.Getenv("KAFKA_BROKERS")
	if brokersStr == "" {
		brokersStr = "kafka:9092"
	}
	brokers := []string{brokersStr}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Version = sarama.V2_7_0_0

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	consumerConfig := sarama.NewConfig()
	consumerConfig.Version = sarama.V2_7_0_0
	consumerConfig.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	consumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumerGroup(brokers, "events-service-group", consumerConfig)
	if err != nil {
		producer.Close()
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	return &EventsService{
		producer: producer,
		consumer: consumer,
		brokers:  brokers,
	}, nil
}

func (es *EventsService) publishEvent(topic string, event interface{}) (*EventResponse, error) {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(eventJSON),
	}

	partition, offset, err := es.producer.SendMessage(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("Published event to topic %s, partition %d, offset %d", topic, partition, offset)

	return &EventResponse{
		Status:   "success",
		Partition: partition,
		Offset:   offset,
		Event:    event,
	}, nil
}

func (es *EventsService) handleMovieEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event MovieEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	event.Timestamp = time.Now()

	response, err := es.publishEvent("movie-events", event)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to publish event: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (es *EventsService) handleUserEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event UserEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	response, err := es.publishEvent("user-events", event)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to publish event: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (es *EventsService) handlePaymentEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event PaymentEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	response, err := es.publishEvent("payment-events", event)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to publish event: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (es *EventsService) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"status": true})
}

func (es *EventsService) startConsumer(ctx context.Context) {
	handler := ConsumerGroupHandler{}
	topics := []string{"movie-events", "user-events", "payment-events"}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := es.consumer.Consume(ctx, topics, handler); err != nil {
				log.Printf("Error from consumer: %v", err)
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func main() {
	service, err := NewEventsService()
	if err != nil {
		log.Fatalf("Failed to create events service: %v", err)
	}
	defer service.producer.Close()
	defer service.consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		service.startConsumer(ctx)
	}()

	http.HandleFunc("/api/events/health", service.handleHealth)
	http.HandleFunc("/api/events/movie", service.handleMovieEvent)
	http.HandleFunc("/api/events/user", service.handleUserEvent)
	http.HandleFunc("/api/events/payment", service.handlePaymentEvent)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: nil,
	}

	go func() {
		log.Printf("Starting events service on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm

	log.Println("Shutting down...")
	cancel()
	wg.Wait()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}
}


