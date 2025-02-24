package patch

import ()

type Patch struct {
	Selectors []string               `yaml:"selectors"`
	Values    map[string]interface{} `yaml:"values"`
}

type PatchFile struct {
	FormatVersion string  `yaml:"_format_version"`
	Patches       []Patch `yaml:"patches"`
}
