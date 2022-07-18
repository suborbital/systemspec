package directive

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v2"

	"github.com/suborbital/appspec/capabilities"
	"github.com/suborbital/appspec/directive/executable"
	fqmn "github.com/suborbital/appspec/fqfn"
)

// InputTypeRequest and others represent consts for Directives.
const (
	InputTypeRequest  = "request"
	InputTypeStream   = "stream"
	InputSourceServer = "server"
	InputSourceNATS   = "nats"
	InputSourceKafka  = "kafka"
)

// Directive describes a set of functions and a set of handlers
// that take an input, and compose a set of functions to handle it.
type Directive struct {
	Identifier     string                         `yaml:"identifier" json:"identifier"`
	AppVersion     string                         `yaml:"appVersion" json:"appVersion"`
	AtmoVersion    string                         `yaml:"atmoVersion" json:"atmoVersion"`
	Headless       bool                           `yaml:"headless,omitempty" json:"headless,omitempty"`
	Connections    *Connections                   `yaml:"connections,omitempty" json:"connections,omitempty"`
	Authentication *Authentication                `yaml:"authentication,omitempty" json:"authentication,omitempty"`
	Capabilities   *capabilities.CapabilityConfig `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
	Handlers       []Handler                      `yaml:"handlers,omitempty" json:"handlers,omitempty"`
	Schedules      []Schedule                     `yaml:"schedules,omitempty" json:"schedules,omitempty"`
	Queries        []DBQuery                      `yaml:"queries,omitempty" json:"queries,omitempty"`

	// Runnables is populated by subo, never by the user.
	Runnables []Runnable `yaml:"runnables" json:"runnables"`
}

// Handler represents the mapping between an input and a composition of functions.
type Handler struct {
	Input     Input                   `yaml:"input,inline" json:"input"`
	Steps     []executable.Executable `yaml:"steps" json:"steps"`
	Response  string                  `yaml:"response,omitempty" json:"response,omitempty"`
	RespondTo string                  `yaml:"respondTo,omitempty" json:"respondTo,omitempty"`
}

// Schedule represents the mapping between an input and a composition of functions.
type Schedule struct {
	Name  string                  `yaml:"name" json:"name"`
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

// Input represents an input source.
type Input struct {
	Type     string `yaml:"type" json:"type"`
	Source   string `yaml:"source,omitempty" json:"source,omitempty"`
	Method   string `yaml:"method" json:"method"`
	Resource string `yaml:"resource" json:"resource"`
}

// Connections describes connections.
type Connections struct {
	NATS     *NATSConnection  `yaml:"nats,omitempty" json:"nats,omitempty"`
	Kafka    *KafkaConnection `yaml:"kafka,omitempty" json:"kafka,omitempty"`
	Database *DBConnection    `yaml:"database,omitempty" json:"database,omitempty"`
	Redis    *RedisConnection `yaml:"redis,omitempty" json:"redis,omitempty"`
}

type Authentication struct {
	Domains map[string]capabilities.AuthHeader `yaml:"domains,omitempty" json:"domains,omitempty"`
}

func (d *Directive) FindRunnable(name string) *Runnable {
	// if this is an FQFN, parse the identifier and bail out
	// if it doesn't match this Directive.

	FQMN, err := fqmn.Parse(name)
	if err != nil {
		return nil
	}

	if FQMN.Tenant != "" && FQMN.Tenant != d.Identifier {
		return nil
	}

	if FQMN.Ref != "" && FQMN.Ref != d.AppVersion {
		return nil
	}

	for i, r := range d.Runnables {
		if r.Name == FQMN.Module && r.Namespace == FQMN.Namespace {
			return &d.Runnables[i]
		}
	}

	return nil
}

// Marshal outputs the YAML bytes of the Directive.
func (d *Directive) Marshal() ([]byte, error) {
	d.calculateFQMNs()

	return yaml.Marshal(d)
}

// Unmarshal unmarshals YAML bytes into a Directive struct
// it also calculates a map of FQMNs for later use.
func (d *Directive) Unmarshal(in []byte) error {
	if err := yaml.Unmarshal(in, d); err != nil {
		return err
	}

	d.calculateFQMNs()

	return nil
}

func (d *Directive) calculateFQMNs() {
	for i, fn := range d.Runnables {
		if fn.FQMN != "" {
			continue
		}

		if fn.Namespace == "" {
			fn.Namespace = fqmn.NamespaceDefault
		}

		if fn.Version == "" {
			fn.Version = d.AppVersion
		}

		d.Runnables[i].FQMN = d.FQMNForFunc(fn.Namespace, fn.Name)
	}
}

func (d *Directive) FQMNForFunc(namespace, fn string) string {
	return fqmn.FromParts(d.Identifier, namespace, fn, d.AppVersion)
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
