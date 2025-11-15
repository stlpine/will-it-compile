import axios, { AxiosError } from 'axios'
import type {
  CompilationRequest,
  CompilationResult,
  JobResponse,
  Environment,
  ErrorResponse,
} from '../types/api'

// API client configuration
const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1'

// Create axios instance with default config
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 60000, // 60 seconds timeout
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor for logging
apiClient.interceptors.request.use(
  (config) => {
    console.log(`[API] ${config.method?.toUpperCase()} ${config.url}`)
    return config
  },
  (error) => {
    console.error('[API] Request error:', error)
    return Promise.reject(error)
  }
)

// Response interceptor for error handling
apiClient.interceptors.response.use(
  (response) => {
    console.log(`[API] Response ${response.status}:`, response.data)
    return response
  },
  (error: AxiosError<ErrorResponse>) => {
    if (error.response) {
      console.error(
        `[API] Error ${error.response.status}:`,
        error.response.data
      )
    } else if (error.request) {
      console.error('[API] No response received:', error.request)
    } else {
      console.error('[API] Error:', error.message)
    }
    return Promise.reject(error)
  }
)

/**
 * Submit a compilation job
 * POST /api/v1/compile
 */
export async function submitCompilation(
  request: CompilationRequest
): Promise<JobResponse> {
  try {
    const response = await apiClient.post<JobResponse>('/compile', request)
    return response.data
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.data) {
      throw new Error(error.response.data.message || error.response.data.error)
    }
    throw error
  }
}

/**
 * Get job result
 * GET /api/v1/compile/:job_id
 */
export async function getJobResult(jobId: string): Promise<CompilationResult> {
  try {
    const response = await apiClient.get<CompilationResult>(`/compile/${jobId}`)
    return response.data
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.data) {
      throw new Error(error.response.data.message || error.response.data.error)
    }
    throw error
  }
}

/**
 * Poll job result with exponential backoff
 * Continues polling until job is completed, failed, or timeout
 */
export async function pollJobResult(
  jobId: string,
  options: {
    maxAttempts?: number
    initialDelay?: number
    maxDelay?: number
    onProgress?: (attempt: number) => void
  } = {}
): Promise<CompilationResult> {
  const {
    maxAttempts = 60,
    initialDelay = 500,
    maxDelay = 5000,
    onProgress,
  } = options

  let attempt = 0
  let delay = initialDelay

  while (attempt < maxAttempts) {
    attempt++
    if (onProgress) {
      onProgress(attempt)
    }

    try {
      const result = await getJobResult(jobId)

      // Check if job is in terminal state
      if (
        result.success !== undefined ||
        ['completed', 'failed', 'timeout'].includes(
          (result as any).status || ''
        )
      ) {
        return result
      }

      // Job still processing, wait before next attempt
      await new Promise((resolve) => setTimeout(resolve, delay))

      // Exponential backoff with max delay
      delay = Math.min(delay * 1.5, maxDelay)
    } catch (error) {
      // On error, retry with backoff unless it's the last attempt
      if (attempt >= maxAttempts) {
        throw error
      }
      await new Promise((resolve) => setTimeout(resolve, delay))
      delay = Math.min(delay * 1.5, maxDelay)
    }
  }

  throw new Error('Job polling timeout: maximum attempts reached')
}

/**
 * Get list of supported environments
 * GET /api/v1/environments
 */
export async function getEnvironments(): Promise<Environment[]> {
  try {
    const response = await apiClient.get<Environment[]>('/environments')
    return response.data
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.data) {
      throw new Error(error.response.data.message || error.response.data.error)
    }
    throw error
  }
}

/**
 * Health check
 * GET /health
 */
export async function checkHealth(): Promise<{ status: string }> {
  try {
    const response = await apiClient.get<{ status: string }>('/health', {
      baseURL: '', // Use root path for health check
    })
    return response.data
  } catch (error) {
    if (axios.isAxiosError(error)) {
      throw new Error('API server is not available')
    }
    throw error
  }
}

/**
 * Compile code and wait for result
 * Convenience function that submits job and polls for result
 */
export async function compileAndWait(
  request: CompilationRequest,
  onProgress?: (status: string) => void
): Promise<CompilationResult> {
  // Submit compilation job
  if (onProgress) onProgress('Submitting compilation job...')
  const jobResponse = await submitCompilation(request)

  // Poll for result
  if (onProgress) onProgress('Compilation in progress...')
  const result = await pollJobResult(jobResponse.job_id, {
    onProgress: (attempt) => {
      if (onProgress && attempt % 5 === 0) {
        onProgress(`Still compiling (${attempt} checks)...`)
      }
    },
  })

  return result
}

// Export API client for advanced usage
export { apiClient }
