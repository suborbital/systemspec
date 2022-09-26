package system

import "github.com/suborbital/vektor/vlog"

// Options describes the options for an system
type Options interface {
	Logger() *vlog.Logger
}
