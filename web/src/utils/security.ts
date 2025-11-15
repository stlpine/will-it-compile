/**
 * Security utilities for input validation and sanitization
 */

// Maximum code size (1MB as per backend limit)
export const MAX_CODE_SIZE = 1 * 1024 * 1024

/**
 * Safely encode string to base64, handling Unicode characters
 * btoa() fails with Unicode, so we need to encode properly
 */
export function safeBase64Encode(str: string): string {
  try {
    // Convert to UTF-8 bytes first, then encode
    // Using TextEncoder for proper Unicode handling
    const encoder = new TextEncoder()
    const uint8Array = encoder.encode(str)

    // Convert Uint8Array to binary string
    let binaryString = ''
    for (let i = 0; i < uint8Array.length; i++) {
      binaryString += String.fromCharCode(uint8Array[i])
    }

    return btoa(binaryString)
  } catch (error) {
    throw new Error(
      'Failed to encode source code. Please check for invalid characters.'
    )
  }
}

/**
 * Safely decode base64 to string, handling Unicode characters
 */
export function safeBase64Decode(base64: string): string {
  try {
    const binaryString = atob(base64)
    const bytes = new Uint8Array(binaryString.length)

    for (let i = 0; i < binaryString.length; i++) {
      bytes[i] = binaryString.charCodeAt(i)
    }

    const decoder = new TextDecoder()
    return decoder.decode(bytes)
  } catch (error) {
    throw new Error('Failed to decode response')
  }
}

/**
 * Validate code input before submission
 */
export function validateCode(code: string): { valid: boolean; error?: string } {
  // Check if code is empty
  if (!code || code.trim().length === 0) {
    return { valid: false, error: 'Code cannot be empty' }
  }

  // Check code size
  const size = new Blob([code]).size
  if (size > MAX_CODE_SIZE) {
    return {
      valid: false,
      error: `Code size (${(size / 1024).toFixed(2)}KB) exceeds maximum allowed size (${MAX_CODE_SIZE / 1024}KB)`,
    }
  }

  // Check for null bytes (potential injection attempts)
  if (code.includes('\0')) {
    return { valid: false, error: 'Code contains invalid null bytes' }
  }

  return { valid: true }
}

/**
 * Sanitize output text for display
 * Removes potentially dangerous ANSI escape sequences while preserving safe ones
 */
export function sanitizeOutput(text: string): string {
  if (!text) return ''

  // Remove dangerous control characters except newlines, tabs, and carriage returns
  // Keep: \n (10), \t (9), \r (13)
  // Remove: other control chars (0-8, 11-12, 14-31, 127)
  let sanitized = text.replace(/[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]/g, '')

  // Remove OSC (Operating System Command) sequences - these can be dangerous
  sanitized = sanitized.replace(/\x1b\][^\x07]*\x07/g, '')
  sanitized = sanitized.replace(/\x1b\][^\x1b]*\x1b\\/g, '')

  // Limit ANSI escape sequences to safe color codes only
  // Allow basic color codes but remove others like cursor movement
  sanitized = sanitized.replace(/\x1b\[([0-9;]+)m/g, (_match, codes) => {
    // Only allow color codes (30-37, 40-47, 90-97, 100-107) and reset (0)
    const safeCodes = codes.split(';').filter((code: string) => {
      const num = parseInt(code, 10)
      return (
        num === 0 ||
        (num >= 30 && num <= 37) ||
        (num >= 40 && num <= 47) ||
        (num >= 90 && num <= 97) ||
        (num >= 100 && num <= 107)
      )
    })
    return safeCodes.length > 0 ? `\x1b[${safeCodes.join(';')}m` : ''
  })

  // Remove any remaining potentially dangerous escape sequences
  sanitized = sanitized.replace(/\x1b\[[^m]*[^m]/g, '')

  return sanitized
}

/**
 * Rate limiter for client-side request throttling
 */
export class RateLimiter {
  private requests: number[] = []
  private readonly maxRequests: number
  private readonly timeWindow: number

  constructor(maxRequests: number = 10, timeWindowMs: number = 60000) {
    this.maxRequests = maxRequests
    this.timeWindow = timeWindowMs
  }

  /**
   * Check if request is allowed
   */
  checkLimit(): { allowed: boolean; retryAfter?: number } {
    const now = Date.now()

    // Remove old requests outside the time window
    this.requests = this.requests.filter((time) => now - time < this.timeWindow)

    if (this.requests.length >= this.maxRequests) {
      const oldestRequest = Math.min(...this.requests)
      const retryAfter = Math.ceil(
        (oldestRequest + this.timeWindow - now) / 1000
      )
      return { allowed: false, retryAfter }
    }

    this.requests.push(now)
    return { allowed: true }
  }

  /**
   * Reset the rate limiter
   */
  reset(): void {
    this.requests = []
  }
}

/**
 * Sanitize error messages to prevent information leakage
 */
export function sanitizeErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    // Remove potential sensitive information from error messages
    let message = error.message

    // Remove file paths
    message = message.replace(/\/[\w\-./]+/g, '[path]')

    // Remove IP addresses
    message = message.replace(/\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b/g, '[ip]')

    // Remove potential secret keys or tokens
    message = message.replace(/[a-zA-Z0-9_-]{32,}/g, '[redacted]')

    return message
  }

  return 'An unexpected error occurred'
}

/**
 * Validate job ID format to prevent injection
 */
export function isValidJobId(jobId: string): boolean {
  // UUIDs are safe, alphanumeric with hyphens
  const uuidPattern =
    /^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$/i
  return uuidPattern.test(jobId)
}

/**
 * Check if running in secure context (HTTPS)
 */
export function isSecureContext(): boolean {
  return window.isSecureContext || window.location.protocol === 'https:'
}

/**
 * Warn if not in secure context
 */
export function checkSecureContext(): void {
  if (!isSecureContext() && window.location.hostname !== 'localhost') {
    console.warn(
      '⚠️ WARNING: Application is not running in a secure context (HTTPS). ' +
        'Data transmission may not be secure.'
    )
  }
}
