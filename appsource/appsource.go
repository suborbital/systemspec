package appsource

import (
	"errors"

	"github.com/suborbital/appspec/capabilities"
	"github.com/suborbital/appspec/directive"
	"github.com/suborbital/appspec/fqfn"
)

var (
	ErrRunnableNotFound     = errors.New("failed to find requested Runnable")
	ErrAuthenticationFailed = errors.New("failed to authenticate")
)

// AuthFunc is a function that receives an application identifier
// and returns an authentication token, used for any authenticated
// AppSource calls in a multi-application system.
type AuthFunc func(appIdent string) string

// AppSource describes how an entire system relays its state to a client
type AppSource interface {
	// Start indicates to the AppSource that it should prepare for app startup.
	Start(opts Options) error

	// State returns the state of the entire system, used for cache invalidation and sync purposes
	State() (*State, error)

	// Overview returns a the system overview, used for incremental sync of the system's applications
	Overview() (*Overview, error)

	// ApplicationOverview returns the overview for the requested application
	ApplicationOverview(ident, appVersion string) (*ApplicationOverview, error)

	// GetFunction attempts to find the given Function by its fqfn, and returns ErrRunnableNotFound if it cannot.
	GetFunction(fqfn fqfn.FQFN) (*directive.Runnable, error)

	// Handlers returns the handlers for the app.
	Handlers(ident, appVersion string) ([]directive.Handler, error)

	// Schedules returns the requested schedules for the app.
	Schedules(ident, appVersion string) ([]directive.Schedule, error)

	// Connections returns the connections needed for the app.
	Connections(ident, appVersion string) (*directive.Connections, error)

	// Authentication provides any auth headers or metadata for the app.
	Authentication(ident, appVersion string) (*directive.Authentication, error)

	// Capabilities provides the application's configured capabilities.
	Capabilities(ident, namespace, appVersion string) (*capabilities.CapabilityConfig, error)

	// StaticFile is a source of static files for the application
	// TODO: refactor this into a set of capabilities / profiles.
	StaticFile(identifier, appVersion, path string) ([]byte, error)

	// Queries returns the database queries that should be made available.
	Queries(ident, appVersion string) []directive.DBQuery

	// UseAuthenticationFunc sets an auth function to be used when making authenticated calls as a client using transport-specific methods.
	// AppSource servers should not use this, and instead validate provided auth info within their implementaitons.
	UseAuthenticationFunc(AuthFunc)
}
