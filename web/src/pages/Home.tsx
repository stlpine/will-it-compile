import { useState, useEffect } from 'react'
import { CodeEditor } from '../components/CodeEditor'
import { EnvironmentSelector } from '../components/EnvironmentSelector'
import { CompilerOutput } from '../components/CompilerOutput'
import { JobStatus } from '../components/JobStatus'
import { useCompilation } from '../hooks/useCompilation'
import {
  Language,
  Standard,
  LANGUAGE_CONFIGS,
  DEFAULT_ARCHITECTURE,
  DEFAULT_OS,
} from '../types/api'

/**
 * Home page - Main compilation interface
 */
export function Home() {
  const [language, setLanguage] = useState<Language>('cpp')
  const [standard, setStandard] = useState<Standard>('c++20')
  const [code, setCode] = useState<string>('')

  const { isCompiling, result, error, statusMessage, compile, reset } =
    useCompilation()

  // Update default code when language changes
  useEffect(() => {
    const config = LANGUAGE_CONFIGS[language]
    if (config) {
      setCode(config.defaultCode)
      if (config.standard) {
        setStandard(config.standard)
      }
    }
  }, [language])

  const handleCompile = async () => {
    const config = LANGUAGE_CONFIGS[language]
    if (!config) {
      alert('Invalid language selected')
      return
    }

    // Encode source code to base64
    const encodedCode = btoa(code)

    await compile({
      code: encodedCode,
      language: config.language,
      compiler: config.compiler,
      standard: language === 'cpp' || language === 'c++' ? standard : undefined,
      architecture: DEFAULT_ARCHITECTURE,
      os: DEFAULT_OS,
    })
  }

  const handleReset = () => {
    reset()
    const config = LANGUAGE_CONFIGS[language]
    if (config) {
      setCode(config.defaultCode)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b border-gray-200 shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">
                Will It Compile?
              </h1>
              <p className="text-sm text-gray-600 mt-1">
                Secure, sandboxed code compilation service
              </p>
            </div>
            <div className="flex items-center gap-2">
              <a
                href="https://github.com/yourusername/will-it-compile"
                target="_blank"
                rel="noopener noreferrer"
                className="text-gray-600 hover:text-gray-900"
              >
                <svg
                  className="w-6 h-6"
                  fill="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    fillRule="evenodd"
                    d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"
                    clipRule="evenodd"
                  />
                </svg>
              </a>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="space-y-6">
          {/* Environment Selector */}
          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">
              Environment
            </h2>
            <EnvironmentSelector
              selectedLanguage={language}
              selectedStandard={standard}
              onLanguageChange={setLanguage}
              onStandardChange={setStandard}
              disabled={isCompiling}
            />
          </div>

          {/* Code Editor */}
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-gray-900">
                Source Code
              </h2>
              <div className="flex gap-2">
                <button
                  onClick={handleReset}
                  disabled={isCompiling}
                  className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Reset
                </button>
                <button
                  onClick={handleCompile}
                  disabled={isCompiling || !code.trim()}
                  className="px-6 py-2 text-sm font-medium text-white bg-primary-600 rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                >
                  {isCompiling ? (
                    <>
                      <svg
                        className="animate-spin h-4 w-4"
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                      >
                        <circle
                          className="opacity-25"
                          cx="12"
                          cy="12"
                          r="10"
                          stroke="currentColor"
                          strokeWidth="4"
                        ></circle>
                        <path
                          className="opacity-75"
                          fill="currentColor"
                          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                        ></path>
                      </svg>
                      Compiling...
                    </>
                  ) : (
                    <>
                      <svg
                        className="w-4 h-4"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
                        />
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                        />
                      </svg>
                      Compile
                    </>
                  )}
                </button>
              </div>
            </div>
            <CodeEditor
              value={code}
              onChange={setCode}
              language={language}
              readOnly={isCompiling}
              height="500px"
            />
          </div>

          {/* Job Status */}
          {(isCompiling || statusMessage) && (
            <JobStatus isCompiling={isCompiling} statusMessage={statusMessage} />
          )}

          {/* Error Display */}
          {error && !result && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-4">
              <h3 className="font-semibold text-red-900 mb-2">Error</h3>
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          {/* Compilation Result */}
          {result && (
            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">
                Result
              </h2>
              <CompilerOutput result={result} />
            </div>
          )}
        </div>
      </main>

      {/* Footer */}
      <footer className="bg-white border-t border-gray-200 mt-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <p className="text-center text-sm text-gray-600">
            Built with React, TypeScript, and Tailwind CSS • Powered by Docker
            containers
          </p>
        </div>
      </footer>
    </div>
  )
}
