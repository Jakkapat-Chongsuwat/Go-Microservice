package kafka

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"notification-service/internal/domain"
	"notification-service/internal/usecases"

	"github.com/IBM/sarama"
	"github.com/linkedin/goavro/v2"
	"github.com/riferrei/srclient"
	"go.uber.org/zap"
)

type KafkaConsumerGroup struct {
	group    sarama.ConsumerGroup
	topic    string
	useCase  usecases.NotificationUseCase
	logger   *zap.Logger
	srClient *srclient.SchemaRegistryClient
}

func NewKafkaConsumerGroup(brokers []string, groupID, topic, schemaRegistryURL string, useCase usecases.NotificationUseCase, logger *zap.Logger) (*KafkaConsumerGroup, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V4_0_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	group, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	srClient := srclient.CreateSchemaRegistryClient(schemaRegistryURL)

	return &KafkaConsumerGroup{
		group:    group,
		topic:    topic,
		useCase:  useCase,
		logger:   logger,
		srClient: srClient,
	}, nil
}

func (kc *KafkaConsumerGroup) Start(ctx context.Context) error {
	consumer := consumerGroupHandler{
		useCase:  kc.useCase,
		logger:   kc.logger,
		srClient: kc.srClient,
		topic:    kc.topic,
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			if err := kc.group.Consume(ctx, []string{kc.topic}, &consumer); err != nil {
				kc.logger.Error("error during consuming", zap.Error(err))
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		kc.logger.Info("context cancelled")
	case <-sigterm:
		kc.logger.Info("termination signal received")
	}

	if err := kc.group.Close(); err != nil {
		kc.logger.Error("error closing consumer group", zap.Error(err))
	}
	wg.Wait()
	return nil
}

type consumerGroupHandler struct {
	useCase  usecases.NotificationUseCase
	logger   *zap.Logger
	srClient *srclient.SchemaRegistryClient
	topic    string
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	h.logger.Info("consumer group session setup")
	return nil
}

func (h *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	h.logger.Info("consumer group session cleanup")
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.logger.Info("message received",
			zap.String("key", string(msg.Key)),
			zap.String("raw_value", string(msg.Value)),
			zap.Int32("partition", msg.Partition),
			zap.Int64("offset", msg.Offset))

		notifMap, err := decodeAvroMessage(h.srClient, msg.Value)
		if err != nil {
			h.logger.Error("failed to decode avro message", zap.Error(err))
			session.MarkMessage(msg, "")
			continue
		}

		notif := domain.NewNotificationWithID(fmt.Sprintf("%v", notifMap["id"]), fmt.Sprintf("%v", notifMap["type"]), fmt.Sprintf("%v", notifMap["message"]))

		if err := h.useCase.ProcessNotification(session.Context(), notif); err != nil {
			h.logger.Error("failed to process notification", zap.Error(err))
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

// Format: [magic byte (0)] + [4-byte schema ID] + [Avro payload].
func decodeAvroMessage(srClient *srclient.SchemaRegistryClient, data []byte) (map[string]interface{}, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("data too short")
	}
	if data[0] != 0 {
		return nil, fmt.Errorf("unknown magic byte: %v", data[0])
	}
	schemaID := int(binary.BigEndian.Uint32(data[1:5]))
	schema, err := srClient.GetSchema(schemaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema for id %d: %w", schemaID, err)
	}

	codec, err := goavro.NewCodec(schema.Schema())
	if err != nil {
		return nil, fmt.Errorf("failed to create codec: %w", err)
	}

	native, _, err := codec.NativeFromBinary(data[5:])
	if err != nil {
		return nil, fmt.Errorf("failed to decode avro payload: %w", err)
	}

	result, ok := native.(map[string]interface{})
	if !ok {
		raw, err := codec.TextualFromNative(nil, native)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to textual: %w", err)
		}
		var m map[string]interface{}
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, fmt.Errorf("failed to unmarshal textual json: %w", err)
		}
		return m, nil
	}
	return result, nil
}
