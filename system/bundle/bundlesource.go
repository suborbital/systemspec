package bundle

import (
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/suborbital/systemspec/bundle"
	"github.com/suborbital/systemspec/capabilities"
	"github.com/suborbital/systemspec/system"
	"github.com/suborbital/systemspec/tenant"
)

// BundleSource is a Source backed by a bundle file.
type BundleSource struct {
	path   string
	bundle *bundle.Bundle

	lock sync.RWMutex
}

// NewBundleSource creates a new BundleSource that looks for a bundle at [path].
func NewBundleSource(path string) system.Source {
	b := &BundleSource{
		path: path,
		lock: sync.RWMutex{},
	}

	return b
}

// Start initializes the system source.
func (b *BundleSource) Start() error {
	if err := b.findBundle(); err != nil {
		return errors.Wrap(err, "failed to findBundle")
	}

	return nil
}

// State returns the state of the entire system.
func (b *BundleSource) State() (*system.State, error) {
	s := &system.State{
		SystemVersion: 1,
	}

	return s, nil
}

// Overview gets the overview for the entire system.
func (b *BundleSource) Overview() (*system.Overview, error) {
	ovv := &system.Overview{
		State: system.State{
			SystemVersion: 1,
		},
		TenantRefs: system.References{
			Identifiers: map[string]int64{
				b.bundle.TenantConfig.Identifier: b.bundle.TenantConfig.TenantVersion,
			},
		},
	}

	return ovv, nil
}

// Modules returns the Modules for the system.
func (b *BundleSource) TenantOverview(ident string) (*system.TenantOverview, error) {
	if !b.checkIdentifier(ident) {
		return nil, system.ErrTenantNotFound
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.bundle == nil {
		return nil, system.ErrTenantNotFound
	}

	ovv := &system.TenantOverview{
		Identifier: ident,
		Version:    b.bundle.TenantConfig.TenantVersion,
		Config:     b.bundle.TenantConfig,
	}

	return ovv, nil
}

// GetModule searches for and returns the requested module
// otherwise system.ErrModuleNotFound.
func (b *BundleSource) GetModule(FQMN string) (*tenant.Module, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.bundle == nil {
		return nil, system.ErrModuleNotFound
	}

	for i, r := range b.bundle.TenantConfig.Modules {
		if r.FQMN == FQMN {
			return &b.bundle.TenantConfig.Modules[i], nil
		}
	}

	return nil, system.ErrModuleNotFound
}

// Workflows returns the workflows for the system.
func (b *BundleSource) Workflows(ident, namespace string, _ int64) ([]tenant.Workflow, error) {
	if !b.checkIdentifier(ident) {
		return nil, system.ErrTenantNotFound
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.bundle == nil {
		return nil, system.ErrTenantNotFound
	}

	if namespace == "default" {
		return b.bundle.TenantConfig.DefaultNamespace.Workflows, nil
	}

	for _, n := range b.bundle.TenantConfig.Namespaces {
		if n.Name == namespace {
			return n.Workflows, nil
		}
	}

	return nil, system.ErrNamespaceNotFound
}

// Connections returns the Connections for the system.
func (b *BundleSource) Connections(ident, namespace string, _ int64) ([]tenant.Connection, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.bundle == nil {
		return nil, system.ErrTenantNotFound
	}

	if !b.checkIdentifier(ident) {
		return nil, system.ErrTenantNotFound
	}

	if namespace == "default" {
		return b.bundle.TenantConfig.DefaultNamespace.Connections, nil
	}

	for _, n := range b.bundle.TenantConfig.Namespaces {
		if n.Name == namespace {
			return n.Connections, nil
		}
	}

	return nil, system.ErrTenantNotFound
}

// Authentication returns the Authentication for the system.
func (b *BundleSource) Authentication(ident, namespace string, _ int64) (*tenant.Authentication, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.bundle == nil || b.bundle.TenantConfig.DefaultNamespace.Authentication == nil {
		return nil, system.ErrTenantNotFound
	}

	if !b.checkIdentifier(ident) {
		return nil, system.ErrTenantNotFound
	}

	if namespace == "default" {
		return b.bundle.TenantConfig.DefaultNamespace.Authentication, nil
	}

	for _, n := range b.bundle.TenantConfig.Namespaces {
		if n.Name == namespace {
			return n.Authentication, nil
		}
	}

	return nil, system.ErrTenantNotFound
}

// Capabilities returns the configuration for the system's capabilities.

func (b *BundleSource) Capabilities(ident, namespace string, _ int64) (*capabilities.CapabilityConfig, error) {
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

	return nil, system.ErrTenantNotFound
}

// findBundle loops forever until it finds a bundle at the configured path.
func (b *BundleSource) findBundle() error {
	for {
		bdl, err := bundle.Read(b.path)
		if err != nil {
			time.Sleep(time.Second)

			continue
		}

		b.lock.Lock()

		b.bundle = bdl

		if err := b.bundle.TenantConfig.Validate(); err != nil {
			return errors.Wrap(err, "failed to Validate tenant config")
		}

		b.lock.Unlock()

		break
	}

	return nil
}

// checkIdentifier checks whether the passed in identifier and version are for the current system running in the
// bundle or not. Returns true only if both match.
func (b *BundleSource) checkIdentifier(identifier string) bool {
	return b.bundle.TenantConfig.Identifier == identifier
}
