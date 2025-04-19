package event

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaEventBus struct {
	writer *kafka.Writer
}

func NewKafkaEventBus(brokers []string) (*KafkaEventBus, error) {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Balancer: &kafka.LeastBytes{},
	}
	return &KafkaEventBus{writer: writer}, nil
}

func (k *KafkaEventBus) Publish(ctx context.Context, topic, key string, value []byte) error {
	return k.writer.WriteMessages(ctx,
		kafka.Message{
			Topic: topic,
			Key:   []byte(key),
			Value: value,
			Time:  time.Now(),
		},
	)
}

func (k *KafkaEventBus) Close() error {
	return k.writer.Close()
}

func NewKafkaReader(brokers []string, topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})

}

// func ensureTopicExists(admin *kafka.AdminClient, topic string) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	_, err := admin.CreateTopics(ctx, []kafka.TopicSpecification{{
// 		Topic:             topic,
// 		NumPartitions:     3,
// 		ReplicationFactor: 1,
// 		ConfigEntries:     map[string]string{"cleanup.policy": "compact"},
// 	}})

// 	if err != nil && !errors.Is(err, kafka.ErrTopicAlreadyExists) {
// 		return fmt.Errorf("failed to create topic %s: %w", topic, err)
// 	}
// 	return nil
// }
