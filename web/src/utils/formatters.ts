/**
 * Format duration from nanoseconds to human-readable string
 */
export function formatDuration(nanoseconds: number): string {
  const milliseconds = nanoseconds / 1_000_000
  const seconds = milliseconds / 1000

  if (seconds < 1) {
    return `${milliseconds.toFixed(0)}ms`
  } else if (seconds < 60) {
    return `${seconds.toFixed(2)}s`
  } else {
    const minutes = Math.floor(seconds / 60)
    const remainingSeconds = (seconds % 60).toFixed(0)
    return `${minutes}m ${remainingSeconds}s`
  }
}

/**
 * Format timestamp to human-readable string
 */
export function formatTimestamp(timestamp: string): string {
  const date = new Date(timestamp)
  return date.toLocaleString()
}

/**
 * Escape HTML in strings to prevent XSS
 */
export function escapeHtml(text: string): string {
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML
}

/**
 * Truncate text to maximum length with ellipsis
 */
export function truncate(text: string, maxLength: number): string {
  if (text.length <= maxLength) {
    return text
  }
  return text.substring(0, maxLength) + '...'
}

/**
 * Parse ANSI color codes and convert to HTML spans
 */
export function parseAnsiColors(text: string): string {
  // Basic ANSI color code parsing
  // This is a simplified version - can be enhanced with a library
  return text
    .replace(/\x1b\[31m/g, '<span style="color: #ef4444;">')
    .replace(/\x1b\[32m/g, '<span style="color: #10b981;">')
    .replace(/\x1b\[33m/g, '<span style="color: #f59e0b;">')
    .replace(/\x1b\[34m/g, '<span style="color: #3b82f6;">')
    .replace(/\x1b\[35m/g, '<span style="color: #a855f7;">')
    .replace(/\x1b\[36m/g, '<span style="color: #06b6d4;">')
    .replace(/\x1b\[0m/g, '</span>')
    .replace(/\x1b\[\d+m/g, '')
}
