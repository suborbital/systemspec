package capabilities

import (
	"os"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var ErrCapabilityNotEnabled = errors.New("capability is not enabled")

// CapabilityConfig is configuration for a Module's capabilities
// NOTE: if any of the individual configs are nil, it will cause a crash,
// but we need to be able to determine if they're set or not, hence the pointers
// we are going to leave capabilities undocumented until we come up with a more elegant solution.
type CapabilityConfig struct {
	Logger  *LoggerConfig         `json:"logger,omitempty" yaml:"logger,omitempty"`
	HTTP    *HTTPConfig           `json:"http,omitempty" yaml:"http,omitempty"`
	Auth    *AuthConfig           `json:"auth,omitempty" yaml:"auth,omitempty"`
	Request *RequestHandlerConfig `json:"requestHandler,omitempty" yaml:"requestHandler,omitempty"`
}

// DefaultCapabilityConfig returns the default all-enabled config (with a default logger).
func DefaultCapabilityConfig() CapabilityConfig {
	return NewConfig(zerolog.New(os.Stderr))
}

func NewConfig(logger zerolog.Logger) CapabilityConfig {
	c := CapabilityConfig{
		Logger: &LoggerConfig{
			Enabled: true,
			Logger:  logger,
		},
		HTTP: &HTTPConfig{
			Enabled: true,
			Rules:   defaultHTTPRules(),
		},
		Auth: &AuthConfig{
			Enabled: true,
		},
		Request: &RequestHandlerConfig{
			Enabled:       true,
			AllowGetField: true,
			AllowSetField: true,
		},
	}

	return c
}
