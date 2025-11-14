package models

import "fmt"

// CompilationRequest represents an incoming request to compile code
type CompilationRequest struct {
	Code         string       `json:"code"`                   // Base64 encoded source code
	Language     Language     `json:"language"`               // e.g., "cpp", "go", "rust"
	Standard     Standard     `json:"standard,omitempty"`     // e.g., "c++20", "c++17"
	Architecture Architecture `json:"architecture,omitempty"` // e.g., "x86_64", "arm64"
	OS           OS           `json:"os,omitempty"`           // e.g., "linux"
	Compiler     Compiler     `json:"compiler,omitempty"`     // e.g., "gcc-13", "clang-15"
}

// Validate validates the compilation request
func (r *CompilationRequest) Validate() error {
	if r.Code == "" {
		return fmt.Errorf("source code is required")
	}

	if !r.Language.Valid() {
		return fmt.Errorf("invalid language: %s", r.Language)
	}

	if !r.Standard.Valid() {
		return fmt.Errorf("invalid standard: %s", r.Standard)
	}

	if !r.Architecture.Valid() {
		return fmt.Errorf("invalid architecture: %s", r.Architecture)
	}

	if !r.OS.Valid() {
		return fmt.Errorf("invalid OS: %s", r.OS)
	}

	if r.Compiler != "" && !r.Compiler.Valid() {
		return fmt.Errorf("invalid compiler: %s", r.Compiler)
	}

	return nil
}
