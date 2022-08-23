package bundle

import (
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/suborbital/appspec/appsource"
	"github.com/suborbital/appspec/bundle"
	"github.com/suborbital/appspec/capabilities"
	"github.com/suborbital/appspec/tenant"
)

// BundleSource is an AppSource backed by a bundle file.
type BundleSource struct {
	path   string
	opts   appsource.Options
	bundle *bundle.Bundle

	lock sync.RWMutex
}

// NewBundleSource creates a new BundleSource that looks for a bundle at [path].
func NewBundleSource(path string) appsource.AppSource {
	b := &BundleSource{
		path: path,
		lock: sync.RWMutex{},
	}

	return b
}

// Start initializes the app source.
func (b *BundleSource) Start(opts appsource.Options) error {
	b.opts = opts

	if err := b.findBundle(); err != nil {
		return errors.Wrap(err, "failed to findBundle")
	}

	return nil
}

// State returns the state of the entire system
func (b *BundleSource) State() (*appsource.State, error) {
	s := &appsource.State{
		SystemVersion: 1,
	}

	return s, nil
}

// Overview gets the overview for the entire system.
func (b *BundleSource) Overview() (*appsource.Overview, error) {
	ovv := &appsource.Overview{
		State: appsource.State{
			SystemVersion: 1,
		},
		TenantRefs: appsource.References{
			Identifiers: map[string]int64{
				b.bundle.TenantConfig.Identifier: b.bundle.TenantConfig.TenantVersion,
			},
		},
	}

	return ovv, nil
}

// Runnables returns the Runnables for the app.
func (b *BundleSource) TenantOverview(ident string) (*appsource.TenantOverview, error) {
	if !b.checkIdentifier(ident) {
		return nil, appsource.ErrTenantNotFound
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.bundle == nil {
		return nil, appsource.ErrTenantNotFound
	}

	ovv := &appsource.TenantOverview{
		Identifier: ident,
		Version:    b.bundle.TenantConfig.TenantVersion,
		Config:     b.bundle.TenantConfig,
	}

	return ovv, nil
}

// FindRunnable searches for and returns the requested runnable
// otherwise appsource.ErrFunctionNotFound.
func (b *BundleSource) GetModule(FQMN string) (*tenant.Module, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.bundle == nil {
		return nil, appsource.ErrModuleNotFound
	}

	for i, r := range b.bundle.TenantConfig.Modules {
		if r.FQMN == FQMN {
			return &b.bundle.TenantConfig.Modules[i], nil
		}
	}

	return nil, appsource.ErrModuleNotFound
}

// Schedules returns the schedules for the app.
func (b *BundleSource) Workflows(ident, namespace string, version int64) ([]tenant.Workflow, error) {
	if !b.checkIdentifier(ident) {
		return nil, appsource.ErrTenantNotFound
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.bundle == nil {
		return nil, appsource.ErrTenantNotFound
	}

	if namespace == "default" {
		return b.bundle.TenantConfig.DefaultNamespace.Workflows, nil
	}

	for _, n := range b.bundle.TenantConfig.Namespaces {
		if n.Name == namespace {
			return n.Workflows, nil
		}
	}

	return nil, appsource.ErrNamespaceNotFound
}

// Connections returns the Connections for the app.
func (b *BundleSource) Connections(ident, namespace string, version int64) ([]tenant.Connection, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.bundle == nil {
		return nil, appsource.ErrTenantNotFound
	}

	if !b.checkIdentifier(ident) {
		return nil, appsource.ErrTenantNotFound
	}

	if namespace == "default" {
		return b.bundle.TenantConfig.DefaultNamespace.Connections, nil
	}

	for _, n := range b.bundle.TenantConfig.Namespaces {
		if n.Name == namespace {
			return n.Connections, nil
		}
	}

	return nil, appsource.ErrTenantNotFound
}

// Authentication returns the Authentication for the app.
func (b *BundleSource) Authentication(ident, namespace string, version int64) (*tenant.Authentication, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.bundle == nil || b.bundle.TenantConfig.DefaultNamespace.Authentication == nil {
		return nil, appsource.ErrTenantNotFound
	}

	if !b.checkIdentifier(ident) {
		return nil, appsource.ErrTenantNotFound
	}

	if namespace == "default" {
		return b.bundle.TenantConfig.DefaultNamespace.Authentication, nil
	}

	for _, n := range b.bundle.TenantConfig.Namespaces {
		if n.Name == namespace {
			return n.Authentication, nil
		}
	}

	return nil, appsource.ErrTenantNotFound
}

// Capabilities returns the configuration for the app's capabilities.

func (b *BundleSource) Capabilities(ident, namespace string, version int64) (*capabilities.CapabilityConfig, error) {
	defaultConfig := capabilities.DefaultCapabilityConfig()

	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.bundle == nil || b.bundle.TenantConfig.DefaultNamespace.Capabilities == nil {
		return &defaultConfig, nil
	}

	if !b.checkIdentifier(ident) {
		return &defaultConfig, nil
	}

	if namespace == "default" {
		return b.bundle.TenantConfig.DefaultNamespace.Capabilities, nil
	}

	for _, n := range b.bundle.TenantConfig.Namespaces {
		if n.Name == namespace {
			return n.Capabilities, nil
		}
	}

	return nil, appsource.ErrTenantNotFound
}

// File returns a requested file.
func (b *BundleSource) StaticFile(ident, namespace, filename string, version int64) ([]byte, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.bundle == nil {
		return nil, os.ErrNotExist
	}

	if !b.checkIdentifier(ident) {
		return nil, os.ErrNotExist
	}

	return b.bundle.StaticFile(filename)
}

// Queries returns the Queries available to the app.
func (b *BundleSource) Queries(ident, namespace string, version int64) ([]tenant.DBQuery, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.bundle == nil || b.bundle.TenantConfig.DefaultNamespace.Queries == nil {
		return nil, appsource.ErrTenantNotFound
	}

	if !b.checkIdentifier(ident) {
		return nil, appsource.ErrTenantNotFound
	}

	if namespace == "default" {
		return b.bundle.TenantConfig.DefaultNamespace.Queries, nil
	}

	for _, n := range b.bundle.TenantConfig.Namespaces {
		if n.Name == namespace {
			return n.Queries, nil
		}
	}

	return nil, appsource.ErrTenantNotFound
}

// findBundle loops forever until it finds a bundle at the configured path.
func (b *BundleSource) findBundle() error {
	for {
		bdl, err := bundle.Read(b.path)
		if err != nil {

			b.opts.Logger().Warn("failed to Read bundle, will try again:", err.Error())
			time.Sleep(time.Second)

			continue
		}

		b.opts.Logger().Debug("loaded bundle from", b.path)

		b.lock.Lock()
		defer b.lock.Unlock()

		b.bundle = bdl

		if err := b.bundle.TenantConfig.Validate(); err != nil {
			return errors.Wrap(err, "failed to Validate tenant config")
		}

		break
	}

	return nil
}

// checkIdentifier checks whether the passed in identifier and version are for the current app running in the
// bundle or not. Returns true only if both match.
func (b *BundleSource) checkIdentifier(identifier string) bool {
	return b.bundle.TenantConfig.Identifier == identifier
}
