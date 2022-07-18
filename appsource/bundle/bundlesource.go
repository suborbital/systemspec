package appsource

import (
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/suborbital/appspec/appsource"
	"github.com/suborbital/appspec/bundle"
	"github.com/suborbital/appspec/capabilities"
	"github.com/suborbital/appspec/directive"
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
				b.bundle.Directive.Identifier: 1,
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

	modules := make([]appsource.Module, len(b.bundle.Directive.Runnables))

	for i, r := range b.bundle.Directive.Runnables {
		m := appsource.Module{
			Name:      r.Name,
			Namespace: r.Namespace,
			Ref:       "",
			FQFN:      r.FQMN,
			Revisions: []appsource.ModuleRevision{},
		}

		modules[i] = m
	}

	ovv := &appsource.TenantOverview{
		Identifier: ident,
		Version:    1,
		Modules:    modules,
	}

	return ovv, nil
}

// FindRunnable searches for and returns the requested runnable
// otherwise appsource.ErrFunctionNotFound.
func (b *BundleSource) GetModule(FQFN string) (*appsource.Module, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.bundle == nil {
		return nil, appsource.ErrModuleNotFound
	}

	for _, r := range b.bundle.Directive.Runnables {
		if r.FQMN == FQFN {
			m := &appsource.Module{
				Name:      r.Name,
				Namespace: r.Namespace,
				Ref:       "",
				FQFN:      r.FQMN,
				Revisions: []appsource.ModuleRevision{},
			}

			return m, nil
		}
	}

	return nil, appsource.ErrModuleNotFound
}

// Schedules returns the schedules for the app.
func (b *BundleSource) Workflows(ident, namespace string, version int64) ([]directive.Schedule, error) {
	if !b.checkIdentifier(ident) {
		return nil, appsource.ErrTenantNotFound
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.bundle == nil {
		return nil, appsource.ErrTenantNotFound
	}

	return b.bundle.Directive.Schedules, nil
}

// Connections returns the Connections for the app.
func (b *BundleSource) Connections(ident, namespace string, version int64) (*directive.Connections, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.bundle == nil || b.bundle.Directive.Connections == nil {
		return nil, appsource.ErrTenantNotFound
	}

	if !b.checkIdentifier(ident) {
		return nil, appsource.ErrTenantNotFound
	}

	return b.bundle.Directive.Connections, nil
}

// Authentication returns the Authentication for the app.
func (b *BundleSource) Authentication(ident, namespace string, version int64) (*directive.Authentication, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.bundle == nil || b.bundle.Directive.Authentication == nil {
		return nil, appsource.ErrTenantNotFound
	}

	if !b.checkIdentifier(ident) {
		return nil, appsource.ErrTenantNotFound
	}

	return b.bundle.Directive.Authentication, nil
}

// Capabilities returns the configuration for the app's capabilities.

func (b *BundleSource) Capabilities(ident, namespace string, version int64) (*capabilities.CapabilityConfig, error) {
	defaultConfig := capabilities.DefaultCapabilityConfig()

	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.bundle == nil || b.bundle.Directive.Capabilities == nil {
		return &defaultConfig, nil
	}

	if !b.checkIdentifier(ident) {
		return &defaultConfig, nil
	}

	return b.bundle.Directive.Capabilities, nil
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
func (b *BundleSource) Queries(ident, namespace string, version int64) ([]directive.DBQuery, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.bundle == nil || b.bundle.Directive.Queries == nil {
		return nil, appsource.ErrTenantNotFound
	}

	if !b.checkIdentifier(ident) {
		return nil, appsource.ErrTenantNotFound
	}

	return b.bundle.Directive.Queries, nil
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

		if err := b.bundle.Directive.Validate(); err != nil {
			return errors.Wrap(err, "failed to Validate Directive")
		}

		break
	}

	return nil
}

// checkIdentifier checks whether the passed in identifier and version are for the current app running in the
// bundle or not. Returns true only if both match.
func (b *BundleSource) checkIdentifier(identifier string) bool {
	return b.bundle.Directive.Identifier == identifier
}
