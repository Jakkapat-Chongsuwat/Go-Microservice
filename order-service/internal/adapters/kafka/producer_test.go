package kafka_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"order-service/internal/adapters/kafka"
	"order-service/internal/adapters/models"
	"order-service/internal/domain"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/linkedin/goavro/v2"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcKafka "github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

type containerCustomizerFunc func(req *testcontainers.GenericContainerRequest) error

func (f containerCustomizerFunc) Customize(req *testcontainers.GenericContainerRequest) error {
	return f(req)
}

func TestProducerIntegration(t *testing.T) {
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()

	net, err := network.New(
		ctx,
		network.WithDriver("bridge"),
	)
	require.NoError(t, err)
	defer net.Remove(ctx)

	kafkaC, err := tcKafka.Run(
		ctx,
		"confluentinc/confluent-local:7.5.0",
		tcKafka.WithClusterID("test-cluster"),
		containerCustomizerFunc(func(req *testcontainers.GenericContainerRequest) error {
			req.Networks = append(req.Networks, net.Name)
			if req.NetworkAliases == nil {
				req.NetworkAliases = make(map[string][]string)
			}
			req.NetworkAliases[net.Name] = []string{"kafka"}
			return nil
		}),
	)
	require.NoError(t, err)
	defer kafkaC.Terminate(ctx)

	brokers, err := kafkaC.Brokers(ctx)
	require.NoError(t, err)

	logReader, err := kafkaC.Logs(ctx)
	if err == nil {
		defer logReader.Close()
		buf := new(bytes.Buffer)
		_, _ = io.Copy(buf, logReader)
		logger.Info("Kafka container logs", zap.String("logs", buf.String()))
	}

	srReq := testcontainers.ContainerRequest{
		Image:        "confluentinc/cp-schema-registry:7.9.0",
		ExposedPorts: []string{"8081/tcp"},
		Env: map[string]string{
			"SCHEMA_REGISTRY_HOST_NAME":                    "schemaregistry",
			"SCHEMA_REGISTRY_LISTENERS":                    "http://0.0.0.0:8081",
			"SCHEMA_REGISTRY_KAFKASTORE_BOOTSTRAP_SERVERS": "PLAINTEXT://kafka:9092",
		},
		WaitingFor: wait.ForHTTP("/subjects").
			WithPort("8081/tcp").
			WithStartupTimeout(120 * time.Second),
		Networks: []string{net.Name},
		NetworkAliases: map[string][]string{
			net.Name: {"schemaregistry"},
		},
	}

	srContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: srReq,
		Started:          true,
	})
	require.NoError(t, err)
	defer srContainer.Terminate(ctx)

	srHost, err := srContainer.Host(ctx)
	require.NoError(t, err)
	srPort, err := srContainer.MappedPort(ctx, "8081")
	require.NoError(t, err)
	schemaRegistryURL := fmt.Sprintf("http://%s:%s", srHost, srPort.Port())

	models.LoadSchema("../../../configs/order_event_schema.json")
	schemaStr := models.OrderEventSchema

	subject := "order-events-value"
	topic := "order-events"

	producer, err := kafka.NewOrderEventProducer(brokers, topic, schemaRegistryURL, subject, schemaStr)
	require.NoError(t, err)
	defer producer.Close()

	event := domain.OrderEvent{
		OrderID:   "test-order-1",
		EventType: "CREATED",
		Timestamp: time.Now(),
	}
	err = producer.SendOrderEvent(event)
	require.NoError(t, err)

	consumer, err := sarama.NewConsumer(brokers, nil)
	require.NoError(t, err)
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	require.NoError(t, err)
	defer partitionConsumer.Close()

	select {
	case msg := <-partitionConsumer.Messages():
		data := msg.Value
		require.GreaterOrEqual(t, len(data), 5, "message too short")
		require.Equal(t, byte(0), data[0])
		schemaID := int(binary.BigEndian.Uint32(data[1:5]))
		require.Greater(t, schemaID, 0, "invalid schema id")

		codec, err := goavro.NewCodec(schemaStr)
		require.NoError(t, err)
		native, _, err := codec.NativeFromBinary(data[5:])
		require.NoError(t, err)
		nativeMap, ok := native.(map[string]interface{})
		require.True(t, ok, "decoded native is not a map")

		var ts int64
		switch v := nativeMap["timestamp"].(type) {
		case int64:
			ts = v
		case float64:
			ts = int64(v)
		case time.Time:
			ts = v.UnixMilli()
		default:
			require.Fail(t, fmt.Sprintf("timestamp is not numeric, got %T", nativeMap["timestamp"]))
		}

		require.Equal(t, event.OrderID, nativeMap["order_id"])
		require.Equal(t, event.EventType, nativeMap["event_type"])
		require.InDelta(t, event.Timestamp.UnixMilli(), ts, 1000, "timestamp mismatch")
	case <-time.After(180 * time.Second):
		t.Fatal("timeout waiting for produced message")
	}
}
