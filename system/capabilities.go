package system

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/suborbital/systemspec/capabilities"
)

// ResolveCapabilitiesFromSource takes the ident, namespace, and version, and looks up the capabilities for that trio from the
// Source applying the user overrides over the default configurations.
func ResolveCapabilitiesFromSource(source Source, ident, namespace string, log zerolog.Logger) (*capabilities.CapabilityConfig, error) {
	defaultConfig := capabilities.DefaultCapabilityConfig()

	tenantOverview, err := source.TenantOverview(ident)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get TenantOverview for %s", ident)
	}

	userConfig, err := source.Capabilities(ident, namespace, tenantOverview.Config.TenantVersion)
	if err != nil || userConfig == nil {
		return &defaultConfig, nil
	}

	if userConfig.Logger != nil {
		defaultConfig.Logger = userConfig.Logger
	}

	if userConfig.HTTP != nil {
		defaultConfig.HTTP = userConfig.HTTP
	}

	if userConfig.Auth != nil {
		defaultConfig.Auth = userConfig.Auth
	}

	if userConfig.Request != nil {
		defaultConfig.Request = userConfig.Request
	}

	defaultConfig.Logger.Logger = log

	return &defaultConfig, nil
}
