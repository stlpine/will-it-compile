// TypeScript types matching Go backend models in pkg/models/

// Enums
export type Language = 'c' | 'cpp' | 'c++' | 'go' | 'rust'
export type Compiler = 'gcc' | 'go' | 'rustc'
export type CompilerVersion = string // e.g., "13", "1.23", "1.80"
export type Standard =
  // C++ standards
  | 'c++11'
  | 'c++14'
  | 'c++17'
  | 'c++20'
  | 'c++23'
  // C standards
  | 'c89'
  | 'c99'
  | 'c11'
  | 'c17'
  | 'c23'
  // Empty string for languages without standards
  | ''
export type Architecture = 'x86_64' | 'arm64' | 'arm' | ''
export type OS = 'linux' | 'windows' | 'macos' | ''
export type JobStatus =
  | 'queued'
  | 'processing'
  | 'completed'
  | 'failed'
  | 'timeout'

// CompilationRequest represents an incoming request to compile code
export interface CompilationRequest {
  code: string // Base64 encoded source code
  language: Language // e.g., "cpp", "go", "rust"
  standard?: Standard // e.g., "c++20", "c++17"
  architecture?: Architecture // e.g., "x86_64", "arm64"
  os?: OS // e.g., "linux"
  compiler?: string // e.g., "gcc-13", "go-1.23", "rustc-1.80"
}

// CompilationResult represents the result of a compilation
export interface CompilationResult {
  job_id: string
  success: boolean
  compiled: boolean // Whether it compiled successfully
  stdout: string
  stderr: string
  exit_code: number
  duration: number // Duration in nanoseconds (converted from time.Duration)
  error?: string
}

// CompilationJob represents a job to be processed
export interface CompilationJob {
  id: string
  request: CompilationRequest
  status: JobStatus
  created_at: string // ISO 8601 timestamp
  started_at?: string // ISO 8601 timestamp
  completed_at?: string // ISO 8601 timestamp
}

// JobResponse is returned when a job is created
export interface JobResponse {
  job_id: string
  status: JobStatus
}

// ErrorResponse represents an API error
export interface ErrorResponse {
  error: string
  message?: string
}

// EnvironmentSpec describes a compilation environment
export interface EnvironmentSpec {
  language: Language
  compiler: Compiler
  version: string
  standard?: Standard
  architecture: Architecture
  os: OS
  image_tag: string // Docker image tag
  flags?: string[]
}

// Environment represents a supported compilation environment
export interface Environment {
  language: string
  compilers: string[]
  standards?: string[]
  oses: string[]
  architectures: string[]
}

// Helper type for UI state
export interface UICompilationState {
  isCompiling: boolean
  result: CompilationResult | null
  error: string | null
}

// WorkerStats represents the current state of the worker pool
export interface WorkerStats {
  max_workers: number
  active_workers: number
  available_slots: number
  queued_jobs: number
  total_processed: number
  total_successful: number
  total_failed: number
  uptime: string
  uptime_seconds: number
  start_time: string // ISO 8601 timestamp
}

// Default values for optional fields
export const DEFAULT_STANDARD: Standard = 'c++20'
export const DEFAULT_ARCHITECTURE: Architecture = 'x86_64'
export const DEFAULT_OS: OS = 'linux'
export const DEFAULT_COMPILER: Compiler = 'gcc'
export const DEFAULT_GCC_VERSION: CompilerVersion = '13'
export const DEFAULT_GO_VERSION: CompilerVersion = '1.23'
export const DEFAULT_RUST_VERSION: CompilerVersion = '1.80'

// Available compiler versions (matching configs/environments.yaml)
export const GCC_VERSIONS: CompilerVersion[] = ['9', '10', '11', '12', '13']
export const GO_VERSIONS: CompilerVersion[] = ['1.20', '1.21', '1.22', '1.23']
export const RUST_VERSIONS: CompilerVersion[] = ['1.70', '1.75', '1.80']

// Language configurations for the UI
export interface LanguageConfig {
  language: Language
  label: string
  defaultCode: string
  compiler: Compiler
  defaultVersion: CompilerVersion
  availableVersions: CompilerVersion[]
  standard?: Standard
  fileExtension: string
}

export const LANGUAGE_CONFIGS: Record<string, LanguageConfig> = {
  cpp: {
    language: 'cpp',
    label: 'C++',
    defaultCode: `#include <iostream>

int main() {
    std::cout << "Hello, World!" << std::endl;
    return 0;
}`,
    compiler: 'gcc',
    defaultVersion: DEFAULT_GCC_VERSION,
    availableVersions: GCC_VERSIONS,
    standard: 'c++20',
    fileExtension: 'cpp',
  },
  c: {
    language: 'c',
    label: 'C',
    defaultCode: `#include <stdio.h>

int main() {
    printf("Hello, World!\\n");
    return 0;
}`,
    compiler: 'gcc',
    defaultVersion: DEFAULT_GCC_VERSION,
    availableVersions: GCC_VERSIONS,
    standard: 'c17',
    fileExtension: 'c',
  },
  go: {
    language: 'go',
    label: 'Go',
    defaultCode: `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}`,
    compiler: 'go',
    defaultVersion: DEFAULT_GO_VERSION,
    availableVersions: GO_VERSIONS,
    fileExtension: 'go',
  },
  rust: {
    language: 'rust',
    label: 'Rust',
    defaultCode: `fn main() {
    println!("Hello, World!");
}`,
    compiler: 'rustc',
    defaultVersion: DEFAULT_RUST_VERSION,
    availableVersions: RUST_VERSIONS,
    fileExtension: 'rs',
  },
}
