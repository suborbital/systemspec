package tenant

import (
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	"github.com/suborbital/appspec/capabilities"
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

// NATSConnection describes a connection to a NATS server.
type NATSConnection struct {
	ServerAddress string `yaml:"serverAddress" json:"serverAddress"`
}

func (n *NATSConnection) Validate() error {
	if n.ServerAddress == "" {
		return errors.New("serverAddress is empty")
	}

	if _, err := url.Parse(n.ServerAddress); err != nil {
		return errors.Wrap(err, "failed to parse serverAddress as URL")
	}

	return nil
}

// KafkaConnection describes a connection to a Kafka broker.
type KafkaConnection struct {
	BrokerAddress string `yaml:"brokerAddress" json:"brokerAddress"`
}

func (k *KafkaConnection) Validate() error {
	if k.BrokerAddress == "" {
		return errors.New("brokerAddress is empty")
	}

	if _, err := url.Parse(k.BrokerAddress); err != nil {
		return errors.Wrap(err, "failed to parse brokerAddress as URL")
	}

	return nil
}

type DBConnection struct {
	Type             string `yaml:"type" json:"type"`
	ConnectionString string `yaml:"connectionString" json:"connectionString"`
}

func (d *DBConnection) ToRCAPConfig(queries []DBQuery) (*capabilities.DatabaseConfig, error) {
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

func (d *DBConnection) Validate() error {
	if d.Type != ConnectionTypeMySQL && d.Type != ConnectionTypePostgres {
		return fmt.Errorf("database type %s is invalid, must be 'mysql' or 'postgres'", d.Type)
	}

	if d.ConnectionString == "" {
		return errors.New("database connectionString is empty")
	}

	return nil
}

// RedisConnection describes a connection to a Redis cache.
type RedisConnection struct {
	ServerAddress string `yaml:"serverAddress" json:"serverAddress"`
	Username      string `yaml:"username" json:"username"`
	Password      string `yaml:"password" json:"password"`
}

func (r *RedisConnection) Validate() error {
	if r.ServerAddress == "" {
		return errors.New("serverAddress is empty")
	}

	if _, err := url.Parse(r.ServerAddress); err != nil {
		return errors.Wrap(err, "failed to parse serverAddress as URL")
	}

	return nil
}
