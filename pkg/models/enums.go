package models

// Language represents a programming language
type Language string

const (
	LanguageC      Language = "c"
	LanguageCpp    Language = "cpp"
	LanguageCPP    Language = "c++" // Alias for cpp
	LanguageGo     Language = "go"
	LanguageRust   Language = "rust"
)

// Valid returns true if the language is valid
func (l Language) Valid() bool {
	switch l {
	case LanguageC, LanguageCpp, LanguageCPP, LanguageGo, LanguageRust:
		return true
	default:
		return false
	}
}

// Normalize converts language aliases to canonical form
func (l Language) Normalize() Language {
	if l == LanguageCPP {
		return LanguageCpp
	}
	return l
}

// Compiler represents a compiler
type Compiler string

const (
	CompilerGCC13   Compiler = "gcc-13"
	CompilerClang15 Compiler = "clang-15"
	CompilerGo      Compiler = "go"
	CompilerRustc   Compiler = "rustc"
)

// Valid returns true if the compiler is valid
func (c Compiler) Valid() bool {
	switch c {
	case CompilerGCC13, CompilerClang15, CompilerGo, CompilerRustc:
		return true
	default:
		return false
	}
}

// Standard represents a language standard (primarily for C++)
type Standard string

const (
	StandardCpp11 Standard = "c++11"
	StandardCpp14 Standard = "c++14"
	StandardCpp17 Standard = "c++17"
	StandardCpp20 Standard = "c++20"
	StandardCpp23 Standard = "c++23"
)

// Valid returns true if the standard is valid
func (s Standard) Valid() bool {
	switch s {
	case StandardCpp11, StandardCpp14, StandardCpp17, StandardCpp20, StandardCpp23:
		return true
	case "": // Empty is valid (will use default)
		return true
	default:
		return false
	}
}

// Architecture represents a CPU architecture
type Architecture string

const (
	ArchX86_64 Architecture = "x86_64"
	ArchARM64  Architecture = "arm64"
	ArchARM    Architecture = "arm"
)

// Valid returns true if the architecture is valid
func (a Architecture) Valid() bool {
	switch a {
	case ArchX86_64, ArchARM64, ArchARM:
		return true
	case "": // Empty is valid (will use default)
		return true
	default:
		return false
	}
}

// OS represents an operating system
type OS string

const (
	OSLinux   OS = "linux"
	OSWindows OS = "windows"
	OSMacOS   OS = "macos"
)

// Valid returns true if the OS is valid
func (o OS) Valid() bool {
	switch o {
	case OSLinux, OSWindows, OSMacOS:
		return true
	case "": // Empty is valid (will use default)
		return true
	default:
		return false
	}
}

// JobStatus represents the current status of a compilation job
type JobStatus string

const (
	StatusQueued     JobStatus = "queued"
	StatusProcessing JobStatus = "processing"
	StatusCompleted  JobStatus = "completed"
	StatusFailed     JobStatus = "failed"
	StatusTimeout    JobStatus = "timeout"
)
