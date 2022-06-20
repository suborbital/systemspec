package appsource

import "github.com/suborbital/vektor/vlog"

// Options describes the options for an appsource
type Options interface {
	Logger() *vlog.Logger
	Headless() bool
}
