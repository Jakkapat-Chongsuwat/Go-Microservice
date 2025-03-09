package kafka

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"order-service/internal/adapters/models"

	"github.com/IBM/sarama"
	"github.com/linkedin/goavro/v2"
	"github.com/riferrei/srclient"
)

type OrderEventProducer struct {
	producer sarama.SyncProducer
	topic    string
	srClient *srclient.SchemaRegistryClient
	codec    *goavro.Codec
	schemaID int
}

func NewOrderEventProducer(brokers []string, topic, schemaRegistryURL, subject, schemaStr string) (*OrderEventProducer, error) {
	srClient := srclient.CreateSchemaRegistryClient(schemaRegistryURL)

	registeredSchema, err := srClient.CreateSchema(subject, schemaStr, srclient.Avro)
	if err != nil {
		return nil, fmt.Errorf("failed to register schema: %w", err)
	}
	schemaID := registeredSchema.ID()

	codec, err := goavro.NewCodec(schemaStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create avro codec: %w", err)
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Version = sarama.V4_0_0_0

	prod, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	return &OrderEventProducer{
		producer: prod,
		topic:    topic,
		srClient: srClient,
		codec:    codec,
		schemaID: schemaID,
	}, nil
}

// Message format: [magic byte (0)] + [4-byte schema ID] + [Avro payload]
func (p *OrderEventProducer) SendOrderEvent(event models.OrderEvent) error {
	native := map[string]interface{}{
		"order_id":   event.OrderID,
		"event_type": event.EventType,
		"timestamp":  event.Timestamp.UnixMilli(),
	}

	avroPayload, err := p.codec.BinaryFromNative(nil, native)
	if err != nil {
		return fmt.Errorf("failed to encode avro message: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteByte(0)
	if err := binary.Write(&buf, binary.BigEndian, int32(p.schemaID)); err != nil {
		return fmt.Errorf("failed to write schema id: %w", err)
	}
	buf.Write(avroPayload)

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(event.OrderID),
		Value: sarama.ByteEncoder(buf.Bytes()),
	}

	_, _, err = p.producer.SendMessage(msg)
	return err
}

func (p *OrderEventProducer) Close() error {
	return p.producer.Close()
}
