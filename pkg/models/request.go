package models

import (
	"errors"
	"fmt"
)

// Sentinel errors for request validation.
var (
	ErrSourceCodeRequired  = errors.New("source code is required")
	ErrInvalidLanguage     = errors.New("invalid language")
	ErrInvalidStandard     = errors.New("invalid standard")
	ErrInvalidArchitecture = errors.New("invalid architecture")
	ErrInvalidOS           = errors.New("invalid OS")
	ErrInvalidCompiler     = errors.New("invalid compiler")
)

// CompilationRequest represents an incoming request to compile code.
type CompilationRequest struct {
	Code         string       `json:"code"`                   // Base64 encoded source code
	Language     Language     `json:"language"`               // e.g., "cpp", "go", "rust"
	Standard     Standard     `json:"standard,omitempty"`     // e.g., "c++20", "c++17"
	Architecture Architecture `json:"architecture,omitempty"` // e.g., "x86_64", "arm64"
	OS           OS           `json:"os,omitempty"`           // e.g., "linux"
	Compiler     Compiler     `json:"compiler,omitempty"`     // e.g., "gcc-13", "clang-15"
}

// Validate validates the compilation request.
func (r *CompilationRequest) Validate() error {
	if r.Code == "" {
		return ErrSourceCodeRequired
	}

	if !r.Language.Valid() {
		return fmt.Errorf("%w: %s", ErrInvalidLanguage, r.Language)
	}

	if !r.Standard.Valid() {
		return fmt.Errorf("%w: %s", ErrInvalidStandard, r.Standard)
	}

	if !r.Architecture.Valid() {
		return fmt.Errorf("%w: %s", ErrInvalidArchitecture, r.Architecture)
	}

	if !r.OS.Valid() {
		return fmt.Errorf("%w: %s", ErrInvalidOS, r.OS)
	}

	if r.Compiler != "" && !r.Compiler.Valid() {
		return fmt.Errorf("%w: %s", ErrInvalidCompiler, r.Compiler)
	}

	return nil
}
