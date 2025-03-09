package kafka

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"notification-service/internal/domain"

	"github.com/IBM/sarama"
	"github.com/linkedin/goavro/v2"
	"github.com/riferrei/srclient"
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

type fakeNotificationUseCase struct {
	processed []*domain.Notification
	mu        sync.Mutex
}

func (f *fakeNotificationUseCase) ProcessNotification(ctx context.Context, notif *domain.Notification) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.processed = append(f.processed, notif)
	return nil
}

func TestConsumerIntegration(t *testing.T) {
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()

	net, err := network.New(ctx, network.WithDriver("bridge"))
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
		_, err := io.Copy(buf, logReader)
		if err == nil {
			logger.Info("Kafka container logs", zap.String("logs", buf.String()))
		}
	}

	srReq := testcontainers.ContainerRequest{
		Image:        "confluentinc/cp-schema-registry:7.9.0",
		ExposedPorts: []string{"8085/tcp"},
		Env: map[string]string{
			"SCHEMA_REGISTRY_HOST_NAME":                    "schemaregistry",
			"SCHEMA_REGISTRY_LISTENERS":                    "http://0.0.0.0:8085",
			"SCHEMA_REGISTRY_KAFKASTORE_BOOTSTRAP_SERVERS": "PLAINTEXT://kafka:9092",
		},
		WaitingFor: wait.ForHTTP("/subjects").
			WithPort("8085/tcp").
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
	srPort, err := srContainer.MappedPort(ctx, "8085")
	require.NoError(t, err)
	schemaRegistryURL := fmt.Sprintf("http://%s:%s", srHost, srPort.Port())

	fakeUC := &fakeNotificationUseCase{}

	consumerGroup, err := NewKafkaConsumerGroup(brokers, "test-group", "test-topic", schemaRegistryURL, fakeUC, logger)
	require.NoError(t, err)

	srClient := srclient.NewSchemaRegistryClient(schemaRegistryURL)
	subject := "test-topic-value"
	schemaStr := `
    {
      "type": "record",
      "name": "Notification",
      "namespace": "com.example.notification",
      "fields": [
        { "name": "id", "type": "string" },
        { "name": "type", "type": "string" },
        { "name": "message", "type": "string" }
      ]
    }`
	registeredSchema, err := srClient.CreateSchema(subject, schemaStr, srclient.Avro)
	require.NoError(t, err)
	schemaID := registeredSchema.ID()

	codec, err := goavro.NewCodec(schemaStr)
	require.NoError(t, err)

	native := map[string]interface{}{
		"id":      "123",
		"type":    "test",
		"message": "hello world",
	}
	binaryData, err := codec.BinaryFromNative(nil, native)
	require.NoError(t, err)

	var b bytes.Buffer
	b.WriteByte(0)
	err = binary.Write(&b, binary.BigEndian, int32(schemaID))
	require.NoError(t, err)
	b.Write(binaryData)
	messageBytes := b.Bytes()

	producer, err := sarama.NewSyncProducer(brokers, nil)
	require.NoError(t, err)
	defer producer.Close()

	prodMsg := &sarama.ProducerMessage{
		Topic: "test-topic",
		Key:   sarama.StringEncoder("key1"),
		Value: sarama.ByteEncoder(messageBytes),
	}
	partition, offset, err := producer.SendMessage(prodMsg)
	require.NoError(t, err)
	logger.Info("Produced test message", zap.Int32("partition", partition), zap.Int64("offset", offset))

	ctxConsumer, cancelConsumer := context.WithTimeout(ctx, 60*time.Second)
	defer cancelConsumer()

	var consumerErr error
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumerErr = consumerGroup.Start(ctxConsumer)
	}()
	wg.Wait()
	require.NoError(t, consumerErr)

	require.Len(t, fakeUC.processed, 1)
	processedNotif := fakeUC.processed[0]
	require.Equal(t, "123", processedNotif.ID)
	require.Equal(t, "test", processedNotif.Type)
	require.Equal(t, "hello world", processedNotif.Message)
}
