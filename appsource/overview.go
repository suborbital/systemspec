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
	Identifier string         `json:"identifier"`
	Version    int64          `json:"version"`
	Modules    []Module       `json:"modules"`
	Config     *tenant.Config `json:"config,omitempty"`
}

// Module is the metadata for a Module that belongs to a tenant
type Module struct {
	Name      string           `json:"name"`
	Namespace string           `json:"namespace"`
	Ref       string           `json:"ref"`
	FQFN      string           `json:"fqfn"`
	Revisions []ModuleRevision `json:"revisions"`
}

// ModuleRevision is a revision of a module
type ModuleRevision struct {
	Ref string `json:"ref"`
}
