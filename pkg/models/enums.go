package models

// Language represents a programming language.
type Language string

const (
	LanguageC    Language = "c"
	LanguageCpp  Language = "cpp"
	LanguageCPP  Language = "c++" // Alias for cpp
	LanguageGo   Language = "go"
	LanguageRust Language = "rust"
)

// Valid returns true if the language is valid.
func (l Language) Valid() bool {
	switch l {
	case LanguageC, LanguageCpp, LanguageCPP, LanguageGo, LanguageRust:
		return true
	default:
		return false
	}
}

// Normalize converts language aliases to canonical form.
func (l Language) Normalize() Language {
	if l == LanguageCPP {
		return LanguageCpp
	}
	return l
}

// Compiler represents a compiler.
type Compiler string

const (
	// GCC versions (C/C++)
	CompilerGCC9  Compiler = "gcc-9"
	CompilerGCC10 Compiler = "gcc-10"
	CompilerGCC11 Compiler = "gcc-11"
	CompilerGCC12 Compiler = "gcc-12"
	CompilerGCC13 Compiler = "gcc-13"

	// Go versions
	CompilerGo120 Compiler = "go-1.20"
	CompilerGo121 Compiler = "go-1.21"
	CompilerGo122 Compiler = "go-1.22"
	CompilerGo123 Compiler = "go-1.23"

	// Rust versions
	CompilerRustc170 Compiler = "rustc-1.70"
	CompilerRustc175 Compiler = "rustc-1.75"
	CompilerRustc180 Compiler = "rustc-1.80"
)

// Valid returns true if the compiler is valid.
func (c Compiler) Valid() bool {
	switch c {
	// GCC versions
	case CompilerGCC9, CompilerGCC10, CompilerGCC11, CompilerGCC12, CompilerGCC13:
		return true
	// Go versions
	case CompilerGo120, CompilerGo121, CompilerGo122, CompilerGo123:
		return true
	// Rust versions
	case CompilerRustc170, CompilerRustc175, CompilerRustc180:
		return true
	default:
		return false
	}
}

// Standard represents a language standard (for C and C++).
type Standard string

const (
	// C++ standards
	StandardCpp11 Standard = "c++11"
	StandardCpp14 Standard = "c++14"
	StandardCpp17 Standard = "c++17"
	StandardCpp20 Standard = "c++20"
	StandardCpp23 Standard = "c++23"

	// C standards
	StandardC89 Standard = "c89"
	StandardC99 Standard = "c99"
	StandardC11 Standard = "c11"
	StandardC17 Standard = "c17"
	StandardC23 Standard = "c23"
)

// Valid returns true if the standard is valid.
func (s Standard) Valid() bool {
	switch s {
	case StandardCpp11, StandardCpp14, StandardCpp17, StandardCpp20, StandardCpp23:
		return true
	case StandardC89, StandardC99, StandardC11, StandardC17, StandardC23:
		return true
	case "": // Empty is valid (will use default)
		return true
	default:
		return false
	}
}

// Architecture represents a CPU architecture.
type Architecture string

const (
	ArchX86_64 Architecture = "x86_64"
	ArchARM64  Architecture = "arm64"
	ArchARM    Architecture = "arm"
)

// Valid returns true if the architecture is valid.
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

// OS represents an operating system.
type OS string

const (
	OSLinux   OS = "linux"
	OSWindows OS = "windows"
	OSMacOS   OS = "macos"
)

// Valid returns true if the OS is valid.
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

// JobStatus represents the current status of a compilation job.
type JobStatus string

const (
	StatusQueued     JobStatus = "queued"
	StatusProcessing JobStatus = "processing"
	StatusCompleted  JobStatus = "completed"
	StatusFailed     JobStatus = "failed"
	StatusTimeout    JobStatus = "timeout"
)
