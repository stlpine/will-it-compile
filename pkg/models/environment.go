package models

// EnvironmentSpec describes a compilation environment.
type EnvironmentSpec struct {
	Language     Language     `json:"language"`
	Compiler     Compiler     `json:"compiler"`
	Version      string       `json:"version"`
	Standard     Standard     `json:"standard,omitempty"`
	Architecture Architecture `json:"architecture"`
	OS           OS           `json:"os"`
	ImageTag     string       `json:"image_tag"` // Docker image tag
	Flags        []string     `json:"flags,omitempty"`
}

// Environment represents a supported compilation environment.
type Environment struct {
	Language  string   `json:"language"`
	Compilers []string `json:"compilers"`
	Standards []string `json:"standards,omitempty"`
	OSes      []string `json:"oses"`
	Arches    []string `json:"architectures"`
}
