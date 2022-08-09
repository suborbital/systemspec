package appsource

import "github.com/suborbital/appspec/tenant"

// State describes the state of the entire system
type State struct {
	SystemVersion int64 `json:"systemVersion"`
}

// Overview is an overview of all the applications within the system
type Overview struct {
	State
	TenantRefs References `json:"tenantReferences"`
}

// References are maps of all the available applications in the system
type References struct {
	// map of all tenant idents to their latest tenant version
	Identifiers map[string]int64 `json:"identifiers"`
}

// TenantOverview describes the metadata for a tenant
type TenantOverview struct {
	Identifier string          `json:"identifier"`
	Version    int64           `json:"version"`
	Modules    []tenant.Module `json:"modules"`
	Config     *tenant.Config  `json:"config,omitempty"`
}
