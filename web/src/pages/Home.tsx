import { useState, useEffect, useRef } from 'react'
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
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card'
import { Button } from '../components/ui/button'
import { Alert, AlertDescription, AlertTitle } from '../components/ui/alert'
import { Separator } from '../components/ui/separator'
import { Github, Play, RotateCcw, Code2, Zap, AlertTriangle, Shield } from 'lucide-react'
import { safeBase64Encode, validateCode, RateLimiter, checkSecureContext } from '../utils/security'

/**
 * Home page - Main compilation interface with shadcn/ui
 */
export function Home() {
  const [language, setLanguage] = useState<Language>('cpp')
  const [standard, setStandard] = useState<Standard>('c++20')
  const [code, setCode] = useState<string>('')
  const [validationError, setValidationError] = useState<string>('')
  const [securityWarning, setSecurityWarning] = useState<string>('')

  // Rate limiter: 10 requests per minute
  const rateLimiterRef = useRef(new RateLimiter(10, 60000))

  const { isCompiling, result, error, statusMessage, compile, reset } =
    useCompilation()

  // Check secure context on mount
  useEffect(() => {
    checkSecureContext()

    // Warn if not in secure context
    if (
      !window.isSecureContext &&
      window.location.hostname !== 'localhost' &&
      window.location.hostname !== '127.0.0.1'
    ) {
      setSecurityWarning(
        'Warning: Application is not running over HTTPS. Data transmission may not be secure.'
      )
    }
  }, [])

  // Update default code when language changes
  useEffect(() => {
    const config = LANGUAGE_CONFIGS[language]
    if (config) {
      setCode(config.defaultCode)
      if (config.standard) {
        setStandard(config.standard)
      }
    }
    // Clear validation error when language changes
    setValidationError('')
  }, [language])

  const handleCompile = async () => {
    setValidationError('')

    const config = LANGUAGE_CONFIGS[language]
    if (!config) {
      setValidationError('Invalid language selected')
      return
    }

    // Validate code input
    const validation = validateCode(code)
    if (!validation.valid) {
      setValidationError(validation.error || 'Invalid code')
      return
    }

    // Check rate limit
    const rateLimit = rateLimiterRef.current.checkLimit()
    if (!rateLimit.allowed) {
      setValidationError(
        `Rate limit exceeded. Please wait ${rateLimit.retryAfter} seconds before trying again.`
      )
      return
    }

    try {
      // Safely encode source code to base64 (handles Unicode)
      const encodedCode = safeBase64Encode(code)

      await compile({
        code: encodedCode,
        language: config.language,
        compiler: config.compiler,
        standard: language === 'cpp' || language === 'c++' ? standard : undefined,
        architecture: DEFAULT_ARCHITECTURE,
        os: DEFAULT_OS,
      })
    } catch (err) {
      setValidationError(err instanceof Error ? err.message : 'Failed to encode code')
    }
  }

  const handleReset = () => {
    reset()
    setValidationError('')
    const config = LANGUAGE_CONFIGS[language]
    if (config) {
      setCode(config.defaultCode)
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-b from-background to-muted/20">
      {/* Header */}
      <header className="border-b bg-card/50 backdrop-blur supports-[backdrop-filter]:bg-card/50 sticky top-0 z-50">
        <div className="container max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 rounded-lg bg-primary flex items-center justify-center">
                <Zap className="h-6 w-6 text-primary-foreground" />
              </div>
              <div>
                <h1 className="text-2xl font-bold tracking-tight">
                  Will It Compile?
                </h1>
                <p className="text-sm text-muted-foreground">
                  Secure, sandboxed code compilation service
                </p>
              </div>
            </div>
            <Button
              variant="ghost"
              size="icon"
              asChild
              className="rounded-full"
            >
              <a
                href="https://github.com/yourusername/will-it-compile"
                target="_blank"
                rel="noopener noreferrer"
              >
                <Github className="h-5 w-5" />
                <span className="sr-only">GitHub</span>
              </a>
            </Button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="space-y-6">
          {/* Security Warning */}
          {securityWarning && (
            <Alert variant="warning">
              <Shield className="h-5 w-5" />
              <AlertTitle>Security Notice</AlertTitle>
              <AlertDescription>{securityWarning}</AlertDescription>
            </Alert>
          )}

          {/* Environment Selector Card */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Code2 className="h-5 w-5" />
                Environment
              </CardTitle>
              <CardDescription>
                Select your programming language and compilation settings
              </CardDescription>
            </CardHeader>
            <Separator />
            <CardContent className="pt-6">
              <EnvironmentSelector
                selectedLanguage={language}
                selectedStandard={standard}
                onLanguageChange={setLanguage}
                onStandardChange={setStandard}
                disabled={isCompiling}
              />
            </CardContent>
          </Card>

          {/* Code Editor Card */}
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle className="flex items-center gap-2">
                    Source Code
                  </CardTitle>
                  <CardDescription>
                    Write your code in the editor below
                  </CardDescription>
                </div>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="default"
                    onClick={handleReset}
                    disabled={isCompiling}
                  >
                    <RotateCcw className="h-4 w-4 mr-2" />
                    Reset
                  </Button>
                  <Button
                    size="default"
                    onClick={handleCompile}
                    disabled={isCompiling || !code.trim()}
                  >
                    {isCompiling ? (
                      <>
                        <RotateCcw className="h-4 w-4 mr-2 animate-spin" />
                        Compiling...
                      </>
                    ) : (
                      <>
                        <Play className="h-4 w-4 mr-2" />
                        Compile
                      </>
                    )}
                  </Button>
                </div>
              </div>
            </CardHeader>
            <Separator />
            <CardContent className="pt-6">
              <CodeEditor
                value={code}
                onChange={setCode}
                language={language}
                readOnly={isCompiling}
                height="500px"
              />
            </CardContent>
          </Card>

          {/* Job Status */}
          {(isCompiling || statusMessage) && (
            <JobStatus isCompiling={isCompiling} statusMessage={statusMessage} />
          )}

          {/* Validation Error */}
          {validationError && (
            <Alert variant="warning">
              <AlertTriangle className="h-5 w-5" />
              <AlertTitle>Validation Error</AlertTitle>
              <AlertDescription>{validationError}</AlertDescription>
            </Alert>
          )}

          {/* Error Display */}
          {error && !result && (
            <Alert variant="destructive">
              <AlertTitle>Error</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {/* Compilation Result */}
          {result && (
            <Card>
              <CardHeader>
                <CardTitle>Compilation Result</CardTitle>
                <CardDescription>
                  View the output of your compilation below
                </CardDescription>
              </CardHeader>
              <Separator />
              <CardContent className="pt-6">
                <CompilerOutput result={result} />
              </CardContent>
            </Card>
          )}
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t bg-card/50 mt-12">
        <div className="container max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <p className="text-center text-sm text-muted-foreground">
            Built with{' '}
            <span className="font-semibold">React</span>,{' '}
            <span className="font-semibold">TypeScript</span>,{' '}
            <span className="font-semibold">shadcn/ui</span>, and{' '}
            <span className="font-semibold">Tailwind CSS</span> •
            Powered by Docker containers 🐳
          </p>
        </div>
      </footer>
    </div>
  )
}
