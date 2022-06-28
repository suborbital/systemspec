package appsource

import "github.com/suborbital/appspec/directive"

// State describes the state of the entire system
type State struct {
	SystemVersion string `json:"systemVersion"`
}

// Overview is an overview of all the applications within the system
type Overview struct {
	SystemVersion   string     `json:"systemVersion"`
	ApplicationRefs References `json:"appReferences"`
}

// References are maps of all the available applications in the system
type References struct {
	// map of all app idents to their HEAD AppVersion
	Identifiers map[string]string `json:"identifiers"`
}

// ApplicationOverview describes the metadata for an application
type ApplicationOverview struct {
	Identifier  string                `json:"identifier"`
	AppVersion  string                `json:"appVersion"`
	Domain      string                `json:"domain"`
	Functions   map[string]Function   `json:"functions"`
	StaticFiles map[string]StaticFile `json:"staticFiles"`
	Directive   *directive.Directive  `json:"directive,omitempty"`
}

// Function is the metadata for a function that belongs to an application
type Function struct {
	Name       string `json:"name"`
	AppVersion string `json:"appVersion"`
	ModuleHash string `json:"moduleHash"`
	FQFN       string `json:"fqfn"`
}

// StaticFile represents a static file belonging to an application
type StaticFile struct {
	Name       string `json:"name"`
	AppVersion string `json:"appVersion"`
	FileHash   string `json:"fileHash"`
}
