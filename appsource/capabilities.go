package system

import (
	"github.com/pkg/errors"

	"github.com/suborbital/appspec/capabilities"
	"github.com/suborbital/appspec/tenant"
	"github.com/suborbital/vektor/vlog"
)

// ResolveCapabilitiesFromSource takes the ident, namespace, and version, and looks up the capabilities for that trio from the
// Source applying the user overrides over the default configurations.
func ResolveCapabilitiesFromSource(source AppSource, ident, namespace string, log *vlog.Logger) (*capabilities.CapabilityConfig, error) {
	defaultConfig := capabilities.DefaultCapabilityConfig()

	tenantOverview, err := source.TenantOverview(ident)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get TenantOverview for %s", ident)
	}

	userConfig, err := source.Capabilities(ident, namespace, tenantOverview.Config.TenantVersion)
	if err != nil || userConfig == nil {
		return &defaultConfig, nil
	}

	connections, err := source.Connections(ident, namespace, tenantOverview.Config.TenantVersion)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get Connections")
	} else if connections == nil {
		connections = []tenant.Connection{}
	}

	if userConfig.Logger != nil {
		defaultConfig.Logger = userConfig.Logger
	}

	if userConfig.HTTP != nil {
		defaultConfig.HTTP = userConfig.HTTP
	}

	if userConfig.GraphQL != nil {
		defaultConfig.GraphQL = userConfig.GraphQL
	}

	if userConfig.Auth != nil {
		defaultConfig.Auth = userConfig.Auth
	}

	// defaultConfig for the cache can come from either the capabilities
	// and/or connections sections of the tenant config.
	if userConfig.Cache != nil {
		defaultConfig.Cache = userConfig.Cache
	}

	for _, c := range connections {
		if c.Type == tenant.ConnectionTypeRedis {
			config := c.Config.(*tenant.RedisConnection)

			redisConfig := &capabilities.RedisConfig{
				ServerAddress: config.ServerAddress,
				Username:      config.Username,
				Password:      config.Password,
			}

			defaultConfig.Cache.RedisConfig = redisConfig
		}

		if c.Type == tenant.ConnectionTypeMySQL || c.Type == tenant.ConnectionTypePostgres {
			queries, err := source.Queries(ident, namespace, tenantOverview.Config.TenantVersion)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get Queries")
			}

			config := c.Config.(*tenant.DBConnection)

			dbConfig, err := config.ToRCAPConfig(queries)
			if err != nil {
				return nil, errors.Wrap(err, "failed to ToRCAPConfig")
			}

			defaultConfig.DB = dbConfig
		}
	}

	if userConfig.File != nil {
		defaultConfig.File = userConfig.File
	}

	// Override the connections.Database struct
	if userConfig.DB != nil && userConfig.DB.Enabled {
		defaultConfig.DB = userConfig.DB
	}

	if userConfig.Request != nil {
		defaultConfig.Request = userConfig.Request
	}

	f := func(pathName string) ([]byte, error) {
		return source.StaticFile(ident, tenantOverview.Config.TenantVersion, pathName)
	}

	defaultConfig.Logger.Logger = log
	defaultConfig.File.FileFunc = f

	return &defaultConfig, nil
}
