interface JobStatusProps {
  isCompiling: boolean
  statusMessage: string
}

/**
 * Component to display job status during compilation
 */
export function JobStatus({ isCompiling, statusMessage }: JobStatusProps) {
  if (!isCompiling && !statusMessage) {
    return null
  }

  return (
    <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
      <div className="flex items-center gap-3">
        {/* Spinner */}
        {isCompiling && (
          <svg
            className="animate-spin h-5 w-5 text-blue-600"
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
        )}

        {/* Status Message */}
        <div>
          <p className="text-sm font-medium text-blue-900">
            {isCompiling ? 'Compiling...' : 'Status'}
          </p>
          {statusMessage && (
            <p className="text-sm text-blue-700">{statusMessage}</p>
          )}
        </div>
      </div>
    </div>
  )
}
