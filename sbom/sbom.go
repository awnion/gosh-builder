package sbom

import (
	"encoding/json"
)

type Component struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Purl    string `json:"purl"`
}

type SBOM struct {
	Components []Component `json:"components"`
}

func ParseSBOM(content []byte) (result *SBOM, err error) {
	err = json.Unmarshal(content, &result)
	return result, err
}
