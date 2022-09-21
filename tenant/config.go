package tenant

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/suborbital/appspec/capabilities"
	fqmn "github.com/suborbital/appspec/fqmn"
	"github.com/suborbital/appspec/tenant/executable"
)

// InputTypeRequest and others represent consts for the tenantConfig.
const (
	InputTypeRequest  = "request"
	InputTypeStream   = "stream"
	InputSourceServer = "server"
	InputSourceNATS   = "nats"
	InputSourceKafka  = "kafka"
)

// Config describes a tenant and its related config
type Config struct {
	Identifier       string            `yaml:"identifier" json:"identifier"`
	SpecVersion      int               `yaml:"specVersion" json:"specVersion"`
	TenantVersion    int64             `yaml:"tenantVersion" json:"tenantVersion"`
	DefaultNamespace NamespaceConfig   `yaml:"defaultNamespace" json:"defaultNamespace"`
	Namespaces       []NamespaceConfig `yaml:"namespaces" json:"namespaces"`

	// Modules is a union of all consituent namespace modules
	Modules []Module `yaml:"modules" json:"modules"`
}

// NamespaceConfig is the configuration for a namespace
type NamespaceConfig struct {
	Name           string                         `yaml:"name" json:"name"`
	Workflows      []Workflow                     `yaml:"workflows,omitempty" json:"workflows,omitempty"`
	Queries        []DBQuery                      `yaml:"queries,omitempty" json:"queries,omitempty"`
	Capabilities   *capabilities.CapabilityConfig `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
	Connections    []Connection                   `yaml:"connections,omitempty" json:"connections,omitempty"`
	Authentication *Authentication                `yaml:"authentication,omitempty" json:"authentication,omitempty"`

	// Modules is populated by subo, never by the user.
	Modules []Module `yaml:"modules" json:"modules"`
}

// Workflow represents the mapping between an input and a composition of functions.
type Workflow struct {
	Name     string                  `yaml:"name" json:"name"`
	Steps    []executable.Executable `yaml:"steps" json:"steps"`
	Response string                  `yaml:"response,omitempty" json:"response,omitempty"`
	Schedule *Schedule               `yaml:"schedule,omitempty" json:"schedule,omitempty"`
	Triggers []Trigger               `yaml:"triggers" json:"triggers"`
}

// Schedule represents the schedule settings for a workflow
type Schedule struct {
	Every ScheduleEvery           `yaml:"every" json:"every"`
	State map[string]string       `yaml:"state,omitempty" json:"state,omitempty"`
	Steps []executable.Executable `yaml:"steps" json:"steps"`
}

// ScheduleEvery represents the 'every' value for a schedule.
type ScheduleEvery struct {
	Seconds int `yaml:"seconds,omitempty" json:"seconds,omitempty"`
	Minutes int `yaml:"minutes,omitempty" json:"minutes,omitempty"`
	Hours   int `yaml:"hours,omitempty" json:"hours,omitempty"`
	Days    int `yaml:"days,omitempty" json:"days,omitempty"`
}

// Trigger represents a trigger for a workflow.
type Trigger struct {
	Source    string `yaml:"source,omitempty" json:"source,omitempty"`
	Topic     string `yaml:"topic" json:"topic"`
	Sink      string `yaml:"sink" json:"sink"`
	SinkTopic string `yaml:"sinkTopic" json:"sinkTopic"`
}

// Connection describes a connection to an external resource
type Connection struct {
	Type   string           `yaml:"type" json:"type"`
	Name   string           `yaml:"name" json:"name"`
	Config ConnectionConfig `yaml:",inline" json:",inline"`
}

type Authentication struct {
	Domains map[string]capabilities.AuthHeader `yaml:"domains,omitempty" json:"domains,omitempty"`
}

func (c *Config) FindModule(name string) (*Module, error) {
	// if this is an FQMN, parse the identifier and bail out
	// if it doesn't match this tenant.

	FQMN, err := fqmn.Parse(name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fqmn.Parse")
	}

	if FQMN.Tenant != "" && FQMN.Tenant != c.Identifier {
		return nil, nil
	}

	for i, r := range c.Modules {
		if r.Name == FQMN.Name && r.Namespace == FQMN.Namespace {
			return &c.Modules[i], nil
		}
	}

	return nil, nil
}

// Marshal outputs the JSON bytes of the config.
func (c *Config) Marshal() ([]byte, error) {
	c.calculateFQMNs()

	return json.Marshal(c)
}

// Unmarshal unmarshals JSON bytes into a TenantConfig struct
// it also calculates a map of FQMNs for later use.
func (c *Config) Unmarshal(in []byte) error {
	if err := json.Unmarshal(in, c); err != nil {
		return err
	}

	c.calculateFQMNs()

	return nil
}

func (c *Config) calculateFQMNs() {
	for i, mod := range c.Modules {
		if mod.FQMN != "" {
			continue
		}

		if mod.Namespace == "" {
			mod.Namespace = fqmn.NamespaceDefault
		}

		// We deliberately ignore returned errors.
		// The module will not be module, but it's not a problem for the
		// system as a whole.
		c.Modules[i].FQMN, _ = c.FQMNForFunc(mod.Namespace, mod.Name, mod.Ref)
	}
}

func (c *Config) FQMNForFunc(namespace, fn, ref string) (string, error) {
	return fqmn.FromParts(c.Identifier, namespace, fn, ref)
}

// NumberOfSeconds calculates the total time in seconds for the schedule's 'every' value.
func (s *Schedule) NumberOfSeconds() int {
	seconds := s.Every.Seconds
	minutes := 60 * s.Every.Minutes
	hours := 60 * 60 * s.Every.Hours
	days := 60 * 60 * 24 * s.Every.Days

	return seconds + minutes + hours + days
}

type problems []error

func (p *problems) add(err error) {
	*p = append(*p, err)
}

func (p *problems) render() error {
	if len(*p) == 0 {
		return nil
	}

	text := fmt.Sprintf("found %d problems:", len(*p))

	for _, err := range *p {
		text += fmt.Sprintf("\n\t%s", err.Error())
	}

	return errors.New(text)
}
