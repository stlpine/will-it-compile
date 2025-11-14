import { CompilationResult } from '../../types/api'
import { formatDuration } from '../../utils/formatters'

interface CompilerOutputProps {
  result: CompilationResult
}

/**
 * Component to display compilation results
 */
export function CompilerOutput({ result }: CompilerOutputProps) {
  const hasStdout = result.stdout && result.stdout.trim().length > 0
  const hasStderr = result.stderr && result.stderr.trim().length > 0

  return (
    <div className="space-y-4">
      {/* Status Banner */}
      <div
        className={`p-4 rounded-lg ${
          result.compiled
            ? 'bg-green-50 border border-green-200'
            : 'bg-red-50 border border-red-200'
        }`}
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            {result.compiled ? (
              <svg
                className="w-6 h-6 text-green-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            ) : (
              <svg
                className="w-6 h-6 text-red-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            )}
            <div>
              <h3
                className={`font-semibold ${
                  result.compiled ? 'text-green-900' : 'text-red-900'
                }`}
              >
                {result.compiled
                  ? 'Compilation Successful'
                  : 'Compilation Failed'}
              </h3>
              <p
                className={`text-sm ${
                  result.compiled ? 'text-green-700' : 'text-red-700'
                }`}
              >
                Exit code: {result.exit_code} • Duration:{' '}
                {formatDuration(result.duration)}
              </p>
            </div>
          </div>
          <div className="text-sm text-gray-500">
            Job ID: <code className="text-xs">{result.job_id}</code>
          </div>
        </div>
      </div>

      {/* Error Message */}
      {result.error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <h4 className="font-semibold text-red-900 mb-2">Error</h4>
          <pre className="text-sm text-red-800 whitespace-pre-wrap font-mono">
            {result.error}
          </pre>
        </div>
      )}

      {/* Standard Output */}
      {hasStdout && (
        <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
          <h4 className="font-semibold text-gray-900 mb-2 flex items-center gap-2">
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
                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
              />
            </svg>
            Standard Output
          </h4>
          <pre className="text-sm text-gray-800 whitespace-pre-wrap font-mono bg-white p-3 rounded border border-gray-200 max-h-64 overflow-y-auto">
            {result.stdout}
          </pre>
        </div>
      )}

      {/* Standard Error */}
      {hasStderr && (
        <div
          className={`border rounded-lg p-4 ${
            result.compiled
              ? 'bg-yellow-50 border-yellow-200'
              : 'bg-red-50 border-red-200'
          }`}
        >
          <h4
            className={`font-semibold mb-2 flex items-center gap-2 ${
              result.compiled ? 'text-yellow-900' : 'text-red-900'
            }`}
          >
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
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
              />
            </svg>
            {result.compiled ? 'Warnings' : 'Compilation Errors'}
          </h4>
          <pre
            className={`text-sm whitespace-pre-wrap font-mono p-3 rounded border max-h-64 overflow-y-auto ${
              result.compiled
                ? 'text-yellow-900 bg-yellow-50 border-yellow-300'
                : 'text-red-900 bg-white border-red-200'
            }`}
          >
            {result.stderr}
          </pre>
        </div>
      )}

      {/* No Output Message */}
      {!hasStdout && !hasStderr && !result.error && (
        <div className="bg-gray-50 border border-gray-200 rounded-lg p-4 text-center text-gray-500">
          No output generated
        </div>
      )}
    </div>
  )
}
