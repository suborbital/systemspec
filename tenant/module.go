package tenant

// Module is the structure of a .Module.yaml file.
type Module struct {
	Name       string           `yaml:"name" json:"name"`
	Namespace  string           `yaml:"namespace" json:"namespace"`
	Lang       string           `yaml:"lang" json:"lang"`
	Ref        string           `yaml:"ref" json:"ref"`
	DraftRef   string           `yaml:"draftRef,omitempty" json:"draftRef,omitempty"`
	APIVersion string           `yaml:"apiVersion,omitempty" json:"apiVersion,omitempty"` // the version of the API / SDK that this module was built with
	FQMN       string           `yaml:"fqmn,omitempty" json:"fqmn,omitempty"`
	Revisions  []ModuleRevision `yaml:"revisions" json:"revisions"`
	WasmRef    *WasmModuleRef   `yaml:"-" json:"wasmRef,omitempty"`
}

// WasmModuleRef is a reference to a Wasm module
// This is a duplicate of sat/engine/moduleref/WasmModuleRef (for JSON serialization purposes).
type WasmModuleRef struct {
	Name string `json:"name"`
	FQMN string `json:"fqmn"`
	Data []byte `json:"data"`
}

// ModuleRevision is a revision of a module.
type ModuleRevision struct {
	Ref string `json:"ref"`
}

func NewWasmModuleRef(name, fqmn string, data []byte) *WasmModuleRef {
	w := &WasmModuleRef{
		Name: name,
		FQMN: fqmn,
		Data: data,
	}

	return w
}
