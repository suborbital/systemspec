package system

import (
	"errors"

	"github.com/suborbital/appspec/capabilities"
	"github.com/suborbital/appspec/tenant"
)

var (
	ErrModuleNotFound       = errors.New("failed to find requested module")
	ErrTenantNotFound       = errors.New("failed to find requested tenant")
	ErrNamespaceNotFound    = errors.New("failed to find requested namespace")
	ErrAuthenticationFailed = errors.New("failed to authenticate")
)

// AppSource  describes how an entire system relays its state to a client
type AppSource interface {
	// Start indicates to the Source that it should prepare for system startup.
	Start(opts Options) error

	// State returns the state of the entire system, used for cache invalidation and sync purposes
	State() (*State, error)

	// Overview returns a the system overview, used for incremental sync of the system's tenants
	Overview() (*Overview, error)

	// TenantOverview returns the overview for the requested tenant
	TenantOverview(ident string) (*TenantOverview, error)

	// GetModule attempts to find the given module by its fqmn, and returns ErrRunnableNotFound if it cannot.
	GetModule(FQMN string) (*tenant.Module, error)

	// Workflows returns the requested workflows for the system.
	Workflows(ident, namespace string, version int64) ([]tenant.Workflow, error)

	// Connections returns the connections needed for the system.
	Connections(ident, namespace string, version int64) ([]tenant.Connection, error)

	// Authentication provides any auth headers or metadata for the system.
	Authentication(ident, namespace string, version int64) (*tenant.Authentication, error)

	// Capabilities provides the tenant's configured capabilities.
	Capabilities(ident, namespace string, version int64) (*capabilities.CapabilityConfig, error)

	// StaticFile is a source of static files for the tenant
	StaticFile(identifier string, tenantVersion int64, path string) ([]byte, error)

	// Queries returns the database queries that should be made available.
	Queries(ident, namespace string, version int64) ([]tenant.DBQuery, error)
}
