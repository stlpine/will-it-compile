import { useState, useCallback } from 'react'
import { compileAndWait } from '../services/api'
import type {
  CompilationRequest,
  CompilationResult,
  UICompilationState,
} from '../types/api'

/**
 * Custom hook for managing compilation state and logic
 */
export function useCompilation() {
  const [state, setState] = useState<UICompilationState>({
    isCompiling: false,
    result: null,
    error: null,
  })

  const [statusMessage, setStatusMessage] = useState<string>('')

  /**
   * Compile code with the provided request
   */
  const compile = useCallback(
    async (request: CompilationRequest): Promise<CompilationResult | null> => {
      setState({
        isCompiling: true,
        result: null,
        error: null,
      })
      setStatusMessage('Initializing compilation...')

      try {
        const result = await compileAndWait(request, (status) => {
          setStatusMessage(status)
        })

        setState({
          isCompiling: false,
          result,
          error: null,
        })

        setStatusMessage('')
        return result
      } catch (error) {
        const errorMessage =
          error instanceof Error ? error.message : 'Compilation failed'
        setState({
          isCompiling: false,
          result: null,
          error: errorMessage,
        })
        setStatusMessage('')
        return null
      }
    },
    []
  )

  /**
   * Reset the compilation state
   */
  const reset = useCallback(() => {
    setState({
      isCompiling: false,
      result: null,
      error: null,
    })
    setStatusMessage('')
  }, [])

  return {
    ...state,
    statusMessage,
    compile,
    reset,
  }
}
