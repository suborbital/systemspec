package tenant

import (
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	"github.com/suborbital/systemspec/capabilities"
)

const (
	ConnectionTypeNATS     = "nats"
	ConnectionTypeKafka    = "kafka"
	ConnectionTypeRedis    = "redis"
	ConnectionTypeMySQL    = "mysql"
	ConnectionTypePostgres = "postgres"
)

// ConnectionConfig is an interface that defines a connection configuration
type ConnectionConfig interface {
	Validate() error
}

// NATSConfig describes a connection to a NATS server.
type NATSConfig struct {
	ServerAddress string `yaml:"serverAddress" json:"serverAddress"`
}

// NATSConfigFromMap returns a Kafka config from a map
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

// KafkaConfigFromMap returns a NATS config from a map
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

type DBConfig struct {
	Type             string `yaml:"type" json:"type"`
	ConnectionString string `yaml:"connectionString" json:"connectionString"`
}

// DBConfigFromMap returns a DB config from a map
func DBConfigFromMap(orig map[string]string) *DBConfig {
	d := &DBConfig{
		Type:             orig["type"],
		ConnectionString: orig["connectionString"],
	}

	return d
}

func (d *DBConfig) ToRCAPConfig(queries []DBQuery) (*capabilities.DatabaseConfig, error) {
	if d == nil {
		return nil, nil
	}

	rcapType := capabilities.DBTypeMySQL
	if d.Type == "postgresql" {
		rcapType = capabilities.DBTypePostgres
	}

	rcapQueries := make([]capabilities.Query, len(queries))
	for i := range queries {
		q, err := queries[i].toRCAPQuery(rcapType)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to toRCAPQuery for %s", queries[i].Name)
		}

		rcapQueries[i] = *q
	}

	config := &capabilities.DatabaseConfig{
		Enabled:          d.ConnectionString != "",
		DBType:           rcapType,
		ConnectionString: d.ConnectionString,
		Queries:          rcapQueries,
	}

	return config, nil
}

func (d *DBConfig) Validate() error {
	if d.Type != ConnectionTypeMySQL && d.Type != ConnectionTypePostgres {
		return fmt.Errorf("database type %s is invalid, must be 'mysql' or 'postgres'", d.Type)
	}

	if d.ConnectionString == "" {
		return errors.New("database connectionString is empty")
	}

	return nil
}

// RedisConfig describes a connection to a Redis cache.
type RedisConfig struct {
	ServerAddress string `yaml:"serverAddress" json:"serverAddress"`
	Username      string `yaml:"username" json:"username"`
	Password      string `yaml:"password" json:"password"`
}

// RedisConfigFromMap returns a Redis config from a map
func RedisConfigFromMap(orig map[string]string) *RedisConfig {
	r := &RedisConfig{
		ServerAddress: orig["serverAddress"],
		Username:      orig["username"],
		Password:      orig["password"],
	}

	return r
}

func (r *RedisConfig) Validate() error {
	if r.ServerAddress == "" {
		return errors.New("serverAddress is empty")
	}

	if _, err := url.Parse(r.ServerAddress); err != nil {
		return errors.Wrap(err, "failed to parse serverAddress as URL")
	}

	return nil
}
