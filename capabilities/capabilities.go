package capabilities

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var ErrCapabilityNotAvailable = errors.New("capability not available")

// Capabilities define the capabilities available to a Module
type Capabilities struct {
	config CapabilityConfig

	Auth         AuthCapability
	LoggerSource LoggerCapability
	HTTPClient   HTTPCapability

	// RequestHandler and doFunc are special because they are more
	// sensitive; they could cause memory leaks or expose internal state,
	// so they cannot be swapped out for a different implementation.
	RequestConfig *RequestHandlerConfig
}

// New returns the default capabilities with the provided Logger
func New(logger zerolog.Logger) *Capabilities {
	// this will never error with the default config, as the db capability is disabled
	caps, _ := NewWithConfig(NewConfig(logger))

	return caps
}

func NewWithConfig(config CapabilityConfig) (*Capabilities, error) {
	caps := &Capabilities{
		config:        config,
		Auth:          DefaultAuthProvider(*config.Auth),
		LoggerSource:  DefaultLoggerSource(*config.Logger),
		HTTPClient:    DefaultHTTPClient(*config.HTTP),
		RequestConfig: config.Request,
	}

	return caps, nil
}

// Config returns the configuration that was used to create the Capabilities
// the config cannot be changed, but it can be used to determine what was
// previously set so that the orginal config (like enabled settings) can be respected
func (c Capabilities) Config() CapabilityConfig {
	return c.config
}
