package kafka

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"order-service/internal/domain"

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

// [magic byte (0)] + [4-byte schema ID] + [Avro payload]
func (p *OrderEventProducer) SendOrderEvent(event domain.OrderEvent) error {
	native := map[string]interface{}{
		"order_id":   event.OrderID,
		"event_type": event.EventType,
		"timestamp":  event.Timestamp.UnixMilli(),
	}

	avroPayload, err := p.codec.BinaryFromNative(nil, native)
	if err != nil {
		return fmt.Errorf("failed to encode avro message: %w", err)
	}

	fmt.Printf("Avro payload size: %d bytes\n", len(avroPayload))

	var buf bytes.Buffer
	buf.WriteByte(0)
	if err := binary.Write(&buf, binary.BigEndian, uint32(p.schemaID)); err != nil {
		return fmt.Errorf("failed to write schema id: %w", err)
	}
	buf.Write(avroPayload)

	totalMessage := buf.Bytes()
	fmt.Printf("Total message size: %d bytes\n", len(totalMessage))

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(event.OrderID),
		Value: sarama.ByteEncoder(totalMessage),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	fmt.Printf("Message sent to partition %d, offset %d\n", partition, offset)
	return nil
}

func (p *OrderEventProducer) Close() error {
	return p.producer.Close()
}
