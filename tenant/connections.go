package tenant

import (
	"net/url"

	"github.com/pkg/errors"
)

const (
	ConnectionTypeNATS  = "nats"
	ConnectionTypeKafka = "kafka"
)

// ConnectionConfig is an interface that defines a connection configuration.
type ConnectionConfig interface {
	Validate() error
}

// NATSConfig describes a connection to a NATS server.
type NATSConfig struct {
	ServerAddress string `yaml:"serverAddress" json:"serverAddress"`
}

// NATSConfigFromMap returns a Kafka config from a map.
func NATSConfigFromMap(orig map[string]string) *NATSConfig {
	n := &NATSConfig{
		ServerAddress: orig["serverAddress"],
	}

	return n
}

func (n *NATSConfig) Validate() error {
	if n.ServerAddress == "" {
		return errors.New("serverAddress is empty")
	}

	if _, err := url.Parse(n.ServerAddress); err != nil {
		return errors.Wrap(err, "failed to parse serverAddress as URL")
	}

	return nil
}

// KafkaConfig describes a connection to a Kafka broker.
type KafkaConfig struct {
	BrokerAddress string `yaml:"brokerAddress" json:"brokerAddress"`
}

// KafkaConfigFromMap returns a NATS config from a map.
func KafkaConfigFromMap(orig map[string]string) *KafkaConfig {
	k := &KafkaConfig{
		BrokerAddress: orig["brokerAddress"],
	}

	return k
}

func (k *KafkaConfig) Validate() error {
	if k.BrokerAddress == "" {
		return errors.New("brokerAddress is empty")
	}

	if _, err := url.Parse(k.BrokerAddress); err != nil {
		return errors.Wrap(err, "failed to parse brokerAddress as URL")
	}

	return nil
}
