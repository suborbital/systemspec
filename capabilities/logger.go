package capabilities

import (
	"github.com/rs/zerolog"
)

// LoggerConfig is configuration for the logger capability
type LoggerConfig struct {
	Enabled bool           `json:"enabled" yaml:"enabled"`
	Logger  zerolog.Logger `json:"-" yaml:"-"`
}

// LoggerCapability provides a logger to Modules
type LoggerCapability interface {
	Log(level int32, msg string, scope interface{})
}

type loggerSource struct {
	config LoggerConfig
	log    zerolog.Logger
}

// DefaultLoggerSource returns a LoggerSource that provides a zerolog.Logger that's in the passed in
// config struct.
func DefaultLoggerSource(config LoggerConfig) LoggerCapability {
	l := &loggerSource{
		config: config,
		log:    config.Logger,
	}

	return l
}

// Log writes a log line to the underlying logger using the data it got:
// level int32, msg string, and scope interface
func (l *loggerSource) Log(level int32, msg string, scope interface{}) {
	if !l.config.Enabled {
		return
	}

	scoped := l.log.With().Interface("scope", scope).Logger()

	switch level {
	case 1:
		scoped.Error().Msg(msg)
	case 2:
		scoped.Warn().Msg(msg)
	case 4:
		scoped.Debug().Msg(msg)
	default:
		scoped.Info().Msg(msg)
	}
}
