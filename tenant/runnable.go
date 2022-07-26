package tenant

// Module is the structure of a .Module.yaml file.
type Module struct {
	Name       string         `yaml:"name" json:"name"`
	Namespace  string         `yaml:"namespace" json:"namespace"`
	Lang       string         `yaml:"lang" json:"lang"`
	Ref        string         `yaml:"version" json:"version"`
	DraftRef   string         `yaml:"draftVersion,omitempty" json:"draftVersion,omitempty"`
	APIVersion string         `yaml:"apiVersion,omitempty" json:"apiVersion,omitempty"`
	FQMN       string         `yaml:"fqfn,omitempty" json:"fqfn,omitempty"`
	FQMNURI    string         `yaml:"fqfnUri" json:"fqfnURI,omitempty"`
	WasmRef    *WasmModuleRef `yaml:"-" json:"moduleRef,omitempty"`
	TokenHash  []byte         `yaml:"-" json:"-"`
}

// WasmModuleRef is a reference to a Wasm module
// This is a duplicate of sat/engine/moduleref/WasmModuleRef (for JSON serialization purposes)
type WasmModuleRef struct {
	Name string `json:"name"`
	FQFN string `json:"fqfn"`
	Data []byte `json:"data"`
}

func NewWasmModuleRef(name, fqfn string, data []byte) *WasmModuleRef {
	w := &WasmModuleRef{
		Name: name,
		FQFN: fqfn,
		Data: data,
	}

	return w
}
