// Package kafka provides a Gnomock Preset for Kafka.
//
// This preset cannot be used in parallel tests due to Kafka's port binding
// limitations. See https://github.com/orlangure/gnomock/issues/1038 for more
// details.
package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/registry"
	"github.com/segmentio/kafka-go"
)

// The following ports are exposed by this preset:.
const (
	BrokerPort         = "broker"
	ZooKeeperPort      = "zookeeper"
	WebPort            = "web"
	SchemaRegistryPort = "registry"
)

const (
	defaultVersion     = "3.6.1-L0"
	brokerPort         = 49092
	zookeeperPort      = 2181
	webPort            = 3030
	schemaRegistryPort = 8081
)

// Message is a single message sent to Kafka.
type Message struct {
	Topic string `json:"topic"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Time  int64  `json:"time"`
}

func init() {
	registry.Register("kafka", func() gnomock.Preset { return &P{} })
}

// Preset creates a new Gmomock Kafka preset. This preset includes a
// Kafka specific healthcheck function and default Kafka image and ports.
//
// Kafka preset uses a constant broker port number (49092) instead of
// allocating a random unoccupied port on every run. Please make sure this port
// is available when using this preset.
//
// By default, this preset uses `lensesio/fast-data-dev` docker image with
// version `2.5.1-L0` (version can be changed using `WithVersion`).
func Preset(opts ...Option) gnomock.Preset {
	p := &P{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

type TopicConfig struct {
	Topic         string
	NumPartitions int
}

// P is a Gnomock Preset implementation of Kafka.
type P struct {
	Version           string    `json:"version"`
	Topics            []string  `json:"topics"`
	Messages          []Message `json:"messages"`
	MessagesFiles     []string  `json:"messages_files"`
	UseSchemaRegistry bool      `json:"use_schema_registry"`

	TopicConfigs []TopicConfig `json:"topic_configs"`
}

// Image returns an image that should be pulled to create this container.
func (p *P) Image() string {
	return fmt.Sprintf("docker.io/lensesio/fast-data-dev:%s", p.Version)
}

// Ports returns ports that should be used to access this container.
func (p *P) Ports() gnomock.NamedPorts {
	namedPorts := make(gnomock.NamedPorts, 3)

	bp := gnomock.TCP(brokerPort)
	bp.HostPort = brokerPort
	namedPorts[BrokerPort] = bp

	namedPorts[ZooKeeperPort] = gnomock.TCP(zookeeperPort)
	namedPorts[WebPort] = gnomock.TCP(webPort)
	namedPorts[SchemaRegistryPort] = gnomock.TCP(schemaRegistryPort)

	return namedPorts
}

// Options returns a list of options to configure this container.
func (p *P) Options() []gnomock.Option {
	p.setDefaults()

	opts := []gnomock.Option{
		gnomock.WithHealthCheck(p.healthcheck),
		gnomock.WithEnv("KAFKA_AUTO_CREATE_TOPICS_ENABLE=true"),
		gnomock.WithEnv("ADV_HOST=127.0.0.1"),
		gnomock.WithEnv(fmt.Sprintf("BROKER_PORT=%d", brokerPort)),
		gnomock.WithEnv("RUNTESTS=0"),
		gnomock.WithEnv("RUNNING_SAMPLEDATA=0"),
		gnomock.WithEnv("SAMPLEDATA=0"),
	}

	if len(p.Topics) > 0 || len(p.TopicConfigs) > 0 || len(p.Messages) > 0 {
		opts = append(opts, gnomock.WithInit(p.initf))
	}

	return opts
}

func (p *P) healthcheck(ctx context.Context, c *gnomock.Container) (err error) {
	conn, err := p.connect(c)
	if err != nil {
		return fmt.Errorf("can't connect to kafka: %w", err)
	}

	defer func() {
		closeErr := conn.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	if _, err := conn.ApiVersions(); err != nil {
		return fmt.Errorf("can't get version info: %w", err)
	}

	if err := conn.CreateTopics(kafka.TopicConfig{
		Topic:             "gnomock",
		ReplicationFactor: 1,
		NumPartitions:     1,
	}); err != nil {
		return fmt.Errorf("can't create topic: %w", err)
	}

	group, err := kafka.NewConsumerGroup(kafka.ConsumerGroupConfig{
		ID:      "gnomock",
		Brokers: []string{c.Address(BrokerPort)},
		Topics:  []string{"gnomock"},
	})
	if err != nil {
		return fmt.Errorf("can't create consumer group: %w", err)
	}

	defer func() { _ = group.Close() }()

	if _, err := group.Next(ctx); err != nil {
		return fmt.Errorf("can't read next consumer group: %w", err)
	}

	if p.UseSchemaRegistry {
		if err := p.healthcheckRegistry(ctx, c); err != nil {
			return err
		}
	}

	return nil
}

func (p *P) healthcheckRegistry(ctx context.Context, c *gnomock.Container) error {
	url := "http://" + c.Address(SchemaRegistryPort)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("schema registry is not available: %w", err)
	}

	if err := res.Body.Close(); err != nil {
		return fmt.Errorf("error closing schema registry response body: %w", err)
	}

	return nil
}

func (p *P) setDefaults() {
	if p.Version == "" {
		p.Version = defaultVersion
	}
}

func (p *P) initf(ctx context.Context, c *gnomock.Container) (err error) {
	conn, err := p.connect(c)
	if err != nil {
		return fmt.Errorf("can't connect to kafka: %w", err)
	}

	defer func() {
		closeErr := conn.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	return p.ingestMessageFiles(ctx, c, conn)
}

func (p *P) ingestMessageFiles(ctx context.Context, c *gnomock.Container, conn *kafka.Conn) error {
	if len(p.MessagesFiles) > 0 {
		for _, fName := range p.MessagesFiles {
			msgs, err := p.loadMessagesFromFile(fName)
			if err != nil {
				return fmt.Errorf("can't read messages from file '%s': %w", fName, err)
			}

			p.Messages = append(p.Messages, msgs...)
		}
	}

	messagesByTopics := make(map[string][]Message)

	for _, m := range p.Messages {
		messagesByTopics[m.Topic] = append(messagesByTopics[m.Topic], m)
	}

	for topic := range messagesByTopics {
		p.Topics = append(p.Topics, topic)
	}

	topics := make([]kafka.TopicConfig, 0, len(p.Topics)+len(p.TopicConfigs))

	for _, topic := range p.Topics {
		topics = append(topics, kafka.TopicConfig{
			Topic:             topic,
			ReplicationFactor: 1,
			NumPartitions:     1,
		})
	}

	for _, topic := range p.TopicConfigs {
		topics = append(topics, kafka.TopicConfig{
			Topic:             topic.Topic,
			ReplicationFactor: 1, // cannot set more; cluster has just 1 node
			NumPartitions:     topic.NumPartitions,
		})
	}

	if err := conn.CreateTopics(topics...); err != nil {
		return fmt.Errorf("can't create topics: %w", err)
	}

	for topic, messages := range messagesByTopics {
		if err := p.sendMessagesIntoTopic(ctx, c.Address(BrokerPort), topic, messages); err != nil {
			return fmt.Errorf("can't send messages into topic '%s': %w", topic, err)
		}
	}

	return nil
}

// nolint:gosec
func (p *P) loadMessagesFromFile(fName string) (msgs []Message, err error) {
	f, err := os.Open(fName)
	if err != nil {
		return nil, fmt.Errorf("can't open messages file '%s': %w", fName, err)
	}

	defer func() {
		closeErr := f.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	decoder := json.NewDecoder(f)

	for {
		var m Message

		err = decoder.Decode(&m)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("can't read message from file '%s': %w", fName, err)
		}

		msgs = append(msgs, m)
	}

	return msgs, nil
}

func (p *P) connect(c *gnomock.Container) (*kafka.Conn, error) {
	return kafka.Dial("tcp", c.Address(BrokerPort))
}

// nolint: lll
func (p *P) sendMessagesIntoTopic(ctx context.Context, brokerAddr, topic string, messages []Message) (err error) {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{brokerAddr},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})

	defer func() {
		closeErr := w.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	kafkaMessages := make([]kafka.Message, len(messages))

	for i, m := range messages {
		kafkaMessages[i] = kafka.Message{
			Key:   []byte(m.Key),
			Value: []byte(m.Value),
			Time:  time.Unix(0, m.Time),
		}
	}

	if err := w.WriteMessages(ctx, kafkaMessages...); err != nil {
		return fmt.Errorf("write messages failed: %w", err)
	}

	return nil
}
